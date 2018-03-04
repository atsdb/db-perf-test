/*
* @Author: ronan
* @Date:   2018-03-04 10:42:09
* @Last Modified by:   ronan
* @Last Modified time: 2018-03-04 12:37:37
 */
package dbtest

import (
	"dbperf"
	"fmt"
	"log"
	"time"
)

func DoPerfTest(dbcon string, mode string, engine string, durationInSecond int) {

	testType := "large-table"
	duration := time.Second * time.Duration(durationInSecond)
	if mode == "write" {

		nThreads := 10

		// ------ No transaction, single connection

		table, generator := testLargeTable(dbcon, engine, "notx-")
		perf := dbperf.NewPerfMonitor(table)
		perf.Start(fmt.Sprintf("write-no-tx/%s/%d-threads/%s", engine, nThreads, testType), duration)
		go perf.WriteNoTx(generator, nThreads)
		perf.Summary()

		// ------ With transaction, single connection

		table, generator = testLargeTable(dbcon, engine, "tx-")
		perf = dbperf.NewPerfMonitor(table)
		perf.Start(fmt.Sprintf("write-tx/%s/%d-threads/%s", engine, nThreads, testType), duration)
		go perf.WriteTx(generator, nThreads)
		perf.Summary()

		// ------ No transaction, mmultiple connections

		nConnections := 10
		mons := make([]*dbperf.PerformanceMonitor, nConnections)
		for i := 0; i < nConnections; i++ {
			table, generator := testLargeTable(dbcon, engine, fmt.Sprintf("conn%d-", i))
			perf := dbperf.NewPerfMonitor(table)
			mons[i] = perf
			if i > 0 {
				perf.LinkMaster(mons[0])
			}

			nThreads := 1
			perf.Start(fmt.Sprintf("write-no-tx/%s/%d-connections/%s", engine, nConnections, testType), duration)
			go perf.WriteNoTx(generator, nThreads)

		}

		for i := 0; i < nConnections; i++ {
			mons[i].Summary()
		}
	}

	if mode == "read" {

		table, generator := testLargeTable(dbcon, engine, "notx-")
		perf := dbperf.NewPerfMonitor(table)
		for perf.NRows() < 50*1000*1000 {
			nThreads := 10
			log.Printf("[read] There are not enough rows (%d) in the DB - generating few now...\n", perf.NRows())
			perf.Start(fmt.Sprintf("write-no-tx/%s/%d-threads/%s", engine, nThreads, testType), time.Minute)
			go perf.WriteTx(generator, nThreads)
			perf.Finish()
		}

		perf.Start("read", duration)
		nrows := perf.NRows()
		go perf.Read(0, nrows+1)
		perf.Summary()

	}

}
