package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// these vars coorespond to docker-compose and internal/db/testing.go to connect to the db
// railway has its own "DATABASE_URL" so I need to remember not to mess it up when deploying
const sqlDSNTesting = "DATABASE_URL_SQL_TESTING"
const gormDSNTesting = "DATABASE_URL_GORM_TESTING"
const gormDSNLocalhost = "DATABASE_URL_GORM_LOCALHOST"
const gormDSNRailway = "DATABASE_URL"

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
func InitGormTestDB() (*gorm.DB, error) {
	return initGormDB(gormDSNTesting)

}

func InitGormLocalDB() (*gorm.DB, error) {
	return initGormDB(gormDSNLocalhost)
}

func InitGormRailwayDB() (*gorm.DB, error) {
	return initGormDB(gormDSNRailway)
}

func initGormDB(dbURL string) (*gorm.DB, error) {
	dsn, err := getDSN(dbURL)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to open DB connection with %q: %w", dbURL, err)
	}

	// Configure connection pool to handle idle connections
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("unable to get underlying sql.DB: %w", err)
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
	// This ensures connections are recycled before they go stale (Railway sleeps after ~5min)
	sqlDB.SetConnMaxLifetime(time.Minute * 3)

	// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle
	// Close idle connections before Railway puts DB to sleep
	sqlDB.SetConnMaxIdleTime(time.Minute * 3)

	return db, nil
}
