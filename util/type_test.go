package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullFloat64(t *testing.T) {
	db := NewTestPostgresDB("")

	var f NullFloat64
	err := db.QueryRow("SELECT null").Scan(&f)
	assert.Nil(t, err)
	assert.True(t, f.Null)
	assert.Zero(t, f.V)

	err = db.QueryRow("SELECT 1").Scan(&f)
	assert.Nil(t, err)
	assert.False(t, f.Null)
	assert.Equal(t, f.V, 1.)

	b, _ := json.Marshal(f)
	assert.Equal(t, string(b), "1")
}
