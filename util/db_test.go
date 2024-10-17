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
	assert.Equal(t, []int{1, 2, 3, 4}, ret)

	floatArray := "{1,2,3,4}"
	floatArr, err := ParseDBArray[float64](floatArray)
	assert.Nil(t, err)
	assert.Equal(t, []float64{1, 2, 3, 4}, floatArr)

	timeArray := "{2020-01-01, 2020-01-01T20:00:00Z}"
	timeArr, err := ParseDBArray[time.Time](timeArray)
	assert.Nil(t, err)
	assert.Equal(t, []time.Time{DateUTC(2020, 1, 1), time.Date(2020, 1, 1, 20, 0, 0, 0, time.UTC)}, timeArr)

	boolArray := "{t,t,t,f}"
	boolArr, err := ParseDBArray[bool](boolArray)
	assert.Nil(t, err)
	assert.Equal(t, []bool{true, true, true, false}, boolArr)
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

func TestPrepareIDPartition(t *testing.T) {
	db := NewTestDB("")
	_, err := db.Exec("CREATE SCHEMA common")
	assert.Nil(t, err)
	_, err = db.Exec("CREATE TABLE common.a (id serial not null, val int) PARTITION BY RANGE (id)")
	assert.Nil(t, err)

	_, err = db.Exec("CREATE SCHEMA partition")
	assert.Nil(t, err)
	assert.Nil(t, PrepareIDPartition(db, "common.a", "partition", 10))
	for i := 10; i < 30; i++ {
		assert.Nil(t, PrepareIDPartition(db, "common.a", "partition", 10))
		_, err = db.Exec("INSERT INTO common.a (val) VALUES ($1)", i)
		assert.Nil(t, err)
	}

}
