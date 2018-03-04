/*
* @Author: ronan
* @Date:   2018-03-04 09:50:43
* @Last Modified by:   ronan
* @Last Modified time: 2018-03-04 13:00:50
 */
package dbperf

import (
	"sync"
	"sync/atomic"
	"time"
)

func (p *PerformanceMonitor) WriteNoTx(generator func() []interface{}, nThreads int) {

	p.testWG.Add(1)
	total := int64(0)
	isRunning := true

	insert := p.table.PrepareInsert()

	wg := &sync.WaitGroup{}

	insertThread := func(i int) {

		wg.Add(1)
		for isRunning {

			values := generator()
			err := insert(values...)
			atomic.AddInt64(&total, 1)
			p.Inc(err)

		}
		wg.Done()

	}

	for i := 0; i < nThreads; i++ {
		go insertThread(i)
	}

	time.Sleep(p.duration)

	isRunning = false
	wg.Wait()
	p.testWG.Done()
}
