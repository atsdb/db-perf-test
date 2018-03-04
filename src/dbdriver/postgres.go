/*
* @Author: ronan
* @Date:   2018-03-04 10:01:22
* @Last Modified by:   ron
* @Last Modified time: 2018-03-04 13:38:45
 */
package dbdriver

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"log"
	"strings"
)

type PostgresDriver struct {
	db *sql.DB
}

type PostgresTable struct {
	*PostgresDriver
	fields []string
	types  []string
	table  string
}

func NewPostgresDriver(conn string) *PostgresDriver {

	db, err1 := sql.Open("postgres", conn)
	if err1 != nil {
		log.Fatal("[atsdb] Unable to open the DB!", err1)
	}

	e1 := db.Ping()
	if e1 != nil {
		log.Fatal("[atsdb] Unable to open DB to '"+conn+"': ", e1)
	}

	return &PostgresDriver{
		db: db,
	}
}

func (d *PostgresDriver) Create(table string, fields []string) (Table, error) {

	fieldTypes := make([]string, 0)
	fieldNames := make([]string, 0)
	tablefields := ""
	tablekeys := ""
	for i, field := range fields {
		if i != 0 {
			tablefields += ",\n"
		}
		p := strings.Split(field, ":")
		tablefields += `"` + p[0] + `" `
		switch p[1] {
		case "index":
			tablefields += "serial"
			tablekeys += ",PRIMARY KEY (\"" + p[0] + "\")"
		case "int":
			tablefields += "integer"
		case "string":
			tablefields += "varchar(500)"
		default:
			tablefields += p[1]
		}
		fieldNames = append(fieldNames, p[0])
		fieldTypes = append(fieldTypes, p[1])
		table += "-" + field
	}
	cquery := "create table IF NOT EXISTS \"" + table + "\" (\n" + tablefields + "\n" + tablekeys + "\n); "

	if _, err := d.db.Exec(cquery); err != nil {
		log.Fatal("[Create] Can not create table: ", err, "\n", cquery)
		return PostgresTable{}, err
	}
	return PostgresTable{
		PostgresDriver: d,
		fields:         fieldNames,
		table:          table,
		types:          fieldTypes,
	}, nil
}

func (d PostgresTable) Name() string {
	return `"` + d.table + `"`
}

func (d PostgresTable) DB() *sql.DB {
	return d.db
}

func (d PostgresTable) PrepareInsert() QueryRunner {
	return d.prepareInsert(nil)
}

func (d PostgresTable) PrepareTxInsert() (*sql.Tx, QueryRunner) {

	tx, err := d.db.Begin()
	if err != nil {
		log.Fatal("[write] can not start the transaction", err)
	}

	runner := d.prepareInsert(tx)
	return tx, runner
}

func (d *PostgresTable) prepareInsert(tx *sql.Tx) QueryRunner {
	values := make([]string, 0)
	fields := make([]string, 0)

	count := 0
	hasIndex := -1
	for i, _ := range d.fields {
		if d.types[i] != "index" {
			values = append(values, fmt.Sprintf("$%d", count+1))
			fields = append(fields, d.fields[i])
			count += 1
		} else {
			hasIndex = i
		}
	}
	svalues := strings.Join(values, ",")
	sfields := strings.Join(fields, ",")

	query := "insert into \"" + d.table + "\" (" + sfields + ") values (" + svalues + ") "

	var stmt *sql.Stmt
	var err error
	if tx != nil {
		stmt, err = tx.Prepare(query)
	} else {
		stmt, err = d.db.Prepare(query)
	}

	if err != nil {
		log.Fatal("[PrepareInsert] Can not prepare query: ", err, "\n", query)
		return nil
	}

	return func(args ...interface{}) error {
		/* Assume that the index is the first column */
		if hasIndex >= 0 {
			args = args[1:]
		}
		_, e := stmt.Exec(args...)
		return e
	}

}
