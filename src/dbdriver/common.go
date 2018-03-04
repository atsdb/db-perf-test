/*
* @Author: ronan
* @Date:   2018-03-04 09:39:00
* @Last Modified by:   ron
* @Last Modified time: 2018-03-04 17:40:18
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

	/* close the DB connection */
	Close()
}

type Driver interface {
	Create(table string, fields []string) (Table, error)
}
