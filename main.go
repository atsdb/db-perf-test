/*
* @Author: ron
* @Date:   2017-05-03 09:12:55
* @Last Modified by:   ron
* @Last Modified time: 2018-03-04 17:37:06
 */

package main

import (
	"dbtest"
	"flag"
	"fmt"
)

func main() {

	var db = flag.String("db", "user:password@/database", "db connection")
	var mode = flag.String("mode", "", "read or write")
	var engine = flag.String("engine", "", "MyISAM or InnoDB")
	var duration = flag.Int("duration", 60, "Test duration in seconds")
	var concurrent = flag.Int("concurrent", 10, "Number of concurrent threeads/connections")
	var table = flag.String("table", "light", "Table type (large or light)")

	flag.Parse()

	if *mode != "read" && *mode != "write" {
		fmt.Println("Please specify the mode (read or write)")
		return
	}

	dbtest.DoPerfTest(*db, *mode, *engine, *table, *duration, *concurrent)
}
