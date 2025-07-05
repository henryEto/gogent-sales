package database

import (
	"database/sql"
	"log"
)

func connectDB() *sql.DB {
	// Replace "username", "password", "dbname" with your database credentials
	connectionString := "username:password@tcp(localhost:3306)/dbname"
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
