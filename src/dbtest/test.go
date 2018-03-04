/*
* @Author: ronan
* @Date:   2018-03-04 10:42:09
* @Last Modified by:   ron
* @Last Modified time: 2018-03-04 18:00:43
 */
package dbtest

import (
	"dbdriver"
	"dbperf"
	"fmt"
	"log"
	"time"
)

func DoPerfTest(dbcon string, mode string, engine string, table string, durationInSecond int, concurrent int) {

	duration := time.Second * time.Duration(durationInSecond)
	switch mode {

	case "write":
		// MyISAM does not support transactions
		if engine != "MyISAM" {
			// ------ With transaction, single connection
			multiTheadTxWrite(dbcon, engine, table, concurrent, duration).Summary()
		}

		// ------ No transaction, single connection
		multiTheadWrite(dbcon, engine, table, concurrent, duration).Summary()

		// ------ No transaction, mmultiple connections
		multiConnWrite(dbcon, engine, table, concurrent, duration)

	case "write-multi-conn":
		multiConnWrite(dbcon, engine, table, concurrent, duration)

	case "read":

		for nThreads := 1; nThreads < concurrent; nThreads++ {
			multiTheadRead(dbcon, engine, table, nThreads).Summary()
		}

	default:
		log.Printf("Unknown test mode '%s'\n", mode)

	}

}

func multiTheadWrite(dbcon string, engine string, testType string, nThreads int, duration time.Duration) *dbperf.PerformanceMonitor {

	table, generator := testTable(dbcon, engine, testType, "notx-")
	perf := dbperf.NewPerfMonitor(table)
	perf.Start(fmt.Sprintf("write-no-tx/%s/%d-threads/%s", engine, nThreads, testType), duration)
	go perf.WriteNoTx(generator, nThreads)
	perf.Finish()
	table.Close()
	return perf

}

func multiTheadTxWrite(dbcon string, engine string, testType string, nThreads int, duration time.Duration) *dbperf.PerformanceMonitor {

	table, generator := testTable(dbcon, engine, testType, "tx-")
	perf := dbperf.NewPerfMonitor(table)
	perf.Start(fmt.Sprintf("write-tx/%s/%d-threads/%s", engine, nThreads, testType), duration)
	go perf.WriteTx(generator, nThreads)
	perf.Finish()
	table.Close()
	return perf

}

func multiConnWrite(dbcon string, engine string, testType string, nConnections int, duration time.Duration) *dbperf.PerformanceMonitor {

	mons := make([]*dbperf.PerformanceMonitor, nConnections)
	tables := make([]dbdriver.Table, nConnections)
	for i := 0; i < nConnections; i++ {

		table, generator := testTable(dbcon, engine, testType, fmt.Sprintf("conn%d-", i))
		perf := dbperf.NewPerfMonitor(table)
		tables[i] = table
		mons[i] = perf
		if i > 0 {
			perf.LinkMaster(mons[0])
		}

		nThreads := 1
		perf.Start(fmt.Sprintf("write-no-tx/%s/%d-connections/%s", engine, nConnections, testType), duration)
		go perf.WriteTx(generator, nThreads)

	}

	for i := 0; i < nConnections; i++ {
		mons[i].Finish()
		tables[i].Close()
	}

	return mons[0]

}

func multiTheadRead(dbcon string, engine string, testType string, nThreads int) *dbperf.PerformanceMonitor {

	table, _ := testTable(dbcon, engine, testType, "notx-")
	perf := dbperf.NewPerfMonitor(table)
	for perf.NRows() < 10*1000*1000 {
		log.Printf("[read] There are not enough rows (%d) in the DB - generating few now...\n", perf.NRows())
		multiConnWrite(dbcon, engine, testType, 10, time.Minute)
	}

	perf.Start(fmt.Sprintf("read/%d-threads", nThreads), time.Minute)
	nrows := perf.NRows()
	log.Printf("[read] There are %d rows\n", nrows)
	start := int64(0)
	for n := 0; n < nThreads; n++ {
		end := (nrows + 1) * int64(n+1) / int64(nThreads)
		go perf.Read(start, end)
		start = end
	}
	perf.Finish()
	return perf
}
