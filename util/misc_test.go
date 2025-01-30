package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamelToSnakeBetter(t *testing.T) {
	group, idx := findNextGroup("ParqTS")
	assert.Equal(t, group, "Parq")
	assert.Equal(t, idx, 4)

	group, idx = findNextGroup("PID")
	assert.Equal(t, group, "PID")
	assert.Equal(t, idx, 3)

	group, idx = findNextGroup("hello")
	assert.Equal(t, group, "hello")
	assert.Equal(t, idx, 5)

	group, idx = findNextGroup("helloTS")
	assert.Equal(t, group, "hello")
	assert.Equal(t, idx, 5)

	ret := CamelToSnake2("ParqTS")
	assert.Equal(t, ret, "parq_ts")

	ret = CamelToSnake2("hello")
	assert.Equal(t, ret, "hello")

	ret = CamelToSnake2("abcDEF")
	assert.Equal(t, ret, "abc_def")

	ret = CamelToSnake2("ABC")
	assert.Equal(t, ret, "abc")

	ret = CamelToSnake2("ABCd")
	assert.Equal(t, ret, "abcd")

	ret = CamelToSnake2("PrescribedTestsList")
	assert.Equal(t, ret, "prescribed_tests_list")
}
