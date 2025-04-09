package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const dbEnvKey = "DATABASE_URL"

func Connect() (*sql.DB, error) {
	dockerStr, ok := os.LookupEnv(dbEnvKey)
	if !ok {
		log.Printf("%q not found; attempting to load env vars", dbEnvKey)
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("failed to load .env file in db.go init(): %w", err)
		}

		dockerStr, ok = os.LookupEnv(dbEnvKey)
		if !ok {
			return nil, fmt.Errorf("env var %q not found after reloading .env file", dbEnvKey)
		}
	}

	db, err := sql.Open("postgres", dockerStr)
	if err != nil {
		return nil, fmt.Errorf("unable to open DB connection with %q: %w", dbEnvKey, err)
	}
	return db, nil
}

func InitGormDB() (*gorm.DB, error) {
	dockerStr, ok := os.LookupEnv(dbEnvKey)
	if !ok {
		log.Printf("%q not found; attempting to load env vars", dbEnvKey)
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("failed to load .env file in db.go init(): %w", err)
		}

		dockerStr, ok = os.LookupEnv(dbEnvKey)
		if !ok {
			return nil, fmt.Errorf("env var %q not found after reloading .env file", dbEnvKey)
		}
	}

	db, err := gorm.Open(postgres.Open(dockerStr), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to open DB connection with %q: %w", dbEnvKey, err)
	}
	return db, nil
}
