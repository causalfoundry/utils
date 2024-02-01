package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const PostgresPlaceholderLimit = 65535

var Psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func UpdateClause(headers, pks []string, mergeStrategy map[string]string) string {
	toUpdate := Except(headers, pks)
	var tmp []string
	for _, update := range toUpdate {
		if val, ok := mergeStrategy[update]; ok {
			tmp = append(tmp, fmt.Sprintf("%s=%s", update, val))
		} else {
			tmp = append(tmp, fmt.Sprintf("%s=excluded.%s", update, update))
		}
	}
	return strings.Join(tmp, ", ")
}

func EQx[T any](name string, v T) string {
	switch val := any(v).(type) {
	case int, float64, bool:
		return fmt.Sprintf("%s = %v", name, v)
	case time.Time:
		return fmt.Sprintf("%s = '%s'", name, val.Format(time.RFC3339))
	case string:
		return fmt.Sprintf("%s = '%v'", name, v)
	case []int:
		return INx(name, any(v).([]int))
	case []string:
		return INx(name, any(v).([]string))
	default:
		panic(fmt.Errorf("unsupported type: %T", any(v)))
	}
}

func INx[T any](name string, ins []T) string {
	var str []string
	for _, i := range ins {
		switch any(i).(type) {
		case int:
			str = append(str, fmt.Sprint(i))
		case string:
			str = append(str, fmt.Sprintf("'%v'", i))
		}
	}
	return fmt.Sprintf("%s IN (%s)", name, strings.Join(str, ","))
}

