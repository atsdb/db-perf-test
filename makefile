

PACKAGE = dbperf
GOSRCS  = $(wildcard *.go)
GODEPS  = $(wildcard src/dbperf/*.go) \
		  $(wildcard src/dbtest/*.go) \
		  $(wildcard src/dbdriver/*.go) \
		  $(GOSRCS)

#--------------------------------------------------


all:
	@echo "Usage: make read|write-innodb|myisam|postgres. Eg make write-myisam"


ifeq ($(PGXDB),)
PGXDB = "postgres://testuser:@localhost/perftest?sslmode=disable"
endif

ifeq ($(MYSQLDB),)
MYSQLDB = "testuser:12345@/testdb"
endif

ifeq ($(DURATION),)
DURATION = 60
endif

ifeq ($(TABLE),)
#Possible choices: light, large
TABLE = light
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
