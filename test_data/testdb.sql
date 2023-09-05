--  https://oracle-base.com/articles/mysql/mysql-converting-table-character-sets-from-latin1-to-utf8#the-problem

CREATE DATABASE IF NOT EXISTS char_test_db;

select sleep(2);

\u char_test_db

DROP TABLE IF EXISTS t1;

CREATE TABLE t1 (
  id          INT(11) NOT NULL AUTO_INCREMENT,
  description VARCHAR(50),
  PRIMARY KEY(id)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=latin1;

SET NAMES latin1;

INSERT INTO t1 (description) VALUES ('¡Volcán!');