package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const dbEnvKey = "DATABASE_URL"

func Connect() (*sql.DB, error) {
	dockerStr, ok := os.LookupEnv(dbEnvKey)
	if !ok {
		log.Printf("%q not found; attempting to load env vars", dbEnvKey)
		err := godotenv.Load()
		if err != nil {
			err = fmt.Errorf("Error loading .env file in db.go init(): %w", err)
			log.Print(err)
			return nil, err
		}
		dockerStr, ok = os.LookupEnv(dbEnvKey)
		if !ok {
			err = fmt.Errorf("Error loading %q after reading .env file: %w", dbEnvKey, err)
			log.Print(err)
			return nil, err
		}
	}

	db, err := sql.Open("postgres", dockerStr)
	if err != nil {
		log.Println("Db connection failed:", err)
		return nil, err
	}
	return db, nil
}
