package util

import (
	"database/sql/driver"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestMutlilineMigrationForClickhouse(t *testing.T) {
	dir, _ := os.Getwd()
	db := NewTestClickhouseDB(dir + "/.." + "/migrations/clickhouse")
	assert.NotNil(t, db)
	assert.Nil(t, db.Ping())

	db = NewTestPostgresDB(dir + "/.." + "/migrations/postgres")
	assert.NotNil(t, db)
	assert.Nil(t, db.Ping())
}

func TestGetDB(t *testing.T) {
	db := NewTestClickhouseDB("")
	assert.NotNil(t, db)
	db = NewTestPostgresDB("")
	assert.NotNil(t, db)
}

func TestShouldSetPostgresTimezone(t *testing.T) {
	db, err := sqlx.Open("pgx", "postgres://user:pwd@localhost:5439/postgres?sslmode=disable")
	assert.Nil(t, err)
	assert.Equal(t, "pgx", db.DriverName())
	assert.True(t, shouldSetPostgresTimezone(db.DriverName()))
	assert.True(t, shouldSetPostgresTimezone("postgres"))
	assert.False(t, shouldSetPostgresTimezone("clickhouse"))
}

func TestDBDriverNameFromURL(t *testing.T) {
	assert.Equal(t, "pgx", dbDriverNameFromURL("postgres://user:pwd@localhost:5439/postgres?sslmode=disable"))
	assert.Equal(t, "clickhouse", dbDriverNameFromURL("clickhouse://user:pwd@localhost:9009/default?"))
	assert.Equal(t, "", dbDriverNameFromURL("mysql://user:pwd@localhost/db"))
}

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

func TestValueListArgs(t *testing.T) {
	query, args, err := ValueListArgs([]string{"a", "O'Brien"}, "name", false)
	assert.Nil(t, err)
	assert.Equal(t, "(VALUES (?),(?)) a(name)", query)
	assert.Equal(t, []any{"a", "O'Brien"}, args)

	query, args, err = ValueListTableArgs([]int{3, 4}, "id")
	assert.Nil(t, err)
	assert.Equal(t, "SELECT * FROM (VALUES (?),(?)) a(id) ", query)
	assert.Equal(t, []any{3, 4}, args)

	query, args, err = ValueListArgs([]string{"a", "b"}, "name", true)
	assert.Nil(t, err)
	assert.Equal(t, "(VALUES (?,?),(?,?)) a(name,sort_order)", query)
	assert.Equal(t, []any{"a", 0, "b", 1}, args)

	_, _, err = ValueListArgs([]string{}, "name", false)
	assert.NotNil(t, err)
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

func TestDBArrayArg(t *testing.T) {
	valuer, ok := DBArrayArg([]string{"a", "O'Brien", "x,y"}).(driver.Valuer)
	assert.True(t, ok)

	val, err := valuer.Value()
	assert.Nil(t, err)
	assert.Equal(t, `{"a","O'Brien","x,y"}`, val)
}

func TestEQxArgs(t *testing.T) {
	query, args := EQxArgs("name", "O'Brien")
	assert.Equal(t, "name = ?", query)
	assert.Equal(t, []any{"O'Brien"}, args)

	query, args = EQxArgs("id", []int{1, 2, 3})
	assert.Equal(t, "id IN (?,?,?)", query)
	assert.Equal(t, []any{1, 2, 3}, args)
}

func TestINxArgs(t *testing.T) {
	query, args := INxArgs("name", []string{"a", "O'Brien"})
	assert.Equal(t, "name IN (?,?)", query)
	assert.Equal(t, []any{"a", "O'Brien"}, args)

	query, args = INxArgs("id", []int{})
	assert.Equal(t, "1=0", query)
	assert.Nil(t, args)
}

func TestMultipleJsonbSetArgs(t *testing.T) {
	query, args, err := MultipleJsonbSetArgs("metric", map[string]any{
		"{a,b,c}": "O'Brien",
	})
	assert.Nil(t, err)
	assert.Equal(t, "JSONB_SET(metric, ?::text[], ?::jsonb)", query)
	assert.Len(t, args, 2)

	valuer, ok := args[0].(driver.Valuer)
	assert.True(t, ok)
	val, err := valuer.Value()
	assert.Nil(t, err)
	assert.Equal(t, `{"a","b","c"}`, val)
	assert.Equal(t, `"O'Brien"`, args[1])
}

func TestTimePointsFillerArgs(t *testing.T) {
	query, args := TimePointsFillerArgs(
		[]string{"a", "b"},
		[]time.Time{DateUTC(2020, 1, 1), DateUTC(2020, 1, 2)},
	)
	assert.True(t, strings.Contains(query, "UNNEST(?::text[])"))
	assert.True(t, strings.Contains(query, "UNNEST(?::timestamptz[])"))
	assert.Len(t, args, 2)
	assert.Equal(t, sqlx.Rebind(sqlx.DOLLAR, query),
		`SELECT id, time FROM 
	    (SELECT id FROM UNNEST($1::text[]) t(id)) a(id)
	    CROSS JOIN 
	    (SELECT time FROM UNNEST($2::timestamptz[]) t(time)) b(time)`)
}

func TestRebindRequiresArgOrder(t *testing.T) {
	whereName, nameArgs := EQxArgs("name", "alice")
	whereIDs, idArgs := INxArgs("id", []int{10, 20})

	query := fmt.Sprintf("SELECT * FROM users WHERE %s AND %s", whereName, whereIDs)
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	assert.Equal(t, "SELECT * FROM users WHERE name = $1 AND id IN ($2,$3)", query)

	correctArgs := append(append([]any{}, nameArgs...), idArgs...)
	assert.Equal(t, []any{"alice", 10, 20}, correctArgs)

	// Wrong order still compiles, but now $1/$2/$3 bind to the wrong values.
	wrongArgs := append(append([]any{}, idArgs...), nameArgs...)
	assert.Equal(t, []any{10, 20, "alice"}, wrongArgs)
	assert.NotEqual(t, correctArgs, wrongArgs)
}

func TestAllTokens(t *testing.T) {
	arr := "{1,2,3, 4, 45 ,55}"
	tokens := allTokens(arr)
	assert.Equal(t, tokens, []string{"1", "2", "3", "4", "45", "55"})

	tokens = allTokens("{}")
	assert.Equal(t, []string{}, tokens)
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

	float32Arr, err := ParseDBArray[float32](floatArray)
	assert.Nil(t, err)
	assert.Equal(t, []float32{1, 2, 3, 4}, float32Arr)

	timeArray := "{2020-01-01, 2020-01-01T20:00:00Z}"
	timeArr, err := ParseDBArray[time.Time](timeArray)
	assert.Nil(t, err)
	assert.Equal(t, []time.Time{DateUTC(2020, 1, 1), time.Date(2020, 1, 1, 20, 0, 0, 0, time.UTC)}, timeArr)

	boolArray := "{t,t,t,f}"
	boolArr, err := ParseDBArray[bool](boolArray)
	assert.Nil(t, err)
	assert.Equal(t, []bool{true, true, true, false}, boolArr)

	emptyStrings, err := ParseDBArray[string]("{}")
	assert.Nil(t, err)
	assert.Equal(t, []string{}, emptyStrings)

	emptyInts, err := ParseDBArray[int]("{}")
	assert.Nil(t, err)
	assert.Equal(t, []int{}, emptyInts)
}

func TestTimePointsToFiller(t *testing.T) {
	db := NewTestPostgresDB("")
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
	db := NewTestPostgresDB("")
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
