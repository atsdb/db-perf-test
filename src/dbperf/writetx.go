/*
* @Author: ronan
* @Date:   2018-03-04 10:02:54
* @Last Modified by:   ronan
* @Last Modified time: 2018-03-04 13:00:57
 */
package dbperf

import (
	"sync"
	"sync/atomic"
	"time"
)

func (p *PerformanceMonitor) WriteTx(generator func() []interface{}, nThreads int) {

	p.testWG.Add(1)
	total := int64(0)
	isRunning := true

	tx, insert := p.table.PrepareTxInsert()

	wg := &sync.WaitGroup{}

	transactionMutex := &sync.RWMutex{}
	insertThread := func() {

		for isRunning {
			values := generator()

			transactionMutex.RLock()
			err := insert(values...)
			transactionMutex.RUnlock()

			atomic.AddInt64(&total, 1)
			p.Inc(err)
		}
	}

	commitThread := func() {

		wg.Add(1)
		for isRunning {
			time.Sleep(time.Second * 1)
			transactionMutex.Lock()
			tx.Commit()
			tx, insert = p.table.PrepareTxInsert()
			transactionMutex.Unlock()
		}
		wg.Done()

	}

	go commitThread()
	for i := 0; i < nThreads; i++ {
		go insertThread()
	}
	time.Sleep(p.duration)

	isRunning = false
	wg.Wait()
	tx.Commit()

	p.testWG.Done()
}
