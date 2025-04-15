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

// set this in a .env file for local testing.
// DATABASE_URL=postgres://my-postgres:my_password@localhost:5432/postgres
// corresponds to
// docker run --name my-postgres -e POSTGRES_PASSWORD=my_password -d -p 5432:5432 postgres
const sqlDSN = "DATABASE_URL"
const gormDSN = "DATABASE_URL_GORM"

func getDSN(env string) (string, error) {
	dsn, ok := os.LookupEnv(env)
	if !ok {
		log.Printf("%q not found; attempting to load env vars", env)
		if err := godotenv.Load(); err != nil {
			return "", fmt.Errorf("failed to load .env file in db.go init(): %w", err)
		}

		dsn, ok = os.LookupEnv(env)
		if !ok {
			return "", fmt.Errorf("env var %q not found after reloading .env file", env)
		}
	}
	return dsn, nil
}

func Connect() (*sql.DB, error) {
	dsn, err := getDSN(sqlDSN)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to open DB connection with %q: %w", sqlDSN, err)
	}
	return db, nil
}

func InitGormDB() (*gorm.DB, error) {
	dsn, err := getDSN(gormDSN)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to open DB connection with %q: %w", gormDSN, err)
	}

	return db, nil
}
