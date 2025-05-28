package db

import "os"

func SetEnvForTesting() {
	os.Setenv(sqlDSNTesting, "postgres://my-postgres:my_password@localhost:5432/postgres")
	os.Setenv(gormDSNTesting, "host=localhost user=postgres password=my_password dbname=postgres port=5432 sslmode=disable")
	os.Setenv(gormDSNLocalhost, "host=localhost user=postgres2 password=my_password2 dbname=my_db port=5433 sslmode=disable")
}
