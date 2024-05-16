package util

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	db := NewTestDB("")
	_, err := db.Exec("CREATE TABLE test (a int, b int)")
	assert.Nil(t, err)

	type row struct {
		A int `db:"a"`
		B int `db:"b"`
	}

	_, err = GetM[row](zerolog.Logger{}, db, "test", map[string]any{"a": 1})
	assert.NotNil(t, err)
}
