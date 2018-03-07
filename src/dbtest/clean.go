/*
* @Author: ronan
* @Date:   2018-03-04 10:42:09
* @Last Modified by:   ron
* @Last Modified time: 2018-03-06 11:02:01
 */
package dbtest

import (
	"dbdriver"
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
	if row := table.DB().QueryRow(query); row != nil {
		var nrows int64
		row.Scan(&nrows)
		if nrows > 0 {
			log.Printf("Table %-42s: %d rows\n", table.Name(), nrows)
			table.DB().Exec("drop table " + table.Name())
		}
	}

}
