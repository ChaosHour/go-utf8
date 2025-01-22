package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
)

// Declare the database connection globally so it can be used by all functions
//var db *sql.DB

// Define flags
var (
	source    = flag.String("s", "", "Source Host")
	database  = flag.String("d", "", "Database Name")
	table     = flag.String("t", "", "Select table")
	show      = flag.Bool("show", false, "Show Databases")
	scanTable = flag.Bool("scan", false, "Scan specific table")
)

// read the ~/.my.cnf file to get the database credentials
func readMyCnf() {
	file, err := os.ReadFile(os.Getenv("HOME") + "/.my.cnf")
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "user") {
			os.Setenv("MYSQL_USER", strings.TrimSpace(line[5:]))
		}
		if strings.HasPrefix(line, "password") {
			os.Setenv("MYSQL_PASSWORD", strings.TrimSpace(line[9:]))
		}
	}
}

func connectToDatabase(source string) (*sql.DB, error) {
	db, err := sql.Open("mysql", os.Getenv("MYSQL_USER")+":"+os.Getenv("MYSQL_PASSWORD")+"@tcp("+source+":3306)/"+*database)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Get the hostname of the connected MySQL server
	var hostname string
	err = db.QueryRow("SELECT @@hostname").Scan(&hostname)
	if err != nil {
		return nil, err
	}

	// Print the result
	if err == nil {
		fmt.Printf("Connected to %s (%s): %s\n", source, hostname, color.GreenString("✔"))
		fmt.Println()
	} else {
		fmt.Printf("Failed to connect to %s (%s): %s\n", source, hostname, color.RedString("✘"))
		fmt.Println()
	}

	return db, nil
}

// Create a function that will get the Character Set and Collation for the table given with the -t table flag and the database with the -d database flag
func getTableCollationCharacterSet(db *sql.DB, database string, table string) (string, string, error) {
	var characterSet string
	var collation string
	err := db.QueryRow(fmt.Sprintf("SELECT CCSA.character_set_name, T.table_collation FROM information_schema.`TABLES` T LEFT JOIN information_schema.`COLLATION_CHARACTER_SET_APPLICABILITY` CCSA ON (T.table_collation = CCSA.collation_name) WHERE T.table_schema = '%s' AND T.table_name = '%s'", database, table)).Scan(&characterSet, &collation)
	if err != nil {
		return "", "", err
	}
	return characterSet, collation, nil
}

func isUnusualLatin1(value string) bool {
	for _, r := range value {
		if r > 255 || (r >= 128 && r <= 159) {
			return true
		}
	}
	return false
}

func isUnusualCP1252(value string) bool {
	problematicChars := map[rune]bool{129: true, 141: true, 143: true, 144: true, 157: true}
	for _, r := range value {
		if r > 255 || problematicChars[r] {
			return true
		}
	}
	return false
}

