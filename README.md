# Go-UTF8

## What it does
Checks if utf8 encoded data is being stored in  latin1 columns.




## ~/my.cnf Used for testing
````ini

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
Usage: go-utf8 -s <source host> [-d <database name>] [-show]
Please specify a source host



go-utf8 on  main [!?] via 🐹 v1.21.0 
❯ go-utf8 -h
Usage of go-utf8:
  -d string
        Database Name
  -e    Show encoding comparison
  -s string
        Source Host
  -show
        Show Databases



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
Column: description
Count of records that need to be fixed: 1

char_test_db   0% [               ]  [0s:0s]
Table: t1, Collation: utf8mb3_general_ci
Default character set: utf8mb3
```




## Screenshots

<img src="screenshots/Screenshot 2023-09-04 at 9.21.40 PM.png" width="1053" height="473" />




## reference

- https://oracle-base.com/articles/mysql/mysql-converting-table-character-sets-from-latin1-to-utf8#the-problem
- https://www.percona.com/blog/utf8-data-on-latin1-tables-converting-to-utf8-without-downtime-or-double-encoding/
- https://mysql.rjweb.org/doc.php/charcoll#diagnosing_charset_issues
- https://www.percona.com/blog/fixing-column-encoding-mess-in-mysql/


## How to install
```Go

go install github.com/ChaosHour/go-utf8@latest


To build:

go build -o go-gtids

FreeBSD:
env GOOS=freebsd GOARCH=amd64 go build .

On Mac:
env GOOS=darwin GOARCH=amd64 go build .

Linux:
env GOOS=linux GOARCH=amd64 go build .



```
