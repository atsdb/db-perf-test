/*
* @Author: ronan
* @Date:   2018-03-04 10:21:02
* @Last Modified by:   ron
* @Last Modified time: 2018-03-06 16:07:14
 */
package dbperf

import (
	"fmt"
	"log"
	"sync"
)

func (p *PerformanceMonitor) Read(from int64, to int64) {

	p.testWG.Add(1)

	where := fmt.Sprintf("where idx >=%d and idx<%d", from, to)
	query := "select * from " + p.table.Name() + " " + where
	stmt, err := p.table.DB().Prepare(query)
	if err != nil {
		log.Fatal("[Read] Can not prepare query: ", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	id2m := make([]interface{}, (to - from + 1))
	m2id := make(map[interface{}]uint32)

	idxsync := make(chan int64)

	readThread := func() {

		rows, err := stmt.Query()
		if err != nil {
			log.Fatal("[Read] Can not run query: ", err)
		}

		var idx int64
		var value interface{}
		nreads := 0
		params := []interface{}{}
		for _, name := range p.table.Columns() {
			if name == "value" {
				params = append(params, &value)
			} else if name == "idx" {
				params = append(params, &idx)
			} else {
				var i interface{}
				params = append(params, &i)
			}
		}
		for rows.Next() {
			e := rows.Scan(params...)
			id2m[idx-from] = value
			nreads += 1
			p.Inc(e)
			if nreads >= 1000*1000 {
				idxsync <- idx - from
				nreads = 0
			}
		}

		idxsync <- 0
	}

	mapThread := func() {
		cidx := int64(1)
		for idx := range idxsync {
			if idx == 0 {
				break
			}
			for cidx <= idx {
				name := id2m[cidx]
				m2id[name] = uint32(cidx)
				cidx += 1
			}
		}
		wg.Done()
	}

	go readThread()
	go mapThread()

	wg.Wait()
	p.testWG.Done()
}
