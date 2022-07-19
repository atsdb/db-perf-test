/*
* @Author: ronan
* @Date:   2018-03-04 10:42:09
* @Last Modified by:   ronanj
* @Last Modified time: 2022-07-19 08:16:34
 */
package dbtest

import (
	"dbperf/src/dbdriver"
	"fmt"
	"log"
)

func DoClean(dbcon string, engine string) {

	for _, testType := range []string{"light", "light-with-index", "large"} {
		for _, prefix := range []string{"notx-", "tx-"} {
			table, _ := testTable(dbcon, engine, testType, prefix)
			cleanupTable(table)
		}

		for i := 0; i < 10; i++ {
			table, _ := testTable(dbcon, engine, testType, fmt.Sprintf("conn%d-", i))
			cleanupTable(table)

		}

		log.Printf("Cleanup done for %s + %s\n", engine, testType)

	}

}

func cleanupTable(table dbdriver.Table) {
	query := "select max(idx) from " + table.Name()
	if rows, err := table.DB().Query(query); err != nil {
		log.Printf("Table %-42s: does not exists [%s]: (%v)\n", table.Name(), query, err)
	} else if rows != nil {
		defer rows.Close()
		var nrows int64
		if rows.Next() {
			rows.Scan(&nrows)
			if nrows > 0 {
				log.Printf("Table %-42s: %d rows\n", table.Name(), nrows)
				if _, err := table.DB().Exec("drop table " + table.Name()); err != nil {
					log.Printf("Table %-42s: can not delete (%v)\n", table.Name(), err)
				}
			}
		}
	}

}
