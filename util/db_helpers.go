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

func UpsertMany[T any](log zerolog.Logger, con sq.BaseRunner, table string, pks []string, tag string, toInsert []T, mergeStrategy map[string]string, partitionFunc PartitionFunc) (err error) {
	if len(toInsert) == 0 {
		return
	}

	if partitionFunc != nil {
		if err = partitionFunc(con, table); err != nil {
			return
		}
	}

	cols, _ := ExtractTags(toInsert[0], tag, []string{})
	merge := fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s",
		strings.Join(pks, ","),
		UpdateClause(cols, pks, mergeStrategy))

	helper := NewInsertHelper(table, cols, merge, con)

	for _, t := range toInsert {
		_, vals := ExtractTags(t, tag, []string{})
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
	ret = []T{}
	order = append(order, "1")

	col, _ := ExtractTags(*new(T), "db", []string{})

	query, args, _ := Psql.Select(col...).
		From(table).
		Where(AndWhere(wheres)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		OrderBy(strings.Join(order, ",")).
		ToSql()

	if err = con.Select(&ret, query, args...); err != nil {
		log.Err(err).Str("table", table).Str("query", query).Interface("args", args).Msg("error select")
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
	ret = []T{}
	tmp := make([]T, 1)
	cols, _ := ExtractTags(tmp[0], "db", nil)
	query, args, _ := Psql.Select(cols...).From(table).
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
	ret = []T{}
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
	ret = []T{}

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

func ListQ[T any](log zerolog.Logger, con *sqlx.DB, page Page, table string, where sq.Sqlizer, order []string) (ret []T, total int, err error) {
	ret = []T{}

	cols := dbColsForType[T]()
	orderBy, err := normalizeOrderBy(order, cols)
	if err != nil {
		log.Err(err).Str("table", table).Strs("order", order).Msg("error normalize order")
		return
	}

	base := Psql.Select(cols...).From(table)
	if where != nil {
		base = base.Where(where)
	}
	if len(orderBy) != 0 {
		base = base.OrderBy(orderBy...)
	}

	query, args, err := base.
		Offset(page.Offset()).
		Limit(page.Limit()).
		ToSql()
	if err != nil {
		log.Err(err).Str("table", table).Interface("where", where).Msg("error build list query")
		return
	}

	if err = con.Select(&ret, query, args...); err != nil {
		log.Err(err).Str("table", table).Str("query", query).Interface("args", args).Msg("error select")
		return
	}

	total, err = CountQ(log, con, table, where)
	return
}

func GetQ[T any](log zerolog.Logger, con *sqlx.DB, table string, where sq.Sqlizer) (ret T, err error) {
	cols := dbColsForType[T]()
	base := Psql.Select(cols...).From(table)
	if where != nil {
		base = base.Where(where)
	}

	query, args, err := base.ToSql()
	if err != nil {
		log.Err(err).Str("table", table).Interface("where", where).Msg("error build get query")
		return
	}

	if err = con.QueryRowx(query, args...).StructScan(&ret); err != nil {
		log.Err(err).Type("type", ret).Str("table", table).Str("query", query).Interface("args", args).Msg("error get")
		if err == sql.ErrNoRows {
			err = NewErr(http.StatusNotFound, "resource not found", map[string]string{
				"type": fmt.Sprintf("%T", ret),
			})
		}
	}
	return
}

func GetManyQ[T any](log zerolog.Logger, con *sqlx.DB, table string, where sq.Sqlizer, order []string) (ret []T, err error) {
	ret = []T{}

	cols := dbColsForType[T]()
	orderBy, err := normalizeOrderBy(order, cols)
	if err != nil {
		log.Err(err).Str("table", table).Strs("order", order).Msg("error normalize order")
		return
	}

	base := Psql.Select(cols...).From(table)
	if where != nil {
		base = base.Where(where)
	}
	if len(orderBy) != 0 {
		base = base.OrderBy(orderBy...)
	}

	query, args, err := base.ToSql()
	if err != nil {
		log.Err(err).Str("table", table).Interface("where", where).Msg("error build get many query")
		return
	}

	if err = con.Select(&ret, query, args...); err != nil {
		log.Err(err).Type("type", ret).Str("table", table).Str("query", query).Interface("args", args).Msg("error get many")
	}
	return
}

func CountQ(log zerolog.Logger, con *sqlx.DB, table string, where sq.Sqlizer) (total int, err error) {
	base := Psql.Select("COUNT(*)").From(table)
	if where != nil {
		base = base.Where(where)
	}

	query, args, err := base.ToSql()
	if err != nil {
		log.Err(err).Str("table", table).Interface("where", where).Msg("error build count query")
		return
	}

	if err = con.QueryRow(query, args...).Scan(&total); err != nil {
		log.Err(err).Str("table", table).Str("query", query).Interface("args", args).Msg("error count")
	}
	return
}

func ExistQ(log zerolog.Logger, con *sqlx.DB, table string, where sq.Sqlizer) (ret bool, err error) {
	base := Psql.Select("1").From(table).Limit(1)
	if where != nil {
		base = base.Where(where)
	}

	query, args, err := base.ToSql()
	if err != nil {
		log.Err(err).Str("table", table).Interface("where", where).Msg("error build exist query")
		return
	}

	var one int
	if err = con.QueryRow(query, args...).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Err(err).Str("table", table).Str("query", query).Interface("args", args).Msg("error exist")
		return false, err
	}
	return true, nil
}

func UpdateQ(log zerolog.Logger, con sq.BaseRunner, table string, where sq.Sqlizer, sets map[string]any) (err error) {
	if where == nil {
		return fmt.Errorf("refuse update without where")
	}
	if len(sets) == 0 {
		return nil
	}

	base := Psql.Update(table).SetMap(sets).Where(where)
	query, args, err := base.ToSql()
	if err != nil {
		log.Err(err).Str("table", table).Interface("where", where).Interface("updates", sets).Msg("error build update query")
		return
	}

	if _, err = con.Exec(query, args...); err != nil {
		log.Err(err).Str("table", table).Str("query", query).Interface("args", args).Interface("updates", sets).Msg("error update")
	}
	return
}

func DeleteQ(log zerolog.Logger, con sq.BaseRunner, table string, where sq.Sqlizer) (err error) {
	if where == nil {
		return fmt.Errorf("refuse delete without where")
	}

	base := Psql.Delete(table).Where(where)
	query, args, err := base.ToSql()
	if err != nil {
		log.Err(err).Str("table", table).Interface("where", where).Msg("error build delete query")
		return
	}

	if _, err = con.Exec(query, args...); err != nil {
		log.Err(err).Str("table", table).Str("query", query).Interface("args", args).Msg("error delete")
	}
	return
}

func dbColsForType[T any]() []string {
	cols, _ := ExtractTags(*new(T), "db", nil)
	return cols
}

func normalizeOrderBy(order, allowed []string) (ret []string, err error) {
	if len(order) == 0 {
		return nil, nil
	}

	allowedSet := make(map[string]struct{}, len(allowed))
	for _, col := range allowed {
		allowedSet[col] = struct{}{}
	}

	for _, raw := range order {
		parts := strings.Fields(strings.TrimSpace(raw))
		if len(parts) == 0 {
			continue
		}
		if len(parts) > 2 {
			return nil, fmt.Errorf("invalid order clause: %q", raw)
		}

		col := parts[0]
		baseCol := col
		if idx := strings.LastIndex(col, "."); idx >= 0 {
			baseCol = col[idx+1:]
		}
		if _, ok := allowedSet[baseCol]; !ok {
			return nil, fmt.Errorf("invalid order column: %q", col)
		}

		dir := "ASC"
		if len(parts) == 2 {
			dir = strings.ToUpper(parts[1])
			if dir != "ASC" && dir != "DESC" {
				return nil, fmt.Errorf("invalid order direction: %q", raw)
			}
		}
		ret = append(ret, fmt.Sprintf("%s %s", col, dir))
	}

	return ret, nil
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

func CreateManySkip[T any](log zerolog.Logger, con sq.BaseRunner, table string, reqs []T, returning, skipping []string, partitionFunc PartitionFunc) (ids []int, err error) {
	if len(reqs) == 0 {
		return
	}
	if partitionFunc != nil {
		if err = partitionFunc(con, table); err != nil {
			return
		}
	}

	cols, _ := ExtractTags(reqs[0], "db", skipping)

	base := Psql.Insert(table).Columns(cols...)
	if len(returning) != 0 {
		base = base.Suffix("RETURNING " + strings.Join(returning, ","))
	}
	for _, req := range reqs {
		_, vals := ExtractTags(req, "db", skipping)
		base = base.Values(vals...)
	}

	query, args, err := base.ToSql()
	if err != nil {
		return
	}
	rows, err := con.Query(query, args...)
	defer func() {
		if rows != nil {
			_ = rows.Close()
		}
	}()
	if err != nil {
		log.Err(err).Str("query", query).Msg("error create many")
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

func CreateMany[T any](log zerolog.Logger, con sq.BaseRunner, table string, reqs []T, returning []string, partitionFunc PartitionFunc) (ids []int, err error) {
	return CreateManySkip[T](log, con, table, reqs, returning, []string{}, partitionFunc)
}

func CreateSkip[T any](log zerolog.Logger, con sq.BaseRunner, table string, req T, returning, skipping []string, partitionFunc PartitionFunc) (id int, err error) {
	if partitionFunc != nil {
		if err = partitionFunc(con, table); err != nil {
			return
		}
	}

	cols, vals := ExtractTags(req, "db", skipping)
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

func Create[T any](log zerolog.Logger, con sq.BaseRunner, table string, req T, returning []string, partitionFunc PartitionFunc) (id int, err error) {
	return CreateSkip[T](log, con, table, req, returning, []string{}, partitionFunc)
}

func GetManyM[T any](log zerolog.Logger, con *sqlx.DB, table string, where map[string]any, orderby []string) (ret []T, err error) {
	ret = []T{}
	tmp := make([]T, 1)
	cols, _ := ExtractTags(tmp[0], "db", nil)
	query, args, _ := Psql.Select(cols...).From(table).
		Where(where).
		OrderBy(strings.Join(orderby, ",")).
		ToSql()
	if err = con.Select(&ret, query, args...); err != nil {
		log.Err(err).Type("type", ret).Interface("where", where).Str("table", table).Msg("error get")
	}
	return
}

func GetM[T any](log zerolog.Logger, con *sqlx.DB, table string, where map[string]any) (ret T, err error) {
	tmp := make([]T, 1)
	cols, _ := ExtractTags(tmp[0], "db", nil)
	query, args, _ := Psql.Select(cols...).From(table).
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