func getPrimaryKeys(db *sql.DB, database, table string) ([]string, error) {
	query := `SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
             WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND CONSTRAINT_NAME = 'PRIMARY'`
	rows, err := db.Query(query, database, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func scanTableForIssues(db *sql.DB, database, table string) error {
	startTime := time.Now()
	batchSize := 80000
	maxRetries := 5
	offendingIDs := make(map[string][]int) // Map to store column name -> slice of IDs

	// Get primary keys
	primaryKeys, err := getPrimaryKeys(db, database, table)
	if err != nil {
		return err
	}
	if len(primaryKeys) == 0 {
		return fmt.Errorf("table %s has no primary key", table)
	}

	// Get the min and max IDs
	primaryKeyStr := "`" + strings.Join(primaryKeys, "`, `") + "`"
	var minID, maxID int
	query := fmt.Sprintf("SELECT MIN(%s), MAX(%s) FROM %s.%s", primaryKeyStr, primaryKeyStr, database, table)
	err = db.QueryRow(query).Scan(&minID, &maxID)
	if err != nil {
		return err
	}

	// Get text columns
	columnsQuery := `SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS 
                    WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? 
                    AND DATA_TYPE IN ('char', 'varchar', 'text', 'tinytext', 'mediumtext', 'longtext')`

	columns, err := db.Query(columnsQuery, database, table)
	if err != nil {
		return err
	}
	defer columns.Close()

	var columnNames []string
	for columns.Next() {
		var colName string
		if err := columns.Scan(&colName); err != nil {
			return err
		}
		columnNames = append(columnNames, colName)
	}

	// Process in batches
	for start := minID; start <= maxID; start += batchSize {
		end := start + batchSize - 1
		if end > maxID {
			end = maxID
		}

		for _, column := range columnNames {
			for retry := 0; retry < maxRetries; retry++ {
				query := fmt.Sprintf(`SELECT %s, %s FROM %s.%s WHERE %s BETWEEN ? AND ?`,
					primaryKeyStr, column, database, table, primaryKeyStr)

				rows, err := db.Query(query, start, end)
				if err != nil {
					if retry == maxRetries-1 {
						return err
					}
					time.Sleep(time.Second * time.Duration(retry+1))
					continue
				}
				defer rows.Close()

				for rows.Next() {
					var id int
					var value sql.NullString
					if err := rows.Scan(&id, &value); err != nil {
						return err
					}

					if value.Valid {
						if isUnusualLatin1(value.String) || isUnusualCP1252(value.String) {
							offendingIDs[column] = append(offendingIDs[column], id)
						}
					}
				}
				break
			}
		}
	}

	// Print results for each column with issues
	for column, ids := range offendingIDs {
		if len(ids) > 0 {
			fmt.Printf("\nCurrent table: %s\n", color.BlueString(table))
			fmt.Printf("Column: %s\n", color.BlueString(column))
			fmt.Printf("Count of records that need to be fixed: %s\n\n", color.RedString(fmt.Sprintf("%d", len(ids))))
			fmt.Println("Offending IDs:")
			fmt.Printf("%v\n\n", ids)
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("Time taken: %v seconds\n", elapsed.Seconds())
	return nil
}

func main() {
	// Parse the command line flags
	flag.Parse()

	// Check if the source flag is set
	if *source == "" {
		fmt.Println("Usage: go-utf8 -s <source host> [-d <database name>] [-show] [-t <table>] [-scan]")
		fmt.Println("Please specify a source host")
		os.Exit(1)
	}

	// read the ~/.my.cnf file to get the database credentials
	readMyCnf()

	// Connect to MySQL database using the credentials from the ~/.my.cnf file and the function above connectToDatabase
	db, err := connectToDatabase(*source)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Show databases if the -show flag is set
	if *show {
		rows, err := db.Query("SHOW DATABASES")
		if err != nil {
			panic(err.Error())
		}
		defer rows.Close()

		for rows.Next() {
			var dbName string
			err := rows.Scan(&dbName)
			if err != nil {
				panic(err.Error())
			}
			fmt.Println(dbName)
		}
		return
	}

	// Check if the database flag is set
	if *database == "" {
		fmt.Println("Usage: go-utf8 -s <source host> -d <database name> [-t <table>] [-scan]")
		fmt.Println("Please specify a database name")
		os.Exit(1)
	}

	// Handle table info request (existing behavior)
	if *database != "" && *table != "" && !*scanTable {
		characterSet, collation, err := getTableCollationCharacterSet(db, *database, *table)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("Default character set:", color.RedString(characterSet))
		fmt.Println("Default collation:", color.RedString(collation))
		fmt.Println()
		return
	}

	// Handle specific table scanning
	if *database != "" && *table != "" && *scanTable {
		if err := scanTableForIssues(db, *database, *table); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Get a list of tables to loop through
	rows, err := db.Query(fmt.Sprintf("SHOW TABLES FROM %s", *database))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Loop through the tables
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			log.Fatal(err)
		}

		// Get a list of columns to loop through
		columns, err := db.Query(fmt.Sprintf("SELECT column_name FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name = '%s' AND table_schema = '%s' ORDER BY ordinal_position", tableName, *database))
		if err != nil {
			log.Fatal(err)
		}
		defer columns.Close()

		// Loop through the columns
		for columns.Next() {
			var columnName string
			err := columns.Scan(&columnName)
			if err != nil {
				log.Fatal(err)
			}

			// Query the column for records that need to be fixed
			query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s` WHERE LENGTH(`%s`) != CHAR_LENGTH(`%s`) AND `%s` IS NOT NULL", *database, tableName, columnName, columnName, columnName)
			var count int
			err = db.QueryRow(query).Scan(&count)
			if err != nil {
				log.Fatal(err)
			}

			// Print the result
			if count > 0 {
				fmt.Println()
				fmt.Printf("Current table: %s\n", color.BlueString(tableName))
				fmt.Printf("Column: %s\n", color.BlueString(columnName))
				fmt.Printf("Count of records that need to be fixed: %s\n", color.RedString(fmt.Sprintf("%d", count)))
				fmt.Println()
			}
		}

		// Get the total number of rows in the table
		var totalRows int
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", *database, tableName)).Scan(&totalRows)
		if err != nil {
			log.Fatal(err)
		}

		/*
			// Create a progress bar
			bar := progressbar.NewOptions(int(totalRows),
				progressbar.OptionSetRenderBlankState(true),
				progressbar.OptionSetWidth(15),
				progressbar.OptionSetDescription(*database),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer:        "=",
					SaucerPadding: " ",
					BarStart:      "[",
					BarEnd:        "]",
				}),
			)
		*/
		// Loop through the columns
		for columns.Next() {
			var columnName string
			err := columns.Scan(&columnName)
			if err != nil {
				log.Fatal(err)
			}

			// Loop through the columns and check if they contain utf8 encoded data
			rows, err := db.Query(fmt.Sprintf("SELECT `%s` FROM `%s`.`%s`", columnName, *database, tableName))
			if err != nil {
				panic(err.Error())
			}
			defer rows.Close()

			for rows.Next() {
				var value sql.NullString
				err := rows.Scan(&value)
				if err != nil {
					panic(err.Error())
				}

				// Check if the value contains non-utf8 characters
				if value.Valid {
					if !utf8.ValidString(value.String) {
						fmt.Println()
						fmt.Printf("Non-UTF8 character found in table: %s, column: %s, value: %s\n", color.BlueString(tableName), color.BlueString(columnName), color.RedString(value.String))
						fmt.Println()
					}
				}

				// Increment the progress bar
				//bar.Add(1)
			}

			// Increment the progress bar
			//bar.Add(1)
			//fmt.Println()
		}

		/*
			// Query the information schema to get the tables and their character set and collation
			tables, err := db.Query(fmt.Sprintf("SELECT table_name, TABLE_COLLATION FROM information_schema.tables WHERE table_schema = '%s'", *database))
			if err != nil {
				panic(err.Error())
			}
			defer tables.Close()
		*/

		/*
			// Print the default character set and collation for each table
			for tables.Next() {
				var tableName string
				var tableCollation string
				err := tables.Scan(&tableName, &tableCollation)
				if err != nil {
					panic(err.Error())
				}
				fmt.Println()
				fmt.Printf("Table: %s, Collation: %s\n", color.BlueString(tableName), color.RedString(tableCollation))
			}
		*/

		/*
			// Query the information schema to get the default character set for the table
			var characterSet string
			err = db.QueryRow(fmt.Sprintf("SELECT CCSA.character_set_name FROM information_schema.`TABLES` T LEFT JOIN information_schema.`COLLATION_CHARACTER_SET_APPLICABILITY` CCSA ON (T.table_collation = CCSA.collation_name) WHERE T.table_schema = '%s' AND T.table_name = '%s'", *database, tableName)).Scan(&characterSet)
			if err != nil {
				panic(err.Error())
			}
			fmt.Println("Default character set:", color.RedString(characterSet))
			fmt.Println()
			// end
		*/

		if *database != "" && *table != "" {
			// call the function getTableCollationCharacterSet to get the character set and collation for the table
			characterSet, collation, err := getTableCollationCharacterSet(db, *database, *table)
			if err != nil {
				panic(err.Error())
			}
			fmt.Println("Default character set:", color.RedString(characterSet))
			fmt.Println("Default collation:", color.RedString(collation))
			fmt.Println()
			os.Exit(0) // Exit the program after printing the character set and collation

		}

	}
}
