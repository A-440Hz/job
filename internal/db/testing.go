package db

import "os"

func SetEnvForTesting() {
	os.Setenv(sqlDSN, "postgres://my-postgres:my_password@localhost:5432/postgres")
	os.Setenv(gormDSN, "host=localhost user=postgres password=my_password dbname=postgres port=5432 sslmode=disable")

}
