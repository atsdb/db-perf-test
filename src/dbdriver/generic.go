/*
* @Author: ronan
* @Date:   2018-03-04 09:39:00
* @Last Modified by:   ron
* @Last Modified time: 2018-03-07 15:19:53
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

func (d GenericTable) DB() *sql.DB {
	return d.db
}

func (d GenericTable) Close() {
	d.db.Close()
}

func (d GenericTable) Columns() []string {
	return d.fields
}
