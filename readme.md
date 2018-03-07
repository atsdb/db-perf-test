
# DB Performance Testing

This tool can be used to benchmark various Database performance under "no-sql" usage conditions.

[![Build Status](https://travis-ci.org/atsdb/db-perf-test.svg?branch=master)](https://travis-ci.org/atsdb/db-perf-test)

## Database/Table Configuration

The database are used in "no-sql" mode, i.e. without any join between table. The objective of the test is to insert as fast a possible into the table, and read-back as fast as possible.

Three different configurations for the tables are defined (see file `dbtest/config.go`):

 * light table:
     - First column: index (int, primary key)
     - Second Column: integer
     - Third Column: integer

 * light table with index:
     - First column: index (int, primary key)
     - Second Column: integer (4 bytes)
     - Third Column: small integer (2 bytes) with key (non primary)
     - Third Column: small integer (2 bytes)

 * large table:
     - First column: index (int, primary key)
     - Second Column: var char of 200 characters
     - Third to 10th Column: integer

## Tests Scenarios

 * Write

     - Insert intto a table with 10 concurrent threads without transaction
     - Insert intto a table with 10 concurrent threads using a transaction flushed every minute (only one thread for the transaction flush).
     - Insert into 10 tables concurrently and one transaction for each table.

  * Read 

      - Read the whole table data with one single thread
      - Read the table data with 2 theads: the first threads read the first half of the values, and the second thread reads the second half.
      - Read the table data with 3 and more theads: split the tables in as many chuncks as there are threads, and read concurrently.
      - The precondition for the read test is to fill-in the table with at least 100 million rows. This tool will take care of filling-in the table if needed.

## Command Line Usage

* Mysql 

   - `make write-innodb` 
   - `make read-innodb` 
   - `make write-myisam` 
   - `make read-myisam` 

 * Postgress

   - `make write-postgres` 
   - `make read-postgres`

