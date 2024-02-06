package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMultipleJsonbSet(t *testing.T) {
	ret := MultipleJsonbSet("metric",
		map[string]any{
			"{a,b,c}": 1,
			"{a,b,d}": 2,
			"{a,b,e}": 3,
		})
	fmt.Println(ret)
}

func TestValueList(t *testing.T) {
	l := ValueList([]time.Time{DateUTC(2020, 1, 1), DateUTC(2020, 1, 2)}, "t", false)
	fmt.Println(l)
	assert.Equal(t, l, "(VALUES ('2020-01-01T00:00:00Z'),('2020-01-02T00:00:00Z')) a(t)")
}

func TestDBArrayValues(t *testing.T) {
	fmt.Println(DBArrayValues([]any{1, 2, 3, time.Now()}))

	assert.Equal(t,
		DBArrayValues([]int{1, 2, 3, 4, 5}),
		"{1,2,3,4,5}",
	)

	t1 := DateUTC(2020, 1, 1)
	assert.Equal(t,
		DBArrayValues([]time.Time{t1, t1}),
		"{2020-01-01T00:00:00Z,2020-01-01T00:00:00Z}",
	)
}

func TestAllTokens(t *testing.T) {
	arr := "{1,2,3, 4, 45 ,55}"
	tokens := allTokens(arr)
	assert.Equal(t, tokens, []string{"1", "2", "3", "4", "45", "55"})
}

func TestParseDBArray(t *testing.T) {
	intArray := "{1,2,3,4}"
	ret, err := ParseDBArray[int](intArray)
	assert.Nil(t, err)
	assert.Equal(t, ret, []int{1, 2, 3, 4})

	floatArray := "{1,2,3,4}"
	floatArr, err := ParseDBArray[float64](floatArray)
	assert.Nil(t, err)
	assert.Equal(t, floatArr, []float64{1, 2, 3, 4})

	timeArray := "{2020-01-01, 2020-01-01T20:00:00Z}"
	timeArr, err := ParseDBArray[time.Time](timeArray)
	assert.Nil(t, err)
	assert.Equal(t, timeArr, []time.Time{DateUTC(2020, 1, 1), time.Date(2020, 1, 1, 20, 0, 0, 0, time.UTC)})
}

func TestTimePointsToFiller(t *testing.T) {
	db := NewTestDB("")
	type result struct {
		ID   string    `db:"id"`
		Time time.Time `db:"time"`
	}
	query := TimePointsFiller(
		[]string{"a", "b"},
		[]time.Time{DateUTC(2020, 1, 1), DateUTC(2020, 1, 2)},
	)
	var results []result
	err := db.Select(&results, query)
	assert.Nil(t, err)
	assert.Len(t, results, 4)
}
