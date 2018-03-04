/*
* @Author: ronan
* @Date:   2018-03-04 10:41:36
* @Last Modified by:   ron
* @Last Modified time: 2018-03-04 14:10:57
 */
package dbperf

import (
	"dbdriver"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type PerfLogEntry struct {
	Duration     time.Duration
	DeltaOps     int64
	TotalOps     int64
	OpsPerSecond float64
	CpuLoad      float64
	CpuUsage     float64
	MemUsage     uint64
}

type PerformanceMonitor struct {
	table     dbdriver.Table
	testWG    *sync.WaitGroup
	reportWG  *sync.WaitGroup
	mutex     *sync.Mutex
	startTime time.Time
	running   bool
	opscount  int64
	testname  string
	duration  time.Duration
	logs      []PerfLogEntry
	master    *PerformanceMonitor
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

	rowcount := int64(-1)
	query := "select max(idx) from " + p.table.Name()
	if row := p.table.DB().QueryRow(query); row != nil {
		row.Scan(&rowcount)
	}
	return rowcount

}

func (p *PerformanceMonitor) Inc(e error) {

	if e != nil {
		log.Fatal("[write] query error", e)
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

	p.startTime = time.Now()
	p.running = true
	p.opscount = 0
	p.testname = what
	p.duration = duration
	p.logs = make([]PerfLogEntry, 0)

	if p.master == nil {
		fmt.Printf("[=============== %s/%v ===============]\n", what, duration)
		go p.periodicReport()
	}
}

func (p *PerformanceMonitor) periodicReport() {
	p.reportWG.Add(1)
	delta := time.Second
	for p.running {

		pcount := p.opscount
		start := time.Now()
		time.Sleep(delta)
		if p.running {

			nops := p.opscount - pcount
			crps := float64(nops) * float64(time.Millisecond) / float64(time.Since(start))
			trps := float64(pcount) * float64(time.Millisecond) / float64(time.Since(p.startTime))

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

			ptime := 100 * time.Since(p.startTime) / p.duration
			fmt.Printf("[perf/%s/%02d%%] %5.2f Kops/sec [%5.2f] %-9d Load:%.1f- Usage:%.1f - Mem:%.1fGB [%v] \n",
				p.testname, ptime, crps, trps,
				p.opscount,
				cpuLoad, cpuUsage,
				float64(memUsage)/float64(1024*1024*1024),
				time.Since(p.startTime),
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

		}

		delta = time.Second*2 - time.Since(start)

	}
	p.reportWG.Done()
}

func (p *PerformanceMonitor) Finish() {

	time.Sleep(time.Second)
	p.testWG.Wait()
	p.running = false

	/* Wait for the reporting to finish */
	p.reportWG.Wait()

}

func (p *PerformanceMonitor) Summary() {

	p.Finish()

	dt := time.Now().Sub(p.startTime)
	rps := float64(p.opscount) * float64(time.Millisecond) / float64(dt)
	fmt.Printf("****** Summary ****** %.2fK ops/sec [%d rows in %v]\n", rps, p.opscount, dt)

	// Duration     time.Duration
	// DeltaOps     int64
	// TotalOps     int64
	// OpsPerSecond float64
	// CpuLoad      float64
	// CpuUsage     float64
	// MemUsage     uint64

	if p.master == nil {
		csv := "#Test generated on " + time.Now().Format(time.RFC1123Z) + "\n"
		csv += "#Test Case: " + p.testname + "\n"
		csv += "#Summary: " + fmt.Sprintf("%.2fK ops/sec [%d rows in %v]\n", rps, p.opscount, dt)
		csv += "Time In Millisecond, Delta Ops, Total Ops, Ops Per Second, Cpu Load, Cpu Usage, Mem Usage\n"
		for _, log := range p.logs {
			csv += fmt.Sprintf("%d, %d, %d, %f, %f, %f, %d\n",
				log.Duration/time.Millisecond, log.DeltaOps, log.TotalOps, log.OpsPerSecond,
				log.CpuLoad, log.CpuUsage, log.MemUsage,
			)
		}
		filename := fmt.Sprintf("%s-%s", strings.Replace(p.testname, "/", "-", -1), time.Now().Format("20060102-150405"))
		ioutil.WriteFile("reports/"+filename+".csv", []byte(csv), 0644)
		log.Println("File " + filename + " written")

	}
	time.Sleep(time.Second * 5)

}
