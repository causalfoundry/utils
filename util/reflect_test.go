package util

import (
	"fmt"
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
	Address   string      `json:"-"`
}

func TestStructReflect(t *testing.T) {
	ti := time.Now()
	a := structA{"john", 3, ti, []int{}, []time.Time{}, "xx"}

	tags := StructTags(a, "json", nil)
	assert.Equal(t, tags, []string{"name", "age", "time", "int_array", "time_array"})

	tags, vals := ExtractTags(a, "json", nil)
	fmt.Println(tags, vals)

	dtypes := StructDataType(a, "")
	assert.Equal(t, dtypes, []string{"string", "int", "time", "_int", "_time", "string"})

	dtypes = StructDataType(a, "json")
	assert.Equal(t, dtypes, []string{"string", "int", "time", "_int", "_time"})
}
