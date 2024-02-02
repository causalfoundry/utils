package docker

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/causalfoundry/utils/util"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func SetupLocalStorage(newDB, baseDB, baseUrl, migrationFile string) {
	db, err := sql.Open("pgx", baseUrl)
	if err != nil {
		panic(fmt.Sprintf("error get base database connection: %s", err.Error()))
	}

	var exist bool
	row := db.QueryRow("select exists (select 1 from pg_database where datname = $1)", newDB)
	err = row.Scan(&exist)
	if err != nil {
		panic(err.Error())
	}

	if !exist {
		_, err = db.Exec("CREATE DATABASE " + newDB)
		if err != nil {
			panic("error create local database: " + err.Error())
		}
		db.Close()

		newUrl := strings.ReplaceAll(baseUrl, baseDB, newDB)
		if db, err = sql.Open("pgx", newUrl); err != nil {
			panic(fmt.Sprintf("error get new database connection: %s", err.Error()))
		}

		driver, _ := postgres.WithInstance(db, &postgres.Config{})
		migrateInstance, err := migrate.NewWithDatabaseInstance(
			"file://"+migrationFile,
			newDB,
			driver,
		)
		util.Panic(err)

		err = migrateInstance.Up()
		if err != nil {
			util.Panic(err)
		}
	}
}
