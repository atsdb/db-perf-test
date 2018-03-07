

PACKAGE = dbperf
GOSRCS  = $(wildcard *.go)
GODEPS  = $(wildcard src/dbperf/*.go) \
		  $(wildcard src/dbtest/*.go) \
		  $(wildcard src/dbdriver/*.go) \
		  $(GOSRCS)

#--------------------------------------------------


all: build
	@echo "Usage: make read|write-innodb|myisam|postgres. Eg make write-myisam"


ifeq ($(PGXDB),)
PGXDB = "postgres://testuser:12345@localhost/testdb?sslmode=disable"
endif

ifeq ($(MYSQLDB),)
MYSQLDB = "testuser:12345@/testdb"
endif

ifeq ($(DURATION),)
DURATION = 60
endif

ifeq ($(TABLE),)
#Possible choices: light, light-with-index, large
TABLE = "light-with-index"
endif

#--------------------------------------------------


build: bin/$(PACKAGE)

bin/$(PACKAGE): $(GODEPS)
	@GOPATH=`pwd` go fmt main.go
	@GOPATH=`pwd` go fmt dbperf
	@GOPATH=`pwd` go fmt dbtest
	@GOPATH=`pwd` go fmt dbdriver
	GOPATH=`pwd` go build -o bin/$(PACKAGE) main.go 
	@mkdir -p reports

#--------------------------------------------------

run-all: run-write run-read

run-write: run-mysql-write run-postgres-write

run-postgres-write: build
	bin/dbperf --mode write --engine postgres --db $(PGXDB) --duration $(DURATION) --table light
	bin/dbperf --mode write --engine postgres --db $(PGXDB) --duration $(DURATION) --table light-with-index
	bin/dbperf --mode write --engine postgres --db $(PGXDB) --duration $(DURATION) --table large

run-mysql-write: build
	bin/dbperf --mode write --engine InnoDB --db $(MYSQLDB) --duration $(DURATION) --table light
	bin/dbperf --mode write --engine MyISAM --db $(MYSQLDB) --duration $(DURATION) --table light
	bin/dbperf --mode write --engine InnoDB --db $(MYSQLDB) --duration $(DURATION) --table light-with-index
	bin/dbperf --mode write --engine MyISAM --db $(MYSQLDB) --duration $(DURATION) --table light-with-index
	bin/dbperf --mode write --engine InnoDB --db $(MYSQLDB) --duration $(DURATION) --table large
	bin/dbperf --mode write --engine MyISAM --db $(MYSQLDB) --duration $(DURATION) --table large

run-read: run-mysql-read run-postgres-read

run-postgres-read: build
	bin/dbperf --mode read --engine postgres --db $(PGXDB) --duration $(DURATION) --table light
	bin/dbperf --mode read --engine postgres --db $(PGXDB) --duration $(DURATION) --table large

run-mysql-read: build
	bin/dbperf --mode read --engine InnoDB --db $(MYSQLDB) --duration $(DURATION) --table light
	bin/dbperf --mode read --engine MyISAM --db $(MYSQLDB) --duration $(DURATION) --table light
	bin/dbperf --mode read --engine InnoDB --db $(MYSQLDB) --duration $(DURATION) --table large
	bin/dbperf --mode read --engine MyISAM --db $(MYSQLDB) --duration $(DURATION) --table large

#--------------------------------------------------

cleandb: build
	bin/dbperf --clean --db $(MYSQLDB) --engine InnoDB
	bin/dbperf --clean --db $(MYSQLDB) --engine MyISAM
	bin/dbperf --clean --db $(PGXDB) --engine postgres

#--------------------------------------------------

write-postgres: build
	bin/dbperf --mode write --engine postgres --db $(PGXDB) --duration $(DURATION) --table $(TABLE)

read-postgres: build
	bin/dbperf --mode read --engine postgres --db $(PGXDB) --duration $(DURATION) --table $(TABLE)

write-innodb: build
	bin/dbperf --mode write --engine InnoDB --db $(MYSQLDB) --duration $(DURATION) --table $(TABLE)

read-innodb: build
	bin/dbperf --mode read --engine InnoDB --db $(MYSQLDB) --duration $(DURATION) --table $(TABLE)

write-myisam: build
	bin/dbperf --mode write --engine MyISAM --db $(MYSQLDB)  --duration $(DURATION) --table $(TABLE)

read-myisam: build
	bin/dbperf --mode read --engine MyISAM  --db $(MYSQLDB) --duration $(DURATION) --table $(TABLE)
