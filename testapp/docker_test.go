package testapp

import (
	"causalfoundry/utils/iokit"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupLocalStorage(t *testing.T) {
	kit := iokit.NewKit()

	r, err := kit.DB.Exec("INSERT INTO test (id, v) VALUES (1,1)")
	assert.Nil(t, err)
	affect, err := r.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, affect, int64(1))
}
