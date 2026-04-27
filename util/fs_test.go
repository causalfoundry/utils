package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindRootPathFrom(t *testing.T) {
	t.Run("find matching segment", func(t *testing.T) {
		ret, err := findRootPathFrom("/Users/dexian/causalfoundry/utils", "causalfoundry")
		assert.Nil(t, err)
		assert.Equal(t, "/Users/dexian/causalfoundry", ret)
	})

	t.Run("return error when missing", func(t *testing.T) {
		_, err := findRootPathFrom("/Users/dexian/causalfoundry/utils", "missing")
		assert.ErrorContains(t, err, `root path "missing" not found`)
	})

	t.Run("return error on empty root path", func(t *testing.T) {
		_, err := findRootPathFrom("/Users/dexian/causalfoundry/utils", "")
		assert.ErrorContains(t, err, "root path cannot be empty")
	})
}
