package testapp

import (
	"os"
	"strings"
	"testing"

	"github.com/causalfoundry/utils/util"
	"github.com/stretchr/testify/assert"
)

func url(dbName string) string {
	baseUrl := "host=localhost port=5432 dbname=<name> user=user password=pwd sslmode=disable"
	return strings.ReplaceAll(baseUrl, "<name>", dbName)
}

func TestSetupLocalStorage(t *testing.T) {
	dbName := util.RandomAlphabets(10, true)
	path, err := os.Getwd()
	util.Panic(err)
	util.SetupLocalStorage(dbName, "postgres", url("postgres"), path+"/migrations/postgres")

	db := util.NewDB(dbName, url(dbName), util.DBConfig{})
	r, err := db.Exec("INSERT INTO test (id, v) VALUES (1,1)")
	assert.Nil(t, err)
	affect, err := r.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, affect, int64(1))
}
