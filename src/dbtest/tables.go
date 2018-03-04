/*
* @Author: ronan
* @Date:   2018-03-04 10:39:07
* @Last Modified by:   ron
* @Last Modified time: 2018-03-04 14:54:31
 */
package dbtest

import (
	"dbdriver"
	"log"
	"math/rand"
)

func testTable(dbcon string, engine string, ttype string, prefix string) (dbdriver.Table, func() []interface{}) {
	switch ttype {
	case "large":
		return testLargeTable(dbcon, engine, prefix)
	case "light":
		return testLightTable(dbcon, engine, prefix)
	default:
		log.Fatal("Unknown test table configuration: ", ttype)
	}
	return nil, nil
}

func testLargeTable(dbcon string, engine string, prefix string) (dbdriver.Table, func() []interface{}) {

	var driver dbdriver.Driver
	if engine == "postgres" {
		driver = dbdriver.NewPostgresDriver(dbcon)
	} else {
		driver = dbdriver.NewMysqlDriver(dbcon, engine)
	}

	tableName := "test-large-table-" + prefix + engine
	db, _ := driver.Create(tableName, []string{
		"idx:index",
		"value:varchar(200)",
		"col1:int",
		"col2:int",
		"col3:int",
		"col4:int",
		"col5:int",
		"col6:int",
		"col7:int",
		"col8:int",
	})

	generator := func() []interface{} {
		return []interface{}{
			0,
			RandomString(200),
			rand.Int31(),
			rand.Int31(),
			rand.Int31(),
			rand.Int31(),
			rand.Int31(),
			rand.Int31(),
			rand.Int31(),
			rand.Int31(),
		}
	}

	return db, generator
}

func testLightTable(dbcon string, engine string, prefix string) (dbdriver.Table, func() []interface{}) {

	var driver dbdriver.Driver
	if engine == "postgres" {
		driver = dbdriver.NewPostgresDriver(dbcon)
	} else {
		driver = dbdriver.NewMysqlDriver(dbcon, engine)
	}

	tableName := "test-light-table-" + prefix + engine
	db, _ := driver.Create(tableName, []string{
		"idx:index",
		"col1:int",
		"col2:int",
	})

	generator := func() []interface{} {
		return []interface{}{
			0,
			rand.Int31(),
			rand.Int31(),
		}
	}

	return db, generator
}
