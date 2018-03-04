/*
* @Author: ronan
* @Date:   2018-03-04 09:39:00
* @Last Modified by:   ronan
* @Last Modified time: 2018-03-04 11:03:33
 */
package dbdriver

import (
	"database/sql"
)

type QueryRunner func(args ...interface{}) error

type Table interface {
	Name() string

	/* Prepare without transaction */
	PrepareInsert() QueryRunner

	/* Prepare on the given transaction */
	PrepareTxInsert() (*sql.Tx, QueryRunner)

	/* Used for executing random queries */
	DB() *sql.DB
}

type Driver interface {
	Create(table string, fields []string) (Table, error)
}
