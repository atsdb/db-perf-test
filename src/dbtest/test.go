/*
* @Author: ronan
* @Date:   2018-03-04 10:42:09
* @Last Modified by:   ron
* @Last Modified time: 2018-03-06 16:22:11
 */
package dbtest

import (
	"dbdriver"
	"dbperf"
	"fmt"
	"log"
	"time"
)

func DoPerfTest(dbcon string, mode string, engine string, table string, durationInSecond int, nthreads int, ntables int) {

	duration := time.Second * time.Duration(durationInSecond)
	switch mode {

	case "write":

		// ------ No transaction, single connection, mmultiple threads
		multiThreadWrite(dbcon, engine, table, nthreads, duration).Summary()

		// MyISAM does not support transactions
		if engine != "MyISAM" {
			// ------ With transaction, single table, mmultiple threads
			multiThreadTxWrite(dbcon, engine, table, nthreads, duration).Summary()
		}

		// ------ With transaction, single connection, mmultiple table
		multiTableTxWrite(dbcon, engine, table, ntables, duration)

	case "read":

		for n := 1; n < nthreads; n++ {
			multiTheadRead(dbcon, engine, table, n).Summary()
		}

	default:
		log.Printf("Unknown test mode '%s'\n", mode)

	}

}

func multiThreadWrite(dbcon string, engine string, testType string, nThreads int, duration time.Duration) *dbperf.PerformanceMonitor {

	table, generator := testTable(dbcon, engine, testType, "notx-")
	perf := dbperf.NewPerfMonitor(table)
	perf.Start(fmt.Sprintf("write-no-tx/%s/%d-threads/%s", engine, nThreads, testType), duration)
	go perf.WriteNoTx(generator, nThreads)
	perf.Finish()
	table.Close()
	return perf

}

func multiThreadTxWrite(dbcon string, engine string, testType string, nThreads int, duration time.Duration) *dbperf.PerformanceMonitor {

	table, generator := testTable(dbcon, engine, testType, "tx-")
	perf := dbperf.NewPerfMonitor(table)
	perf.Start(fmt.Sprintf("write-tx/%s/%d-threads/%s", engine, nThreads, testType), duration)
	go perf.WriteTx(generator, nThreads)
	perf.Finish()
	table.Close()
	return perf

}

func multiTableTxWrite(dbcon string, engine string, testType string, nTables int, duration time.Duration) *dbperf.PerformanceMonitor {

	mons := make([]*dbperf.PerformanceMonitor, nTables)
	tables := make([]dbdriver.Table, nTables)
	for i := 0; i < nTables; i++ {

		table, generator := testTable(dbcon, engine, testType, fmt.Sprintf("conn%d-", i))
		perf := dbperf.NewPerfMonitor(table)
		tables[i] = table
		mons[i] = perf
		if i > 0 {
			perf.LinkMaster(mons[0])
		}

		nThreads := 1
		perf.Start(fmt.Sprintf("write-tx/%s/%d-tables/%s", engine, nTables, testType), duration)
		go perf.WriteTx(generator, nThreads)

	}

	for i := 0; i < nTables; i++ {
		mons[i].Finish()
		tables[i].Close()
	}

	return mons[0]

}

func multiTheadRead(dbcon string, engine string, testType string, nThreads int) *dbperf.PerformanceMonitor {

	table, generator := testTable(dbcon, engine, testType, "read-")
	perf := dbperf.NewPerfMonitor(table)
	for perf.NRows() < 100*1000*1000 {
		log.Printf("[read] There are not enough rows (%d) in the DB - generating few now...\n", perf.NRows())

		perf := dbperf.NewPerfMonitor(table)
		perf.Start(fmt.Sprintf("write-tx/%s/%d-threads/%s", engine, 10, testType), time.Minute*10)
		go perf.WriteTx(generator, 10)
		perf.Finish()

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
