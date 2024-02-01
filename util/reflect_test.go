package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type structA struct {
	Name      string      `json:"name"`
	Age       int         `json:"age"`
	Time      time.Time   `json:"time"`
	IntArray  []int       `json:"int_array"`
	TimeArray []time.Time `json:"time_array"`
}

func TestStructReflect(t *testing.T) {
	ti := time.Now()
	a := structA{"john", 3, ti, []int{}, []time.Time{}}

	tags := StructTags(a, "json", nil)
	assert.Equal(t, tags, []string{"name", "age", "time", "int_array", "time_array"})

	dtypes := StructDataType(a)
	assert.Equal(t, dtypes, []string{"string", "int", "time", "_int", "_time"})
}
