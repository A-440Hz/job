package db

import (
	"os"

	"gorm.io/gorm"
)

func SetEnvForTesting() {
	os.Setenv(sqlDSNTesting, "postgres://my-postgres:my_password@localhost:5432/postgres")
	os.Setenv(gormDSNTesting, "host=localhost user=postgres password=my_password dbname=postgres port=5432 sslmode=disable")
	os.Setenv(gormDSNLocalhost, "host=localhost user=postgres2 password=my_password2 dbname=my_db port=5433 sslmode=disable")
}

func CleanDB(b gorm.DB, structs ...any) {

	b.AutoMigrate(structs...)
	// TRUNCATE - efficient delete
	// RESTART IDENTITY - Automatically restart sequences owned by columns of the truncated table(s)
	// CASCADE - Automatically truncate all tables that have foreign-key references to any of the named tables, or to any tables added to the group due to CASCADE
	b.Exec("TRUNCATE TABLE users, job_app_trackers, job_app_items, user_inventories RESTART IDENTITY CASCADE")

}
