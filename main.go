/*
* @Author: ron
* @Date:   2017-05-03 09:12:55
* @Last Modified by:   ronan
* @Last Modified time: 2018-03-04 11:45:23
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
	flag.Parse()

	if *mode != "read" && *mode != "write" {
		fmt.Println("Please specify the mode (read or write)")
		return
	}

	dbtest.DoPerfTest(*db, *mode, *engine, *duration)
}
