package dbtest

// import (
// 	"fmt"
// 	"sync"
// 	"sync/atomic"
// 	"time"
// )

// func (p *PerformanceMonitor) WriteNoTx(d time.Duration, generator func() []interface{}) {

// 	p.wg.Add(1)
// 	total := int64(0)
// 	startTime := time.Now()
// 	isRunning := true

// 	insert := p.table.PrepareInsert()

// 	insertThread := func() {

// 		for isRunning {

// 			values := generator()
// 			err := insert(values...)
// 			atomic.AddInt64(&total, 1)
// 			p.Inc(err)

// 		}

// 	}

// 	go insertThread()
// 	time.Sleep(d)

// 	isRunning = false
// 	dt := time.Now().Sub(startTime)
// 	rps := float64(total) * float64(time.Millisecond) / float64(dt)
// 	fmt.Printf("[write] %.2fK insert/sec [%d in %v]\n", rps, total, dt)
// 	p.wg.Done()
// }

// func (p *PerformanceMonitor) WriteTx(d time.Duration, generator func() []interface{}, nThreads int) {

// 	p.wg.Add(1)
// 	total := int64(0)
// 	startTime := time.Now()
// 	isRunning := true

// 	tx, insert := p.table.PrepareTxInsert()

// 	var done chan bool
// 	transactionMutex := &sync.RWMutex{}
// 	insertThread := func() {

// 		for isRunning {
// 			values := generator()

// 			transactionMutex.RLock()
// 			err := insert(values...)
// 			transactionMutex.RUnlock()

// 			atomic.AddInt64(&total, 1)
// 			p.Inc(err)
// 		}
// 	}

// 	commitThread := func() {

// 		for isRunning {
// 			time.Sleep(time.Second * 5)
// 			transactionMutex.Lock()
// 			tx.Commit()
// 			tx, insert = p.table.PrepareTxInsert()
// 			transactionMutex.Unlock()
// 		}
// 		done <- true

// 	}

// 	go commitThread()
// 	for i := 0; i < nThreads; i++ {
// 		go insertThread()
// 	}
// 	time.Sleep(d)

// 	isRunning = false
// 	<-done
// 	tx.Commit()

// 	dt := time.Now().Sub(startTime)
// 	rps := float64(total) * float64(time.Millisecond) / float64(dt)
// 	fmt.Printf("[write/tx] %.2fK insert/sec [%d in %v]\n", rps, total, dt)
// 	p.wg.Done()
// }
