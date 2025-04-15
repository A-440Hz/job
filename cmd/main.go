package main

/*
   The cmd directory handles entry points for the application.

   package main means this file will be compiled into a runnable executable
   to interface with other frameworks.
*/

import (
	"fmt"
	"job/internal/db"
)

func main() {
	// connect to PostgreSQL
	db, err := db.Connect()
	defer db.Close()
	fmt.Println(db, err)
}
