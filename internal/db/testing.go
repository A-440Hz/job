package db

import (
	"os"
	"path"
	"runtime"

	"gorm.io/gorm"
)

func SetEnvForTesting() {
	os.Setenv(sqlDSNTesting, "postgres://my-postgres:my_password@localhost:5432/postgres")
	os.Setenv(gormDSNTesting, "host=localhost user=postgres password=my_password dbname=postgres port=5432 sslmode=disable")
	os.Setenv(gormDSNLocalhost, "host=localhost user=postgres2 password=my_password2 dbname=my_db port=5433 sslmode=disable")
}

func CleanDB(db gorm.DB, structs ...any) {
	db.AutoMigrate(structs...)
	// TRUNCATE - efficient delete
	// RESTART IDENTITY - Automatically restart sequences owned by columns of the truncated table(s)
	// CASCADE - Automatically truncate all tables that have foreign-key references to any of the named tables, or to any tables added to the group due to CASCADE
	db.Exec("TRUNCATE TABLE users, job_app_trackers, job_app_items, user_inventories, sessions RESTART IDENTITY CASCADE")
}

// https://intellij-support.jetbrains.com/hc/en-us/community/posts/360009685279-Go-test-working-directory-keeps-changing-to-dir-of-the-test-file-instead-of-value-in-template
// it's pretty cool that you get to do this in golang to get releative paths to work
func ChdirGoTests() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
