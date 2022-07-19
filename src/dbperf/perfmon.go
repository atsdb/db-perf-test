/*
* @Author: ronan
* @Date:   2018-03-04 10:41:36
* @Last Modified by:   ronanj
* @Last Modified time: 2022-07-19 08:17:14
 */
package dbperf

import (
	"dbperf/src/dbdriver"
	"fmt"
	"github.com/mgutz/ansi"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type testState int

const (
	warming testState = iota
	running
	ending
	stopped
)

type PerfLogEntry struct {
	EntryCount   int
	Duration     time.Duration
	DeltaOps     int64
	TotalOps     int64
	OpsPerSecond float64
	CpuLoad      float64
	CpuUsage     float64
	MemUsage     uint64
}

type PerformanceMonitor struct {
	table           dbdriver.Table
	testWG          *sync.WaitGroup
	reportWG        *sync.WaitGroup
	mutex           *sync.Mutex
	state           testState
	startTime       time.Time
	startRunTime    time.Time
	runningDuration time.Duration

	opscount int64
	testname string
	duration time.Duration
	logs     []PerfLogEntry
	master   *PerformanceMonitor

	stats PerfLogEntry
}

func NewPerfMonitor(table dbdriver.Table) *PerformanceMonitor {

	perf := &PerformanceMonitor{
		testWG:   &sync.WaitGroup{},
		reportWG: &sync.WaitGroup{},
		mutex:    &sync.Mutex{},
		table:    table,
		logs:     make([]PerfLogEntry, 0),
	}

	return perf

}

func (p *PerformanceMonitor) NRows() int64 {

	rowcount := int64(0)
	query := "select max(idx) from " + p.table.Name()
	if row := p.table.DB().QueryRow(query); row != nil {
		row.Scan(&rowcount)
	}
	return rowcount

}

func (p *PerformanceMonitor) Inc(e error) {

	if e != nil {
		log.Fatal("[write] query error: ", e)
	} else {
		if p.master != nil {
			atomic.AddInt64(&p.master.opscount, 1)

		} else {
			atomic.AddInt64(&p.opscount, 1)
		}
	}
}

func (p *PerformanceMonitor) LinkMaster(master *PerformanceMonitor) {
	p.master = master
}

func (p *PerformanceMonitor) Start(what string, duration time.Duration) {

	p.startRunTime = time.Now()
	p.startTime = time.Now()
	p.state = warming
	p.opscount = 0
	p.testname = what
	p.duration = duration
	p.logs = make([]PerfLogEntry, 0)
	p.state = warming
	p.stats = PerfLogEntry{}

	if p.master == nil {
		nrows := p.NRows()
		fmt.Printf("=============== %s [%v - %d rows already in the table] ===============\n",
			ansi.Color(what, "green"), duration, nrows)
		go p.periodicReport()

		/* Monitor the CPU load for 5 seconds before starting */
		time.Sleep(time.Second * 5)
	}

	p.state = running
	p.startRunTime = time.Now()
}

func (p *PerformanceMonitor) periodicReport() {
	p.reportWG.Add(1)
	delta := time.Second
	for p.state != stopped {

		pcount := p.opscount
		start := time.Now()
		time.Sleep(delta)
		if p.state != stopped {

			nops := p.opscount - pcount
			cdt := float64(time.Since(start))
			tdt := float64(time.Since(p.startRunTime))

			if p.state != running {
				tdt = float64(p.runningDuration)
			}

			if p.state == warming {
				tdt = 1
			}

			crps := float64(nops) * float64(time.Millisecond) / cdt
			trps := float64(p.opscount) * float64(time.Millisecond) / tdt

			var cpuUsage float64
			if usage, err := cpu.Percent(0, false); err == nil {
				cpuUsage = usage[0]
			}

			var cpuLoad float64
			if load, err := load.Avg(); err == nil && load != nil {
				cpuLoad = load.Load1
			}

			var memUsage uint64
			if mem, err := mem.VirtualMemory(); err == nil && mem != nil {
				memUsage = mem.Used
			}

			ptime := 100 * time.Since(p.startRunTime) / p.duration
			sptime := ansi.Color(fmt.Sprintf("%2d%%", ptime), "red")
			if p.state != running {
				sptime = "---"
			}

			fmt.Printf(
				"[%s] ops: %5.2f/%5.2f - %9d - Load:%4.1f - Cpu:%5.1f - Mem:%.1fGB - %3dsec[%s]\n",
				ansi.Color(p.testname, "blue"),
				crps, trps, /* Kops/sec*/
				p.opscount,
				cpuLoad,
				cpuUsage,
				float64(memUsage)/float64(1024*1024*1024),
				time.Since(p.startTime)/time.Second,
				sptime,
			)

			p.logs = append(p.logs, PerfLogEntry{
				Duration:     time.Since(p.startTime),
				TotalOps:     p.opscount,
				DeltaOps:     nops,
				OpsPerSecond: crps,
				CpuLoad:      cpuLoad,
				CpuUsage:     cpuUsage,
				MemUsage:     memUsage,
			})

			if p.state == running {
				p.stats.CpuLoad += cpuLoad
				p.stats.CpuUsage += cpuUsage
				p.stats.MemUsage += memUsage
				p.stats.EntryCount += 1
			}

		}

		delta = time.Second*2 - time.Since(start)

	}
	p.reportWG.Done()
}

func (p *PerformanceMonitor) Finish() {

	if p.state != stopped {
		time.Sleep(time.Second)
		p.testWG.Wait()
		p.runningDuration = time.Since(p.startRunTime)
		p.state = ending

		/* Keep the CPU load counter active for 10 seconds */
		time.Sleep(time.Second * 5)
		p.state = stopped

		/* Wait for the reporting to finish */
		p.reportWG.Wait()
	}

}

func (p *PerformanceMonitor) Summary() {

	p.Finish()

	rps := int(float64(p.opscount) * float64(time.Second) / float64(p.runningDuration))

	filename := fmt.Sprintf("%s-%s", strings.Replace(p.testname, "/", "-", -1), time.Now().Format("20060102-150405"))

	header := "#Summary: Generated on: " + time.Now().Format(time.RFC1123Z) + "\n"
	header += "#Summary: Test Case: " + p.testname + "\n"
	header += "#Summary: Ops/Sec: " + fmt.Sprintf("%d", rps) + "\n"
	header += "#Summary: Operations: " + fmt.Sprintf("%d", p.opscount) + " rows\n"
	header += "#Summary: Duration: " + fmt.Sprintf("%d", p.runningDuration/time.Millisecond) + " msec\n"
	header += "#Summary: Avg Cpu Load: " + fmt.Sprintf("%.2f", float64(p.stats.CpuLoad)/float64(p.stats.EntryCount)) + "\n"
	header += "#Summary: Avg Cpu Usage: " + fmt.Sprintf("%.2f", float64(p.stats.CpuUsage)/float64(p.stats.EntryCount)) + " %%\n"
	header += "#Summary: Avg Mem Usage: " + fmt.Sprintf("%d", p.stats.MemUsage/uint64(p.stats.EntryCount)) + " B\n"
	header += "#Summary: File: " + filename + "\n"

	fmt.Printf(strings.Replace(header, "#Summary:", ansi.Color("Summary:", "green"), -1))

	logFilename := fmt.Sprintf("results/dbperf-results-%s.csv", time.Now().Format("20060102"))
	if f, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
		hostname, _ := os.Hostname()
		result := fmt.Sprintf("%s, %s, %s, %d, %d, %d, %.2f, %.2f, %d, %s\n",
			hostname,
			time.Now().Format("2006-01-02 15:04:05"),
			p.testname,
			p.opscount,
			p.runningDuration/time.Millisecond,
			rps,
			float64(p.stats.CpuLoad)/float64(p.stats.EntryCount),
			float64(p.stats.CpuUsage)/float64(p.stats.EntryCount),
			p.stats.MemUsage/uint64(p.stats.EntryCount),
			filename,
		)
		if _, err := f.WriteString(result); err != nil {
			fmt.Printf("%s: %v\n", ansi.Color("Oops, can not write result log", "red"), err)
		}
		f.Close()
		fmt.Println(ansi.Color(result, "blue"))
	} else {
		fmt.Printf("%s: %v\n", ansi.Color("Oops, can not write result log", "red"), err)
	}

	if p.master == nil {

		// Duration     time.Duration
		// DeltaOps     int64
		// TotalOps     int64
		// OpsPerSecond float64
		// CpuLoad      float64
		// CpuUsage     float64
		// MemUsage     uint64

		csv := header
		csv += "Time In Millisecond, Delta Ops, Total Ops, Ops Per Second, Cpu Load, Cpu Usage, Mem Usage\n"
		for _, log := range p.logs {
			csv += fmt.Sprintf("%d, %d, %d, %f, %f, %f, %d\n",
				log.Duration/time.Millisecond, log.DeltaOps, log.TotalOps, log.OpsPerSecond,
				log.CpuLoad, log.CpuUsage, log.MemUsage,
			)
		}
		ioutil.WriteFile("reports/"+filename+".csv", []byte(csv), 0644)
	}

}
