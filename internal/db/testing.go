package db

import "os"

func SetEnvForTesting() {
	os.Setenv(sqlDSNTesting, "postgres://my-postgres:my_password@localhost:5432/postgres")
	os.Setenv(gormDSNTesting, "host=localhost user=postgres password=my_password dbname=postgres port=5432 sslmode=disable")
}
