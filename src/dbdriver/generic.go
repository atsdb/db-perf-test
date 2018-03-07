/*
* @Author: ronan
* @Date:   2018-03-04 09:39:00
* @Last Modified by:   ron
* @Last Modified time: 2018-03-06 12:14:43
 */
package dbdriver

import (
	"database/sql"
)

type GenericTable struct {
	Table
	db     *sql.DB
	fields []string
	types  []string
	table  string
}

func (d GenericTable) Name() string {
	return "`" + d.table + "`"
}

func (d GenericTable) DB() *sql.DB {
	return d.db
}

func (d GenericTable) Close() {
	d.db.Close()
}

func (d GenericTable) Columns() []string {
	return d.fields
}
