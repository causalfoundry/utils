package util

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const PostgresPlaceholderLimit = 65535

var Psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type PostgresConfig struct {
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	DatabaseName    string `yaml:"database_name"`
	Host            string `yaml:"host"`
	DatabaseSSLMode string `yaml:"database_ssl_mode"`
	Port            int    `yaml:"port"`
}

func (p PostgresConfig) GetLocalBaseURL() string {
	return fmt.Sprintf("host=%s port=%d dbname=postgres user=%s password=%s sslmode=%s",
		p.Host, p.Port, p.Username, p.Password, p.DatabaseSSLMode)
}

func (p PostgresConfig) GetURL() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		p.Host, p.Port, p.DatabaseName, p.Username, p.Password, p.DatabaseSSLMode)
}

type RedisConfig struct {
	Master   string `yaml:"master"`
	Slave    string `yaml:"slave"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
}

type DBConfig struct {
	MaxOpenConn     int
	MaxIdelConn     int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

func (d *DBConfig) SetDefault() {
	if d.MaxOpenConn == 0 {
		d.MaxOpenConn = 100
	}
	if d.MaxIdelConn == 0 {
		d.MaxIdelConn = 10
	}
	if d.ConnMaxIdleTime == 0 {
		d.ConnMaxIdleTime = time.Second * 60
	}
	if d.ConnMaxLifetime == 0 {
		d.ConnMaxLifetime = time.Second * 300
	}
}

// base db name and access is hardcoded, and it's linked with Storage-Up in Makefile
func NewTestDB(migrationPath string) *sqlx.DB {
	dbName := RandomAlphabets(10, true)
	baseUrl := "host=localhost port=5432 dbname=postgres user=user password=pwd sslmode=disable"
	dbUrl := strings.ReplaceAll(baseUrl, "postgres", dbName)
	SetupLocalStorage(dbName, "postgres", baseUrl, migrationPath)
	fmt.Println("---- " + dbName)
	return NewDB(dbName, dbUrl, DBConfig{})
}

func NewDB(dbName, url string, dcfg DBConfig) *sqlx.DB {
	db, err := sqlx.Open("pgx", url)
	if err != nil {
		panic("cannot get postgres connection: " + err.Error())
	}

	if err = db.Ping(); err != nil {
		panic("cannot ping postgres: " + err.Error())
	}

	dcfg.SetDefault()
	db.SetMaxOpenConns(dcfg.MaxOpenConn)
	db.SetConnMaxIdleTime(dcfg.ConnMaxIdleTime)
	db.SetMaxIdleConns(dcfg.MaxIdelConn)
	db.SetConnMaxLifetime(dcfg.ConnMaxLifetime)

	_, err = db.Exec(fmt.Sprintf("ALTER DATABASE %s SET timezone TO 'UTC'", dbName))
	if err != nil {
		panic("Error in setting timezone: " + err.Error())
	}

	// Set the timezone for the current session
	_, err = db.Exec("SET TIMEZONE TO 'UTC'")
	if err != nil {
		panic("Error in setting session timezone: " + err.Error())
	}

	return db
}

// skip migration if migration file path is empty
func SetupLocalStorage(newDB, baseDB, baseUrl, migrationFile string) {
	db, err := sql.Open("pgx", baseUrl)
	if err != nil {
		panic(fmt.Sprintf("error get base database connection: %s", err.Error()))
	}

	var exist bool
	row := db.QueryRow("select exists (select 1 from pg_database where datname = $1)", newDB)
	err = row.Scan(&exist)
	if err != nil {
		panic(err.Error())
	}

	if !exist {
		_, err = db.Exec("CREATE DATABASE " + newDB)
		if err != nil {
			panic("error create local database: " + err.Error())
		}
		db.Close()

		newUrl := strings.ReplaceAll(baseUrl, baseDB, newDB)
		if db, err = sql.Open("pgx", newUrl); err != nil {
			panic(fmt.Sprintf("error get new database connection: %s", err.Error()))
		}

		if migrationFile != "" {
			driver, _ := postgres.WithInstance(db, &postgres.Config{})
			migrateInstance, err := migrate.NewWithDatabaseInstance(
				"file://"+migrationFile,
				newDB,
				driver,
			)
			Panic(err)

			err = migrateInstance.Up()
			if err != nil {
				Panic(err)
			}
		}
	}
}

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
		if len(query) > 100 {
			query = query[:100]
		}
		if len(args) > 100 {
			args = args[:100]
		}
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
	case bool:
		var res []bool
		var v bool
		for _, token := range tokens {
			if token == "t" {
				v = true
			} else if token == "f" {
				v = false
			} else {
				err = fmt.Errorf("unrecognised token %s in bool array", token)
				return
			}
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

func toSqlStr[T uint | int | string](v T) (res string) {
	switch v := any(v).(type) {
	case uint, int:
		res = fmt.Sprintf("%d", v)
	case string:
		res = fmt.Sprintf("'%s'", v)
	}
	return
}

// there are three types of partition in postgres
// - range
// - list
// - hash
func CreateRangePartition[T uint | int | string](partitionName string, parentName string, from, to T) string {
	return fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s
		PARTITION OF %s
		FOR VALUES FROM (%s) TO (%s)`,
		partitionName,
		parentName,
		toSqlStr(from),
		toSqlStr(to),
	)
}

func CreateListPartition[T uint | int | string](partitionName string, parentName string, val T) string {
	return fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s
		PARTITION OF %s
		FOR VALUES IN (%s)`,
		partitionName,
		parentName,
		toSqlStr(val),
	)
}

// it assume certain structure of the table and index name
func PrepareIDPartition(con squirrel.BaseRunner, table, partitionSchema string, modulos int) error {
	shortTable := strings.Split(table, ".")[1]
	query := fmt.Sprintf("SELECT last_value FROM %s_id_seq LIMIT 1", table)
	row, err := con.Query(query)
	if err != nil {
		return err
	}

	var next uint
	// the for loop should always have 1 iteration
	for row.Next() {
		if err = row.Scan(&next); err != nil {
			return err
		}
	}

	a := next / uint(modulos)

	currentTable := fmt.Sprintf("%s.%s_%d", partitionSchema, shortTable, a)
	currentPartitionQuery := CreateRangePartition[uint](currentTable, table, (a)*uint(modulos), (a+1)*uint(modulos))

	partitionTable := fmt.Sprintf("%s.%s_%d", partitionSchema, shortTable, a+1)
	partitionQuery := CreateRangePartition[uint](partitionTable, table, (a+1)*uint(modulos), (a+2)*uint(modulos))
	_, err = con.Exec(currentPartitionQuery + ";" + partitionQuery)
	return err
}
