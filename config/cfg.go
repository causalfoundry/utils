package config

import (
	"encoding/base64"
	"fmt"
	"github.com/causalfoundry/utils/util"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Mode          Mode
	AppRoot       string         `yaml:"app_root"`
	Postgres      PostgresConfig `yaml:"postgres_config"`
	Redis         RedisConfig    `yaml:"redis_config"`
	GoogleAuth    GoogleAuth     `yaml:"google_auth"`
	LocalAuthMode bool           `yaml:"local_auth_mode"`
}

type GoogleAuth struct {
	ClientID string `yaml:"client_id" validate:"required"`
}

type PostgresConfig struct {
	Username        string `yaml:"username" validate:"required"`
	Password        string `yaml:"password" validate:"required"`
	DatabaseName    string `yaml:"database_name" validate:"required"`
	Host            string `yaml:"host" validate:"required"`
	DatabaseSSLMode string `yaml:"database_ssl_mode" validate:"required"`
	Port            int    `yaml:"port" validate:"required"`
}

func (p PostgresConfig) GetLocalBaseURL() string {
	return fmt.Sprintf("host=%s port=%d dbname=postgres user=%s password=%s sslmode=%s",
		p.Host, p.Port, p.Username, p.Password, p.DatabaseSSLMode)
}

func (p PostgresConfig) GetURL() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		p.Host, p.Port, p.DatabaseName, p.Username, p.Password, p.DatabaseSSLMode)
}

type RedisConfig struct {
	Master       string `yaml:"master"`
	Slave        string `yaml:"slave"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	GlobalPrefix string `yaml:"global_prefix"`
}

type Mode string

const (
	LocalHost   Mode = "local_host"   // running as binary on host
	LocalDocker Mode = "local_docker" // running as docker container
	Cloud       Mode = "cloud"
)

func (m Mode) IsLocal() bool {
	return strings.HasPrefix(string(m), "local_")
}

func setupLocalConfigEnv(uniqueAppFolder string) {
	path := util.AppRootPath(uniqueAppFolder)
	b, err := os.ReadFile(path + "/config.yml")
	util.Panic(err)

	switch os.Getenv("MODE") {
	case "", "local_host", "local_docker":
		os.Setenv("CONFIG", string(b))
	case "cloud":
		decodedData, err := base64.StdEncoding.DecodeString(string(b))
		util.Panic(fmt.Errorf("error decode base64 of config, %w, %s", err, string(b)))
		os.Setenv("CONFIG", string(decodedData))
	}
}

func prepareConfig(appFolderName string) (ret Config) {
	switch value, _ := os.LookupEnv("MODE"); value {
	case "", "local_host":
		ret.Mode = LocalHost
		setupLocalConfigEnv(appFolderName)
	case "local_docker":
		ret.Mode = LocalDocker
		setupLocalConfigEnv(appFolderName)
	case "cloud":
		ret.Mode = Cloud
	}

	cfg := os.Getenv("CONFIG")
	if cfg == "" {
		panic("CONFIG not found")
	}
	if err := yaml.Unmarshal([]byte(cfg), &ret); err != nil {
		panic("error unmarshal config to Config type: " + err.Error())
	}
	util.Panic(util.Validator.Struct(ret))
	fmt.Println("-- successful config load --")
	return
}

// appFolderName is the unique folder that can identify the app root path
// we assume the config file (yaml) is located in this root
func NewConfig(appRoot string) (ret Config) {
	ret = prepareConfig(appRoot)

	switch ret.Mode {
	case LocalHost:
		ret.Postgres.Host = "localhost"
		ret.Redis.Master = "localhost"
		ret.Redis.GlobalPrefix = util.RandomAlphanumeric(10, false)
		ret.Postgres.DatabaseName = util.RandomAlphabets(10, true)

	case LocalDocker:
		ret.Postgres.DatabaseName = os.Getenv("LOCAL_DB_NAME")
		if ret.Postgres.DatabaseName == "" {
			panic("LOCAL_DB_NAME not set")
		}

	case Cloud:
		// nothing here
	default:
		panic(fmt.Sprintf("invalid mode: %s", ret.Mode))
	}
	fmt.Printf("=========\ndb_name: %s\n=======\n", ret.Postgres.DatabaseName)
	if appRoot != ret.AppRoot {
		panic("app root is not the same in config.yml")
	}
	return ret
}
