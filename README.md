# Go-UTF8

## What it does
Checks if utf8 encoded data is being stored in  latin1 columns.




## ~/my.cnf Used for testing
````ini
Make sure there are no spaces before and after the = sign.  The spaces will cause the connection to fail.
Do not use single quotes around the password.  

The single quotes will cause the connection to fail. If you have special characters in your password, you will need to escape them with a backslash.

[mysql]
#default-character-set=latin1
[client_ps]
user=root
password=xxxxx
host=192.168.50.10
[client_primary]
user=dba_util
password=xxxxx
host=192.168.50.152
[client_replica]
user=dba_util
password=xxxxx
host=192.168.50.153
[client_etlreplica]
user=dba_util
password=xxxxx
host=192.168.50.154
````

## How to use
```sql
You can then use it like so:
mysql --defaults-group-suffix=_primary -e "SELECT CONVERT(CONVERT(description USING BINARY) USING latin1) AS latin1, CONVERT(CONVERT(description USING BINARY) USING utf8) AS utf8 FROM char_test_db.t1 WHERE CONVERT(description USING BINARY) RLIKE CONCAT('[', UNHEX('60'), '-', UNHEX('FF'), ']')"
+-------------------------+----------------+
| latin1                  | utf8           |
+-------------------------+----------------+
| Ã‚Â¡VolcÃƒÂ¡n!          | Â¡VolcÃ¡n!     |
+-------------------------+----------------+
```


## Tetsting
```sql
Using queries from https://www.percona.com/blog/utf8-data-on-latin1-tables-converting-to-utf8-without-downtime-or-double-encoding/

mysql> SET NAMES latin1;
Query OK, 0 rows affected (0.00 sec)

mysql> SELECT id, description, HEX(description) FROM t1;
+----+-------------+----------------------+
| id | description | HEX(description)     |
+----+-------------+----------------------+
|  1 | ¡Volcán!    | C2A1566F6C63C3A16E21 |
+----+-------------+----------------------+
1 row in set (0.00 sec)


mysql> SET NAMES utf8;
Query OK, 0 rows affected, 1 warning (0.00 sec)


mysql> SELECT id, description, HEX(description) FROM t1;
+----+----------------+----------------------+
| id | description    | HEX(description)     |
+----+----------------+----------------------+
|  1 | Â¡VolcÃ¡n!     | C2A1566F6C63C3A16E21 |
+----+----------------+----------------------+
1 row in set (0.00 sec)


mysql> SELECT CONVERT(CONVERT(description USING BINARY) USING latin1) AS latin1, CONVERT(CONVERT(description USING BINARY) USING utf8) AS utf8 FROM t1 WHERE CONVERT(description USING BINARY) RLIKE CONCAT('[', UNHEX('60'), '-', UNHEX('FF'), ']');
+----------------+------------+
| latin1         | utf8       |
+----------------+------------+
| Â¡VolcÃ¡n!     | ¡Volcán!   |
+----------------+------------+
1 row in set, 1 warning (0.00 sec)



mysql> SET NAMES latin1;
Query OK, 0 rows affected (0.00 sec)

mysql> SELECT CONVERT(CONVERT(description USING BINARY) USING latin1) AS latin1, CONVERT(CONVERT(description USING BINARY) USING utf8) AS utf8 FROM  t1 WHERE CONVERT(description USING BINARY) RLIKE CONCAT('[', UNHEX('60'), '-', UNHEX('FF'), ']');
+------------+----------+
| latin1     | utf8     |
+------------+----------+
| ¡Volcán!   | Volc�!   |
+------------+----------+
1 row in set, 1 warning (0.00 sec)

```

### How to use the cli
```go   
go-utf8 on  main via 🐹 v1.21.0 
❯ go-utf8
Usage: go-utf8 -s <source host> [-d <database name>] [-show] [-t <table name>]
Please specify a source host



go-utf8 on  main [!?] via 🐹 v1.21.0 
❯ go-utf8 -h
Usage of go-utf8:
  -d string
        Database Name
  -s string
        Source Host
  -show
        Show Databases
  -t string
        Select table



go-utf8 on  main [!?] via 🐹 v1.21.0 
❯ go-utf8 -s primary -show
Connected to primary (primary): ✔

char_test_db
information_schema
mysql
performance_schema
sys


go-utf8 on  main [!?] via 🐹 v1.21.0 
❯ go-utf8 -s primary -d char_test_db
Connected to primary (primary): ✔


Current table: t1
Column: name
Count of records that need to be fixed: 18


Current table: t1
Column: address1
Count of records that need to be fixed: 42


Current table: t1
Column: address2
Count of records that need to be fixed: 1


go-utf8 on  main [!?] via 🐹 v1.21.0 
❯ go-utf8 -s primary -d char_test_db -t t1
Connected to primary (primary): ✔

Default character set: utf8
Default collation: utf8_unicode_ci


Scan a single table:
go-utf8 -s ChaosHour.com -d chaos -t Ads -scan
Connected to ChaosHour.com (localhost): ✔

Current table: Ads
Column: add_phone
Count of records that need to be fixed: 8

Offending IDs:
[17454 23951 27531 34206 35838 39497 39889 44513]
```



## reference

- https://oracle-base.com/articles/mysql/mysql-converting-table-character-sets-from-latin1-to-utf8#the-problem
- https://www.percona.com/blog/utf8-data-on-latin1-tables-converting-to-utf8-without-downtime-or-double-encoding/
- https://mysql.rjweb.org/doc.php/charcoll#diagnosing_charset_issues
- https://www.percona.com/blog/fixing-column-encoding-mess-in-mysql/


## How to install
```Go

go install github.com/ChaosHour/go-utf8@latest


To build:

go build -o go-utf8

FreeBSD:
env GOOS=freebsd GOARCH=amd64 go build .

On Mac:
env GOOS=darwin GOARCH=amd64 go build .

Linux:
env GOOS=linux GOARCH=amd64 go build .



```
