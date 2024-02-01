package iokit

import (
	"causalfoundry/utils/config"
	"causalfoundry/utils/dbutil"
	"causalfoundry/utils/docker"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
)

type Kit struct {
	Cfg config.Config
	DB  *sqlx.DB
	Rds *redis.Client
}

func NewKit() Kit {
	cfg := config.NewConfig("testapp")

	if cfg.Mode.IsLocal() {
		docker.SetupLocalStorage(cfg)
		fmt.Println("local storage setup done")
	}

	db := dbtuil.NewDB(cfg, dbtuil.DBConfig{})
	return Kit{
		Cfg: cfg,
		DB:  db,
		Rds: nil,
	}
}