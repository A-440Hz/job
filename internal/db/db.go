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

func getDSN() (string, error) {
	dsn, ok := os.LookupEnv(dbEnvKey)
	if !ok {
		log.Printf("%q not found; attempting to load env vars", dbEnvKey)
		if err := godotenv.Load(); err != nil {
			return "", fmt.Errorf("failed to load .env file in db.go init(): %w", err)
		}

		dsn, ok = os.LookupEnv(dbEnvKey)
		if !ok {
			return "", fmt.Errorf("env var %q not found after reloading .env file", dbEnvKey)
		}
	}
	return dsn, nil
}

func Connect() (*sql.DB, error) {
	dsn, err := getDSN()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to open DB connection with %q: %w", dbEnvKey, err)
	}
	return db, nil
}

func InitGormDB() (*gorm.DB, error) {
	dsn, err := getDSN()
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to open DB connection with %q: %w", dbEnvKey, err)
	}
	return db, nil
}
