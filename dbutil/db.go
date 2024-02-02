package dbutil

import (
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var Psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type DBConfig struct {
	MaxOpenConn     int
	MaxIdelConn     int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

func (d *DBConfig) SetDefault() {
	if d.MaxOpenConn == 0 {
		d.MaxOpenConn = 100
	}
	if d.MaxIdelConn == 0 {
		d.MaxIdelConn = 10
	}
	if d.ConnMaxIdleTime == 0 {
		d.ConnMaxIdleTime = time.Second * 60
	}
	if d.ConnMaxLifetime == 0 {
		d.ConnMaxLifetime = time.Second * 300
	}
}

func NewDB(dbName, url string, dcfg DBConfig) *sqlx.DB {
	db, err := sqlx.Open("pgx", url)
	if err != nil {
		panic("cannot get postgres connection: " + err.Error())
	}

	if err = db.Ping(); err != nil {
		panic("cannot ping postgres: " + err.Error())
	}

	db.SetMaxOpenConns(dcfg.MaxOpenConn)
	db.SetConnMaxIdleTime(dcfg.ConnMaxIdleTime)
	db.SetMaxIdleConns(dcfg.MaxIdelConn)
	db.SetConnMaxLifetime(dcfg.ConnMaxLifetime)

	_, err = db.Exec(fmt.Sprintf("ALTER DATABASE %s SET timezone TO 'UTC'", dbName))
	if err != nil {
		panic("Error in setting timezone: " + err.Error())
	}

	// Set the timezone for the current session
	_, err = db.Exec("SET TIMEZONE TO 'UTC'")
	if err != nil {
		panic("Error in setting session timezone: " + err.Error())
	}

	return db
}
