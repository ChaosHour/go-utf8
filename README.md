# Go-UTF8

## What it does

This Go CLI checks if any columns in a specified MySQL database have latin1 character set instead of utf8 and if they contain utf8 encoded data. It connects to the MySQL database using the credentials from the `~/.my.cnf` file and the command line flags. It then loops through all the tables in the specified database and checks each column for non-utf8 encoded data. If any columns are found to have non-utf8 encoded data, it prints the table name, column name, and the non-utf8 character value in red color. 

The CLI also creates a progress bar to show the progress of checking each column for non-utf8 encoded data. 

## How to use it
To use this CLI, you need to run it with the appropriate command line flags. You need to specify the source host using the `-s` flag and the database name using the `-d` flag. For example, to check the `mydatabase` database on the `localhost` server, you would run:

```
./go-utf8 -s localhost -d mydatabase
```

This will output any columns that have non-utf8 encoded data. You can then fix the data by converting the column character set to utf8 using the `ALTER TABLE` statement.


