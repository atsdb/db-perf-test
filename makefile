

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

ifeq ($(CLKDB),)
CLKDB = "http://testuser:12345@tcp(127.0.0.1:9000)/testdb"
endif

ifeq ($(DURATION),)
DURATION = 60
endif

ifeq ($(TABLE),)
#Possible choices: slim, light, light-with-index, large
TABLE = "light"
endif

#--------------------------------------------------


build: bin/$(PACKAGE)

bin/$(PACKAGE): $(GODEPS)
	go fmt ./...
	go build -o bin/$(PACKAGE) main.go 
	@mkdir -p reports results

#--------------------------------------------------

run-all: build run-write run-read

#-------------------------------------------------- WRITE

run-write:
	make perform-all-engines MODE=write DURATION=$(DURATION) TABLE=light
	make perform-all-engines MODE=write  DURATION=$(DURATION) TABLE=light-with-index
	make perform-all-engines MODE=write  DURATION=$(DURATION) TABLE=large

#-------------------------------------------------- READ

run-read:
	make perform-all-engines MODE=read DURATION=$(DURATION) TABLE=light
	make perform-all-engines MODE=read DURATION=$(DURATION) TABLE=light-with-index
	make perform-all-engines MODE=read DURATION=$(DURATION) TABLE=large

#--------------------------------------------------

perform-all-engines:
	bin/dbperf --mode $(MODE) --engine MyISAM --db $(MYSQLDB) --duration $(DURATION) --table $(TABLE)
	bin/dbperf --mode $(MODE) --engine InnoDB --db $(MYSQLDB) --duration $(DURATION) --table $(TABLE)
	bin/dbperf --mode $(MODE) --engine postgres --db $(PGXDB) --duration $(DURATION) --table $(TABLE)

#--------------------------------------------------

cleandb: build
	bin/dbperf --clean --db $(MYSQLDB) --engine InnoDB
	bin/dbperf --clean --db $(MYSQLDB) --engine MyISAM
	bin/dbperf --clean --db $(PGXDB) --engine postgres

#--------------------------------------------------

write-clickhouse: build
	bin/dbperf --mode write --engine clickhouse --db $(CLKDB) --duration $(DURATION) --table $(TABLE)

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



