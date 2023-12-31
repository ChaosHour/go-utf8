package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
)

// Declare the database connection globally so it can be used by all functions
//var db *sql.DB

// Define flags
var (
	source   = flag.String("s", "", "Source Host")
	database = flag.String("d", "", "Database Name")
	//showEncoding = flag.Bool("e", false, "Show encoding comparison")
	table = flag.String("t", "", "Select table")
	show  = flag.Bool("show", false, "Show Databases")
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

func main() {
	// Parse the command line flags
	flag.Parse()

	// Check if the source flag is set
	if *source == "" {
		fmt.Println("Usage: go-utf8 -s <source host> [-d <database name>] [-show] [-t <table>]")
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
		fmt.Println("Usage: go-utf8 -s <source host> -d <database name>")
		fmt.Println("Please specify a database name")
		os.Exit(1)
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
			query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s` WHERE LENGTH(`%s`) != CHAR_LENGTH(`%s`)", *database, tableName, columnName, columnName)
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
				var value string
				err := rows.Scan(&value)
				if err != nil {
					panic(err.Error())
				}

				// Check if the value contains non-utf8 characters
				if !utf8.ValidString(value) {
					fmt.Println()
					fmt.Printf("Non-UTF8 character found in table: %s, column: %s, value: %s\n", color.BlueString(tableName), color.BlueString(columnName), color.RedString(value))
					fmt.Println()
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