func Partition[T int | string](partitionTableName string, fullTableName string, partitionValue T) (res string) {
	var valueStr string

	// can perform a type assertion only on interface values
	// https://stackoverflow.com/questions/71587996/cannot-use-type-assertion-on-type-parameter-value
	switch v := any(partitionValue).(type) {
	case int:
		valueStr = fmt.Sprintf("%d", v)
	case string:
		valueStr = fmt.Sprintf("'%s'", v)
	}
	res = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS partition.%s
		PARTITION OF %s
		FOR VALUES IN (%s)`,
		partitionTableName,
		fullTableName,
		valueStr,
	)
	return
}

func MultipleJsonbSet(fieldName string, setMap map[string]any) (ret string) {
	build := func(field, k string, v any) string {
		switch vv := v.(type) {
		case string:
			return fmt.Sprintf(`JSONB_SET(%s, '%s', '"%s"')`, field, k, vv)
		default:
			return fmt.Sprintf(`JSONB_SET(%s, '%s', '%v')`, field, k, vv)
		}
	}

	for k, v := range setMap {
		if ret == "" {
			ret = build(fieldName, k, v)
		} else {
			ret = build(ret, k, v)
		}
	}
	return
}

type InsertHelper struct {
	base    squirrel.InsertBuilder
	builder squirrel.InsertBuilder

	execer sqlx.Execer
	lens   int
	cnt    int
}

func NewInsertHelper(table string, cols []string, mergeStrategy string, execer sqlx.Execer) InsertHelper {
	base := Psql.Insert(table).Columns(cols...).Suffix(mergeStrategy)
	return InsertHelper{
		base:    base,
		builder: base,
		execer:  execer,
		lens:    len(cols),
	}
}

func (ih *InsertHelper) Add(args []any) (err error) {
	if len(args) != ih.lens {
		return errors.New("args size not consistent")
	}
	ih.builder = ih.builder.Values(args...)
	ih.cnt += ih.lens

	if ih.cnt > PostgresPlaceholderLimit*0.8 {
		query, args, _ := ih.builder.ToSql()
		_, err = ih.execer.Exec(query, args...)
		if err != nil {
			return
		}
		ih.builder = ih.base
		ih.cnt = 0
	}
	return
}

func (ih *InsertHelper) Finish() (err error) {
	if ih.cnt > 0 {
		query, args, _ := ih.builder.ToSql()
		_, err = ih.execer.Exec(query, args...)
		if err != nil {
			err = fmt.Errorf("err: %w, query: %s, args: %v", err, query, args)
		}
	}
	return
}

func CloseToPlaceholderLimit(cnt, stride int) int {
	if PostgresPlaceholderLimit-cnt < stride {
		return 0
	}
	return cnt + stride
}

func ToPqStrArray[T any](arr []T) (ret pq.StringArray) {
	ret = make(pq.StringArray, len(arr))
	for i, a := range arr {
		ret[i] = fmt.Sprint(a)
	}
	return
}

func ValueListTable[T any](arr []T, label string) string {
	return fmt.Sprintf("SELECT * FROM %s ", ValueList(arr, label, false))
}

func ValueList[T any](arr []T, label string, withOrder bool) string {
	var tmp = make([]string, len(arr))

	for i := range arr {
		switch any(*new(T)).(type) {
		case string:
			tmp[i] = fmt.Sprintf("('%v')", arr[i])
		case time.Time:
			tmp[i] = fmt.Sprintf("('%v')", any(arr[i]).(time.Time).Format(time.RFC3339))
		default:
			tmp[i] = fmt.Sprintf("(%v)", arr[i])
		}

		if withOrder {
			tmp[i] = tmp[i][:len(tmp[i])-1] + fmt.Sprintf(",%d)", i)
		}
	}

	if withOrder {
		label = fmt.Sprintf("%s,sort_order", label)
	}
	return fmt.Sprintf("(VALUES %s) a(%s)", strings.Join(tmp, ","), label)
}

func SelectStr(cols ...string) string {
	return strings.Join(cols, ", ")
}

func UpsertQuery(cols []string) string {
	for i := range cols {
		cols[i] = fmt.Sprintf("%s=excluded.%s", cols[i], cols[i])
	}
	return strings.Join(cols, ", ")
}

func TimePointsFiller(ids []string, pts []time.Time) string {
	return fmt.Sprintf(
		`SELECT id, time FROM 
    (SELECT id FROM UNNEST('%s'::_varchar) t(id)) a(id)
    CROSS JOIN 
    (SELECT time FROM UNNEST('%s'::_timestamptz) t(time)) b(time)`,
		dbArrayStr(ids),
		timeArrayToStrArray(pts),
	)
}

func timeArrayToStrArray(ts []time.Time) string {
	var s []string
	for _, t := range ts {
		s = append(s, t.Format(time.RFC3339))
	}
	return fmt.Sprintf("{%s}", strings.Join(s, ","))
}

func dbArrayStr(s []string) string {
	return fmt.Sprintf("{%s}", strings.Join(s, ","))
}

func Range(start, end int) (ret []int) {
	for s := start; s <= end; s++ {
		ret = append(ret, s)
	}
	return
}

func DBArrayValues[T any](array []T) string {
	var tmp = make([]string, len(array))
	for i := range array {
		switch v := any(array[i]).(type) {
		case time.Time:
			tmp[i] = v.Format(time.RFC3339)
		case string:
			if v == "" {
				tmp[i] = `""`
			} else {
				tmp[i] = v
			}
		default:
			tmp[i] = fmt.Sprint(array[i])
		}
	}
	return fmt.Sprintf("{%s}", strings.Join(tmp, ","))
}

// collect coma separated one-level array token
func allTokens(repr string) []string {
	// remove trailing or leading comma (for whatever reason)
	repr = strings.Trim(repr, ",")

	// remove brackets
	repr = strings.Trim(strings.Trim(repr, "{"), "}")

	ret := strings.Split(repr, ",")
	for i, r := range ret {
		ret[i] = strings.Trim(r, " ")
	}
	return ret
}

// ParseDBArray
// supported type is based on the needed usecases (number, time, etc)
func ParseDBArray[T any](repr string) (ret []T, err error) {
	tokens := allTokens(repr)
	var result any

	switch any(*new(T)).(type) {
	case string:
		result = tokens
	case int:
		var res []int
		var v int
		for _, token := range tokens {
			v, err = strconv.Atoi(token)
			res = append(res, v)
		}
		result = res
	case float64, float32:
		var res []float64
		var v float64
		for _, token := range tokens {
			if v, err = strconv.ParseFloat(token, 64); err != nil {
				return
			}
			res = append(res, v)
		}
		result = res
	case time.Time:
		var res []time.Time
		var v time.Time
		for _, token := range tokens {
			// remove escape & quotes
			token = strings.ReplaceAll(token, "\"", "")

			if v, err = time.Parse(DBTimestamptzFormat, token); err == nil {
				res = append(res, v.UTC())
				continue
			}

			if v, err = time.Parse(TimestamptzFormat, token); err == nil {
				res = append(res, v.UTC())
				continue
			}

			if v, err = time.Parse(DateFormat, token); err == nil {
				res = append(res, v.UTC())
				continue
			}

			// failed to parse the result
			return
		}
		result = res
	default:
		err = errors.New("invalid type")
	}

	asserted, ok := result.([]T)
	if !ok {
		err = errors.New("error type assertion")
	}
	return asserted, err
}

func AndWhere(wheres []string) string {
	if len(wheres) == 0 {
		wheres = []string{"1=1"}
	}
	return strings.Join(wheres, " AND ")
}

func EitherRunner(tx *sqlx.Tx, db *sqlx.DB) squirrel.BaseRunner {
	if tx != nil {
		return tx
	}
	return db
}