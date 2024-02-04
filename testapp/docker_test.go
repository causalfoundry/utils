package testapp

import (
	"os"
	"testing"

	"github.com/causalfoundry/utils/util"
	"github.com/stretchr/testify/assert"
)

func TestSetupLocalStorage(t *testing.T) {
	pwd, err := os.Getwd()
	assert.Nil(t, err)

	db := util.NewTestDB(pwd + "/migrations/postgres")
	r, err := db.Exec("INSERT INTO test (id, v) VALUES (1,1)")
	assert.Nil(t, err)
	affect, err := r.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, affect, int64(1))
}
