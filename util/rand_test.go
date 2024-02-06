package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomAlphabets(t *testing.T) {
	ret := RandomAlphabets(10, false)
	assert.Len(t, ret, 10)
}
