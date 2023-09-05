package encoding

import (
	"database/sql"
	"fmt"
	"log"
)

func compareEncodings(db *sql.DB, columnName string, tableName string, database string) {
	query := fmt.Sprintf("SELECT CONVERT(CONVERT(`%s` USING BINARY) USING latin1) AS latin1, CONVERT(CONVERT(`%s` USING BINARY) USING utf8) AS utf8 FROM `%s`.`%s` WHERE CONVERT(`%s` USING BINARY) RLIKE CONCAT('[', UNHEX('80'), '-', UNHEX('FF'), ']+')", columnName, columnName, database, tableName, columnName)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("+-------------------------+----------------+")
	fmt.Println("| latin1                  | utf8           |")
	fmt.Println("+-------------------------+----------------+")
	for rows.Next() {
		var latin1 string
		var utf8 string
		err := rows.Scan(&latin1, &utf8)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("| %-23s | %-14s |\n", latin1, utf8)
	}
	fmt.Println("+-------------------------+----------------+")
}
