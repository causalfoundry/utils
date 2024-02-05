package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMapInPlace(t *testing.T) {
	ints := []int{1, 2, 3, 4, 5}
	MapInPlace(ints, func(t int) int {
		return t * 2
	})
	expectedInts := []int{2, 4, 6, 8, 10}
	assert.Equal(t, expectedInts, ints)

	strs := []string{"a", "b", "c"}
	MapInPlace(strs, func(t string) string {
		return t + t
	})
	expectedStrs := []string{"aa", "bb", "cc"}
	assert.Equal(t, expectedStrs, strs)
}

func TestExcept(t *testing.T) {
	t.Run("", func(t *testing.T) {
		a := []string{"1", "2", "3", "4"}
		b := []string{"1", "2"}
		c := Except(a, b)
		assert.Equal(t, c, []string{"3", "4"})
	})

	t.Run("", func(t *testing.T) {
		a := []time.Time{DateUTC(2020, 1, 1), DateUTC(2020, 1, 2)}
		b := []time.Time{DateUTC(2020, 1, 2)}
		c := Except(a, b)
		assert.Equal(t, c, []time.Time{DateUTC(2020, 1, 1)})
	})
}

func TestMap(t *testing.T) {
	a := []int{1, 2, 3, 4}
	b := Map(a, func(e int) string {
		return fmt.Sprint(e, "a")
	})
	assert.Equal(t, b, []string{"1a", "2a", "3a", "4a"})
}

func TestUnique(t *testing.T) {
	a := []any{DateUTC(2020, 1, 1), DateUTC(2020, 1, 1)}
	assert.Equal(t, Unique(a), []any{DateUTC(2020, 1, 1)})
}

func TestUniqueOf(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ret := UniqueOf(a, func(a int) int {
		if a/2 == 0 {
			return 1
		}
		return 2
	})
	assert.Equal(t, ret, []int{1, 2})
}
