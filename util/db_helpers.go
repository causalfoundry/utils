package util

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

func UpsertManyNoPartition[T any](log zerolog.Logger, con sq.BaseRunner, table string, pks []string, tag string, toInsert []T, mergeStrategy map[string]string) (err error) {
	if len(toInsert) == 0 {
		return
	}
	cols, _ := ExtractTags(toInsert[0], tag)
	merge := fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s",
		strings.Join(pks, ","),
		UpdateClause(cols, pks, mergeStrategy))

	helper := NewInsertHelper(table, cols, merge, con)

	for _, t := range toInsert {
		_, vals := ExtractTags(t, "json")
		if err = helper.Add(vals); err != nil {
			log.Err(err).Msg("error save")
			return
		}
	}
	if err = helper.Finish(); err != nil {
		log.Err(err).Msg("error finish insert many")
	}
	return
}

func UpdateS(log zerolog.Logger, con sq.BaseRunner, table string, where []string, sets map[string]any) (err error) {
	_, err = Psql.Update(table).
		SetMap(sets).
		Where(AndWhere(where)).
		RunWith(con).
		Exec()
	if err != nil {
		log.Err(err).Interface("where", where).Str("table", table).Interface("updates", sets).Msg("error update")
	}
	return err
}

func ListS[T any](log zerolog.Logger, con *sqlx.DB, page Page, table string, wheres, order []string) (ret []T, total int, err error) {
	order = append(order, "1")

	col, _ := ExtractTags(*new(T), "db")

	query, args, _ := Psql.Select(col...).
		From(table).
		Where(AndWhere(wheres)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		OrderBy(strings.Join(order, ",")).
		ToSql()

	if err = con.Select(&ret, query, args...); err != nil {
		log.Err(err).Str("table", table).Msg("error select")
		return
	}

	err = Psql.Select("COUNT(*)").From(table).Where(AndWhere(wheres)).
		RunWith(con).QueryRow().Scan(&total)
	if err != nil {
		log.Err(err).Str("table", table).Msg("error count total")
	}
	return
}

func GetS[T any](log zerolog.Logger, con *sqlx.DB, table string, where []string) (ret T, err error) {
	query, args, _ := Psql.Select("*").
		From(table).
		Where(AndWhere(where)).
		ToSql()

	if err = con.QueryRowx(query, args...).StructScan(&ret); err != nil {
		log.Err(err).Strs("where", where).Str("table", table).Msg("error delete")
	}
	return
}

func GetManyS[T any](log zerolog.Logger, con *sqlx.DB, table string, where []string, orderby []string) (ret []T, err error) {
	query, args, _ := Psql.Select("*").From(table).
		Where(AndWhere(where)).
		OrderBy(strings.Join(orderby, ",")).
		ToSql()
	if err = con.Select(&ret, query, args...); err != nil {
		log.Err(err).Type("type", ret).Interface("where", where).Str("table", table).Msg("error get many")
	}
	return
}

func ExistS(log zerolog.Logger, con *sqlx.DB, table string, where []string) (bool, error) {
	var cnt int
	query, args, _ := Psql.Select("count(*)").From(table).Where(AndWhere(where)).ToSql()
	row := con.QueryRow(query, args...)
	err := row.Scan(&cnt)
	if err != nil {
		log.Err(err).Str("query", query).Interface("args", args).Str("table", table).Msg("error delete")
	}
	return cnt > 0, err
}

func Delete(log zerolog.Logger, con sq.BaseRunner, table string, where []string) (err error) {
	if len(where) == 0 {
		return nil
	}
	if _, err = Psql.Delete(table).Where(AndWhere(where)).RunWith(con).Exec(); err != nil {
		log.Err(err).Strs("where", where).Str("table", table).Msg("error delete")
	}
	return err
}

func ListFlex[T any](log zerolog.Logger, con *sqlx.DB, page Page, table string, selects, where, orderby []string) (ret []T, total int, err error) {
	var toSelect = "*"
	where = append(where, "1=1")
	orderby = append(orderby, "(SELECT NULL)")
	if len(selects) != 0 {
		toSelect = strings.Join(selects, ",")
	}

	w := strings.Join(where, " AND ")
	o := strings.Join(orderby, ",")
	query, args, _ := Psql.Select(toSelect).From(table).
		Where(w).
		OrderBy(o).
		Offset(page.Offset()).
		Limit(page.Limit()).
		ToSql()

	ret = make([]T, 0)
	if err = con.Select(&ret, query, args...); err != nil {
		log.Err(err).Str("table", table).Str("query", query).Interface("args", args).Msg("error select")
		return
	}

	err = Psql.Select("COUNT(*)").
		From(table).
		Where(w).
		RunWith(con).
		QueryRow().Scan(&total)
	if err != nil {
		log.Err(err).Str("table", table).Str("query", query).Interface("args", args).Msg("error get count")
	}

	if table == "common.product" {
		log.Info().Str("query", query).Interface("args", args).Msg("list flex")
	}
	return
}

func ListM[T any](log zerolog.Logger, con *sqlx.DB, page Page, table string, where map[string]any, orderBy []string) (ret []T, total int, err error) {
	ret = make([]T, 0)

	base := Psql.Select("*").From(table)

	if len(where) != 0 {
		base = base.Where(where)
	}
	if len(orderBy) != 0 {
		base = base.OrderBy(strings.Join(orderBy, ","))
	}

	query, args, _ := base.
		Offset(page.Offset()).
		Limit(page.Limit()).
		ToSql()

	if err = con.Select(&ret, query, args...); err != nil {
		log.Err(err).Str("table", table).Str("query", query).Interface("args", args).Msg("error select")
		return
	}

	err = Psql.Select("COUNT(*)").From(table).Where(where).
		RunWith(con).QueryRow().Scan(&total)
	if err != nil {
		log.Err(err).Str("table", table).Msg("error count total")
	}
	return
}

func UpdateMore(log zerolog.Logger, con sq.BaseRunner, table string, ids []int, sets map[string]any) (err error) {
	_, err = Psql.Update(table).
		SetMap(sets).
		Where(sq.Eq{"id": ids}).
		RunWith(con).
		Exec()
	if err != nil {
		log.Err(err).Ints("id", ids).Str("table", table).Interface("updates", sets).Msg("error update more")
	}
	return err
}

func UpdateM(log zerolog.Logger, con sq.BaseRunner, table string, where map[string]any, sets map[string]any) (err error) {
	_, err = Psql.Update(table).
		SetMap(sets).
		Where(where).
		RunWith(con).
		Exec()
	if err != nil {
		log.Err(err).Interface("where", where).Str("table", table).Interface("updates", sets).Msg("error update")
	}
	return err
}

func UpdateWithID(log zerolog.Logger, con sq.BaseRunner, table string, id int, sets map[string]any) (err error) {
	where := Obj{"id": id}
	return UpdateM(log, con, table, where, sets)
}

func SetDelete(log zerolog.Logger, con *sqlx.DB, table string, id int, del bool) (err error) {
	_, err = Psql.
		Update(table).
		Where(sq.Eq{"id": id}).
		Set("is_deleted", del).
		RunWith(con).Exec()
	if err != nil {
		log.Err(err).Int("id", id).Bool("del", del).Str("table", table).Msg("error set delete")
	}
	return
}

func SetActive(log zerolog.Logger, con *sqlx.DB, table string, id int, active bool) (err error) {
	_, err = Psql.
		Update(table).
		Where(sq.Eq{"id": id}).
		Set("is_active", active).
		RunWith(con).Exec()
	if err != nil {
		log.Err(err).Int("id", id).Bool("active", active).Msg("error set active")
	}
	return
}

type PartitionFunc func(runner sq.BaseRunner, table string) error

func CreateMany[T any](log zerolog.Logger, con sq.BaseRunner, table string, reqs []T, returning []string, partitionFunc PartitionFunc) (ids []int, err error) {
	if len(reqs) == 0 {
		return
	}
	if partitionFunc != nil {
		if err = partitionFunc(con, table); err != nil {
			return
		}
	}

	cols, _ := ExtractDBTags(reqs[0])

	base := Psql.Insert(table).Columns(cols...)
	if len(returning) != 0 {
		base = base.Suffix("RETURNING " + strings.Join(returning, ","))
	}
	for _, req := range reqs {
		_, vals := ExtractDBTags(req)
		base = base.Values(vals...)
	}

	rows, err := base.RunWith(con).Query()
	defer func() { _ = rows.Close() }()
	if err != nil {
		log.Err(err).Msg("error create many")
		return
	}
	var id int
	for rows.Next() {
		if err = rows.Scan(&id); err != nil {
			log.Err(err).Msg("error scan id")
			return
		}
		ids = append(ids, id)
	}
	return
}

func Create[T any](log zerolog.Logger, con sq.BaseRunner, table string, req T, returning []string, partitionFunc PartitionFunc) (id int, err error) {
	if partitionFunc != nil {
		if err = partitionFunc(con, table); err != nil {
			return
		}
	}

	cols, vals := ExtractDBTags(req)
	base := Psql.Insert(table).
		Columns(cols...).
		Values(vals...)
	if len(returning) != 0 {
		base = base.Suffix("RETURNING " + strings.Join(returning, ","))
		err = base.RunWith(con).QueryRow().Scan(&id)
	} else {
		_, err = base.RunWith(con).Exec()
	}
	if err != nil {
		log.Err(err).Interface("req", req).Str("table", table).Msg("error create")
	}
	return
}

func GetManyM[T any](log zerolog.Logger, con *sqlx.DB, table string, where map[string]any, orderby []string) (ret []T, err error) {
	query, args, _ := Psql.Select("*").From(table).
		Where(where).
		OrderBy(strings.Join(orderby, ",")).
		ToSql()
	if err = con.Select(&ret, query, args...); err != nil {
		log.Err(err).Type("type", ret).Interface("where", where).Str("table", table).Msg("error get")
	}
	return
}

func GetM[T any](log zerolog.Logger, con *sqlx.DB, table string, where map[string]any) (ret T, err error) {
	query, args, _ := Psql.Select("*").From(table).
		Where(where).
		ToSql()
	row := con.QueryRowx(query, args...)
	if err = row.StructScan(&ret); err != nil {
		log.Err(err).Type("type", ret).Interface("where", where).Str("table", table).Msg("error get")
		if err == sql.ErrNoRows {
			err = NewErr(http.StatusNotFound, "resource not found", map[string]string{
				"type": fmt.Sprintf("%T", ret),
			})
			return
		}
	}
	return
}

func ExistM(log zerolog.Logger, con *sqlx.DB, table string, where map[string]any) (ret bool, err error) {
	query, args, _ := Psql.Select("COUNT(*)").From(table).
		Where(where).
		ToSql()

	var cnt int
	if err = con.QueryRow(query, args...).Scan(&cnt); err != nil {
		log.Err(err).Interface("where", where).Str("table", table).Msg("error check exist")
	}
	ret = cnt > 0
	return
}
