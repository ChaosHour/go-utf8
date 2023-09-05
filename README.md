# Go-UTF8

## What it does
Checks if utf8 encoded data is being stored in  latin1 columns.






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
