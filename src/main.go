/*
* @Author: ron
* @Date:   2017-05-03 09:12:55
* @Last Modified by:   ron
* @Last Modified time: 2018-03-06 10:42:31
 */

package main

import (
	"dbtest"
	"flag"
)

func main() {

	var db = flag.String("db", "user:password@/database", "db connection")
	var mode = flag.String("mode", "", "read or write")
	var engine = flag.String("engine", "", "MyISAM or InnoDB")
	var duration = flag.Int("duration", 60, "Test duration in seconds")
	var nthreads = flag.Int("threads", 10, "Number of concurrent threads for read/write ops")
	var ntables = flag.Int("tables", 10, "Number of concurrent tables to write to")
	var table = flag.String("table", "light-with-key", "Table type (large or light)")

	var clean = flag.Bool("clean", false, "True to delete the data from the write table")

	flag.Parse()

	if *clean {
		dbtest.DoClean(*db, *engine)
	} else {
		dbtest.DoPerfTest(*db, *mode, *engine, *table, *duration, *nthreads, *ntables)
	}
}
