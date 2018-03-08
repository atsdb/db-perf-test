/*
* @Author: ronan
* @Date:   2018-03-04 10:39:07
* @Last Modified by:   ron
* @Last Modified time: 2018-03-08 07:58:14
 */
package dbtest

import (
	"dbdriver"
	"log"
	"math/rand"
)

func testTable(dbcon string, engine string, ttype string, prefix string) (dbdriver.Table, func() []interface{}) {

	var driver dbdriver.Driver
	if engine == "postgres" {
		driver = dbdriver.NewPostgresDriver(dbcon)
	} else if engine == "InnoDB" || engine == "MyISAM" {
		driver = dbdriver.NewMysqlDriver(dbcon, engine)
	} else if engine == "clickhouse" {
		driver = dbdriver.NewClickHouseDriver(dbcon)
	}

	switch ttype {
	case "large":
		return testLargeTable(driver, engine, prefix)
	case "slim":
		return testSlimTable(driver, engine, prefix)
	case "light":
		return testLightTable(driver, engine, prefix)
	case "light-with-index":
		return testLightTableWithIndex(driver, engine, prefix)
	default:
		log.Fatal("Unknown test table configuration: ", ttype)
	}
	return nil, nil
}

func testLargeTable(driver dbdriver.Driver, engine string, prefix string) (dbdriver.Table, func() []interface{}) {

	tableName := "test-large-table-" + prefix + engine
	db, _ := driver.Create(tableName, []string{
		"idx:index",
		"value:varchar(200)",
		"type1:int",
		"type2:int",
		"type3:int",
		"type4:int",
		"type5:int",
		"type6:int",
		"type7:int",
		"type8:int",
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

func testSlimTable(driver dbdriver.Driver, engine string, prefix string) (dbdriver.Table, func() []interface{}) {

	tableName := "test-slim-table-" + prefix + engine
	db, _ := driver.Create(tableName, []string{
		"idx:int",
		"value:int",
	})

	generator := func() []interface{} {
		return []interface{}{
			rand.Int31(),
			rand.Int31(),
		}
	}

	return db, generator
}

func testLightTable(driver dbdriver.Driver, engine string, prefix string) (dbdriver.Table, func() []interface{}) {

	tableName := "test-light-table-" + prefix + engine
	db, _ := driver.Create(tableName, []string{
		"idx:int",
		"value:int",
		"col1:int16",
		"col2:int16",
	})

	generator := func() []interface{} {
		return []interface{}{
			0,
			rand.Int31(),
			rand.Int31() & 0xff,
			rand.Int31() & 0x7fff,
		}
	}

	return db, generator
}

func testLightTableWithIndex(driver dbdriver.Driver, engine string, prefix string) (dbdriver.Table, func() []interface{}) {

	tableName := "test-light-table-index-" + prefix + engine
	db, _ := driver.Create(tableName, []string{
		"idx:index",
		"value:int",
		"col1:int16:index",
		"col2:int16",
	})

	generator := func() []interface{} {
		return []interface{}{
			0,
			rand.Int31(),
			rand.Int31() & 0xff,
			rand.Int31() & 0x7fff,
		}
	}

	return db, generator
}
