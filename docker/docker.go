package docker

import (
	"database/sql"
	"fmt"
	"github.com/causalfoundry/utils/config"
	"github.com/causalfoundry/utils/util"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func SetupLocalStorage(cfg config.Config) {
	db, err := sql.Open("pgx", cfg.Postgres.GetLocalBaseURL())
	if err != nil {
		panic(fmt.Sprintf("error get base database connection: %s", err.Error()))
	}

	var exist bool
	row := db.QueryRow("select exists (select 1 from pg_database where datname = $1)", cfg.Postgres.DatabaseName)
	err = row.Scan(&exist)
	if err != nil {
		panic(err.Error())
	}

	if !exist {
		_, err = db.Exec("CREATE DATABASE " + cfg.Postgres.DatabaseName)
		if err != nil {
			panic("error create local database: " + err.Error())
		}
		db.Close()

		if db, err = sql.Open("pgx", cfg.Postgres.GetURL()); err != nil {
			panic(fmt.Sprintf("error get new database connection: %s", err.Error()))
		}

		driver, _ := postgres.WithInstance(db, &postgres.Config{})
		appPath := util.AppRootPath(cfg.AppRoot)
		migrateInstance, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s/migrations/postgres", appPath),
			cfg.Postgres.DatabaseName,
			driver,
		)
		util.Panic(err)

		err = migrateInstance.Up()
		if err != nil {
			util.Panic(err)
		}
	}
}
