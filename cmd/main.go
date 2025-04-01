package main

/*
   The cmd directory handles entry points for the application.

   package main means this file will be compiled into a runnable executable
   to interface with other frameworks.
*/

import (
	"database/sql"
	"log"
	"os"
)

func main() {
	// connect to Railway PostgreSQL
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	log.Print("Successfully connected to PostgreSQL")
}
