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

// these vars coorespond to docker-compose and internal/db/testing.go to connect to the db
// firebase has its own "DATABASE_URL" so I need to remember not to mess it up when deploying
const sqlDSNTesting = "DATABASE_URL_SQL_TESTING"
const gormDSNTesting = "DATABASE_URL_GORM_TESTING"

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
	dsn, err := getDSN(sqlDSNTesting)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to open DB connection with %q: %w", sqlDSNTesting, err)
	}
	return db, nil
}

// Why gorm?
// syncs structs with db schema
// connections pooling and can group multiple db operations into single atomic operation
// manages migrations so I don't have to
// summarize and rejustify this later
func InitGormDB() (*gorm.DB, error) {
	dsn, err := getDSN(gormDSNTesting)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to open DB connection with %q: %w", gormDSNTesting, err)
	}

	return db, nil
}
