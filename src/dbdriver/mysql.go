/*
* @Author: ronan
* @Date:   2018-03-04 10:22:12
* @Last Modified by:   ron
* @Last Modified time: 2018-03-06 12:25:22
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

type MysqlDriver struct {
	mysqldb *sql.DB
	engine  string
}

type MysqlTable struct {
	GenericTable
	MysqlDriver
}

func NewMysqlDriver(conn string, engine string) MysqlDriver {

	db, err1 := sql.Open("mysql", conn)
	if err1 != nil {
		log.Fatal("[atsdb] Unable to open the DB!", err1)
	}

	e1 := db.Ping()
	if e1 != nil {
		log.Fatal("[atsdb] Unable to open DB to '"+conn+"': ", e1)
	}

	return MysqlDriver{
		engine:  engine,
		mysqldb: db,
	}
}

/* Create a new connection for each table */

func (d MysqlDriver) Create(table string, fields []string) (Table, error) {

	fieldTypes := make([]string, 0)
	fieldNames := make([]string, 0)
	tablefields := ""
	tablekeys := ""
	md5 := md5.New()
	for i, field := range fields {
		if i != 0 {
			tablefields += ",\n"
		}
		p := strings.Split(field, ":")
		tablefields += "`" + p[0] + "` "
		switch p[1] {
		case "index":
			tablefields += "int unsigned NOT NULL AUTO_INCREMENT"
			tablekeys += ",PRIMARY KEY (`" + p[0] + "`)"
		case "int":
			tablefields += "int"
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
				tablekeys += ",KEY (`" + p[0] + "`)"

			default:
				log.Fatal("Unknown field qualifier " + field)
			}
		}

		io.WriteString(md5, field)
	}

	table += "-" + fmt.Sprintf("%x", md5.Sum(nil))[:4]

	cquery := "create table IF NOT EXISTS `" + table + "` (\n" + tablefields + "\n" + tablekeys + "\n) "
	cquery += "ENGINE=" + d.engine + "  DEFAULT CHARSET=utf8;"

	if _, err := d.mysqldb.Exec(cquery); err != nil {
		log.Fatal("[Create] Can not create table: ", err, "\n\n", cquery)
		return MysqlTable{}, err
	}
	return MysqlTable{
		MysqlDriver: d,
		GenericTable: GenericTable{
			db:     d.mysqldb,
			fields: fieldNames,
			table:  table,
			types:  fieldTypes,
		},
	}, nil
}

func (d MysqlTable) PrepareInsert() QueryRunner {
	return d.prepareInsert(nil)
}

func (d MysqlTable) PrepareTxInsert() (*sql.Tx, QueryRunner) {

	tx, err := d.db.Begin()
	if err != nil {
		log.Fatal("[write] can not start the transaction", err)
	}

	runner := d.prepareInsert(tx)
	return tx, runner
}

func (d *MysqlTable) prepareInsert(tx *sql.Tx) QueryRunner {
	values := make([]string, 0)
	fields := make([]string, 0)

	for i, _ := range d.fields {
		values = append(values, "?")
		fields = append(fields, d.fields[i])
	}
	svalues := strings.Join(values, ",")
	sfields := strings.Join(fields, ",")

	query := "insert into `" + d.table + "` (" + sfields + ") values (" + svalues + ") "

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
		_, e := stmt.Exec(args...)
		return e
	}

}
