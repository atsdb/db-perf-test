/*
* @Author: ronan
* @Date:   2018-03-04 10:01:22
* @Last Modified by:   ron
* @Last Modified time: 2018-03-06 17:21:44
 */
package dbdriver

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"io"
	"log"
	"strings"
)

type PostgresDriver struct {
	pgxdb *sql.DB
}

type PostgresTable struct {
	GenericTable
	PostgresDriver
}

func NewPostgresDriver(conn string) PostgresDriver {

	db, err1 := sql.Open("postgres", conn)
	if err1 != nil {
		log.Fatal("[atsdb] Unable to open the DB!", err1)
	}

	e1 := db.Ping()
	if e1 != nil {
		log.Fatal("[atsdb] Unable to open DB to '"+conn+"': ", e1)
	}

	return PostgresDriver{
		pgxdb: db,
	}
}

func (d PostgresDriver) Create(table string, fields []string) (Table, error) {

	fieldTypes := make([]string, 0)
	fieldNames := make([]string, 0)
	tablefields := ""
	tablekeys := ""
	postquery := ""

	md5 := md5.New()
	for _, field := range fields {
		io.WriteString(md5, field)
	}
	table += "-" + fmt.Sprintf("%x", md5.Sum(nil))[:4]

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
		case "int16":
			tablefields += "smallint"
		case "string":
			tablefields += "varchar(500)"
		default:
			tablefields += p[1]
		}
		fieldNames = append(fieldNames, p[0])
		fieldTypes = append(fieldTypes, p[1])

		if len(p) > 2 {
			switch p[2] {
			case "index":
				postquery += `CREATE INDEX ON "` + table + `"("` + p[0] + `");`

			default:
				log.Fatal("Unknown field qualifier " + field)
			}
		}
	}

	cquery := "create table IF NOT EXISTS \"" + table + "\" (\n" + tablefields + "\n" + tablekeys + "\n); "
	cquery += postquery

	if _, err := d.pgxdb.Exec(cquery); err != nil {
		log.Fatal("[Create] Can not create table: ", err, "\n", cquery)
		return PostgresTable{}, err
	}
	return PostgresTable{
		PostgresDriver: d,
		GenericTable: GenericTable{
			db:     d.pgxdb,
			fields: fieldNames,
			table:  table,
			types:  fieldTypes,
		},
	}, nil
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
