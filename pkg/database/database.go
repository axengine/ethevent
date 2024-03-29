package database

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"strings"
)

type DBO struct {
	conn   *sqlx.DB
	logger *zap.Logger
}

func New(dial string, logger *zap.Logger) *DBO {
	conn, err := sqlx.Connect("sqlite3", dial)
	if err != nil {
		logger.Panic("Connect", zap.Error(err))
	}

	if _, err = conn.Exec("PRAGMA cache_size = 8000;"); err != nil {
		logger.Error("Init DB Set cachesize", zap.Error(err))
	}
	if _, err = conn.Exec("PRAGMA synchronous = OFF;"); err != nil {
		logger.Error("Init DB Set synchronous", zap.Error(err))
	}
	if _, err = conn.Exec("PRAGMA temp_store = MEMORY;"); err != nil {
		logger.Error("Init DB Set temp_store", zap.Error(err))
	}
	return &DBO{
		conn:   conn,
		logger: logger,
	}
}

func (dbo *DBO) Transaction(fn func(tx *sql.Tx) error) error {
	tx, err := dbo.conn.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (dbo *DBO) Exec(ctx context.Context, tx *sql.Tx, query string, args ...any) (sql.Result, error) {
	if tx != nil {
		if len(args) == 0 {
			return tx.ExecContext(ctx, query)
		}
		return tx.ExecContext(ctx, query, args)
	}
	dbo.logger.Debug("Exec", zap.String("query", query), zap.Any("args", args))

	if len(args) == 0 {
		return dbo.conn.ExecContext(ctx, query)
	}
	return dbo.conn.ExecContext(ctx, query, args)
}

func (dbo *DBO) Insert(ctx context.Context, tx *sql.Tx, table string, fields []Feild) (sql.Result, error) {
	if table == "" || len(fields) == 0 {
		return nil, errors.New("nothing to insert")
	}

	var sqlBuff bytes.Buffer

	// fill field name
	sqlBuff.WriteString(fmt.Sprintf("insert into %s (", table))
	for i := 0; i < len(fields)-1; i++ {
		sqlBuff.WriteString(fmt.Sprintf("[%s],", fields[i].Name))
	}
	sqlBuff.WriteString(fmt.Sprintf("%s) values (", fields[len(fields)-1].Name))

	// fill field value
	for i := 0; i < len(fields)-1; i++ {
		sqlBuff.WriteString("?,")
	}
	sqlBuff.WriteString("?);")

	// execute
	values := make([]interface{}, len(fields))
	for i, v := range fields {
		values[i] = v.Value
	}
	dbo.logger.Debug("Insert", zap.String("sql", sqlBuff.String()), zap.Any("values", values))

	if tx != nil {
		return tx.ExecContext(ctx, sqlBuff.String(), values...)
	}
	return dbo.conn.ExecContext(ctx, sqlBuff.String(), values...)
}

// Delete delete records
func (dbo *DBO) Delete(ctx context.Context, tx *sql.Tx, table string, where []Where) (sql.Result, error) {
	if table == "" {
		return nil, errors.New("table name is required")
	}
	if len(where) == 0 {
		return nil, errors.New("table-clearing is not allowed")
	}

	var sqlBuff bytes.Buffer
	sqlBuff.WriteString(fmt.Sprintf("delete from %s where 1 = 1", table))
	for i := 0; i < len(where); i++ {
		sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
	}

	// execute
	values := make([]interface{}, len(where))
	for i, v := range where {
		values[i] = v.Value
	}

	dbo.logger.Debug("Delete", zap.String("sql", sqlBuff.String()), zap.Any("values", values))

	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.ExecContext(ctx, sqlBuff.String(), values...)
	} else {
		res, err = dbo.conn.ExecContext(ctx, sqlBuff.String(), values...)
	}

	return res, err
}

// Update update records
func (dbo *DBO) Update(ctx context.Context, tx *sql.Tx, table string, toupdate []Feild, where []Where) (sql.Result, error) {
	if table == "" {
		return nil, errors.New("table name is required")
	}
	if len(where) == 0 {
		return nil, errors.New("full-table-update is not allowed")
	}
	if len(toupdate) == 0 {
		return nil, errors.New("to-update-nothing is not allowed")
	}

	var sqlBuff bytes.Buffer
	sqlBuff.WriteString(fmt.Sprintf(" update %s set %s = ? ", table, toupdate[0].Name))
	for i := 1; i < len(toupdate); i++ {
		sqlBuff.WriteString(fmt.Sprintf(", %s = ? ", toupdate[i].Name))
	}

	sqlBuff.WriteString(fmt.Sprintf(" where 1 = 1 "))
	for i := 0; i < len(where); i++ {
		sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
	}

	// execute
	values := make([]interface{}, len(toupdate)+len(where))
	for i, v := range toupdate {
		values[i] = v.Value
	}
	for i, v := range where {
		values[len(toupdate)+i] = v.Value
	}

	dbo.logger.Debug("Update", zap.String("sql", sqlBuff.String()), zap.Any("values", values))

	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.ExecContext(ctx, sqlBuff.String(), values...)
	} else {
		res, err = dbo.conn.ExecContext(ctx, sqlBuff.String(), values...)
	}

	return res, err
}

func (dbo *DBO) SelectRowsUnion(ctx context.Context, table string, cols []string, wheres [][]Where, order *Order, paging *Paging, result interface{}) error {
	if table == "" {
		return errors.New("table name is required")
	}
	if len(wheres) == 0 {
		return errors.New("full-table-select is not allowed")
	}
	if order != nil && (len(order.Feilds) == 0 || order.Type == "") {
		return errors.New("order type and fields is required")
	}

	var values []interface{}

	wheresLen := len(wheres)

	var sqlBuff bytes.Buffer

	for _, where := range wheres {

		wheresLen--

		for _, v := range where {
			values = append(values, v.Value)
		}

		if len(cols) > 0 {
			sqlBuff.WriteString(fmt.Sprintf("select %s from %s where 1 = 1", strings.Join(cols, ","), table))
		} else {
			sqlBuff.WriteString(fmt.Sprintf("select * from %s where 1 = 1", table))
		}

		for i := 0; i < len(where); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
		}
		if wheresLen > 0 {
			sqlBuff.WriteString(" union ")
		}
	}

	if order != nil {
		// append order by clause for ordering
		sqlBuff.WriteString(fmt.Sprintf(" order by %s ", order.Feilds[0]))
		for i := 1; i < len(order.Feilds); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" , %s ", order.Feilds[i]))
		}
		sqlBuff.WriteString(order.Type)

		sqlBuff.WriteString(fmt.Sprintf(" limit %d offset %d ", paging.Limit, paging.CursorValue))
	}
	dbo.logger.Debug("SelectRowsUnion", zap.String("sql", sqlBuff.String()), zap.Any("values", values))

	return dbo.conn.SelectContext(ctx, result, sqlBuff.String(), values...)
}

func (dbo *DBO) SelectRowsUnionToMaps(ctx context.Context, table string, cols []string, wheres [][]Where, order *Order, paging *Paging) ([]map[string]interface{}, error) {
	if table == "" {
		return nil, errors.New("table name is required")
	}
	if len(wheres) == 0 {
		return nil, errors.New("full-table-select is not allowed")
	}
	if order != nil && (len(order.Feilds) == 0 || order.Type == "") {
		return nil, errors.New("order type and fields is required")
	}

	var values []interface{}
	wheresLen := len(wheres)
	var sqlBuff bytes.Buffer

	for _, where := range wheres {
		wheresLen--
		for _, v := range where {
			values = append(values, v.Value)
		}

		if len(cols) > 0 {
			sqlBuff.WriteString(fmt.Sprintf("select %s from %s where 1 = 1", strings.Join(cols, ","), table))
		} else {
			sqlBuff.WriteString(fmt.Sprintf("select * from %s where 1 = 1", table))
		}

		for i := 0; i < len(where); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
		}
		if wheresLen > 0 {
			sqlBuff.WriteString(" union ")
		}
	}

	if order != nil {
		// append order by clause for ordering
		sqlBuff.WriteString(fmt.Sprintf(" order by %s ", order.Feilds[0]))
		for i := 1; i < len(order.Feilds); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" , %s ", order.Feilds[i]))
		}
		sqlBuff.WriteString(order.Type)

		sqlBuff.WriteString(fmt.Sprintf(" limit %d offset %d ", paging.Limit, paging.CursorValue))
	}
	dbo.logger.Debug("SelectRowsUnionToMaps", zap.String("sql", sqlBuff.String()), zap.Any("values", values))

	rows, err := dbo.conn.QueryxContext(ctx, sqlBuff.String(), values...)
	if err != nil {
		return nil, err
	}

	var result = make([]map[string]interface{}, 0)
	for rows.Next() {
		var dest = make(map[string]interface{})
		if err = rows.MapScan(dest); err != nil {
			return nil, err
		}
		result = append(result, dest)
	}
	return result, nil
}

// SelectRows select rows to struct slice
func (dbo *DBO) SelectRows(ctx context.Context, table string, cols []string, where []Where, order *Order, paging *Paging, result interface{}) error {
	if table == "" {
		return errors.New("table name is required")
	}
	if len(where) == 0 {
		return errors.New("full-table-select is not allowed")
	}
	if order != nil && (len(order.Feilds) == 0 || order.Type == "") {
		return errors.New("order type and fields is required")
	}

	values := make([]interface{}, len(where))
	for i, v := range where {
		values[i] = v.Value
	}

	var sqlBuff bytes.Buffer
	if len(cols) > 0 {
		sqlBuff.WriteString(fmt.Sprintf("select %s from %s where 1 = 1", strings.Join(cols, ","), table))
	} else {
		sqlBuff.WriteString(fmt.Sprintf("select * from %s where 1 = 1", table))
	}
	for i := 0; i < len(where); i++ {
		sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
	}
	if order != nil {
		// append where clause for paging
		//		if paging != nil && paging.CursorValue != 0 {
		//			sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", paging.CursorName, order.GetOp()))
		//			values = append(values, paging.CursorValue)
		//		}

		// append order by clause for ordering
		sqlBuff.WriteString(fmt.Sprintf(" order by %s ", order.Feilds[0]))
		for i := 1; i < len(order.Feilds); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" , %s ", order.Feilds[i]))
		}
		sqlBuff.WriteString(order.Type)

		// append limit clause for paging
		if paging != nil {
			//sqlBuff.WriteString(" limit ? ")
			//values = append(values, paging.Limit)
			sqlBuff.WriteString(fmt.Sprintf(" limit %d offset %d ", paging.Limit, paging.CursorValue))
		}
	}

	//dbo.logger.Debug("SelectRows", zap.String("sql", sqlBuff.String()), zap.Any("values", values))
	return dbo.conn.SelectContext(ctx, result, sqlBuff.String(), values...)
}

// SelectRowsToMaps select rows to map slice
func (dbo *DBO) SelectRowsToMaps(ctx context.Context, table string, cols []string, where []Where, order *Order, paging *Paging) ([]map[string]interface{}, error) {
	if table == "" {
		return nil, errors.New("table name is required")
	}
	if len(where) == 0 {
		return nil, errors.New("full-table-select is not allowed")
	}
	if order != nil && (len(order.Feilds) == 0 || order.Type == "") {
		return nil, errors.New("order type and fields is required")
	}

	values := make([]interface{}, len(where))
	for i, v := range where {
		values[i] = v.Value
	}

	var sqlBuff bytes.Buffer
	if len(cols) > 0 {
		sqlBuff.WriteString(fmt.Sprintf("select %s from %s where 1 = 1", strings.Join(cols, ","), table))
	} else {
		sqlBuff.WriteString(fmt.Sprintf("select * from %s where 1 = 1", table))
	}
	for i := 0; i < len(where); i++ {
		sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
	}
	if order != nil {
		// append where clause for paging
		//		if paging != nil && paging.CursorValue != 0 {
		//			sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", paging.CursorName, order.GetOp()))
		//			values = append(values, paging.CursorValue)
		//		}

		// append order by clause for ordering
		sqlBuff.WriteString(fmt.Sprintf(" order by %s ", order.Feilds[0]))
		for i := 1; i < len(order.Feilds); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" , %s ", order.Feilds[i]))
		}
		sqlBuff.WriteString(order.Type)

		// append limit clause for paging
		if paging != nil {
			//sqlBuff.WriteString(" limit ? ")
			//values = append(values, paging.Limit)
			sqlBuff.WriteString(fmt.Sprintf(" limit %d offset %d ", paging.Limit, paging.CursorValue))
		}
	}

	//dbo.logger.Debug("SelectRowsToMaps", zap.String("sql", sqlBuff.String()), zap.Any("values", values))

	rows, err := dbo.conn.QueryxContext(ctx, sqlBuff.String(), values...)
	if err != nil {
		return nil, err
	}

	var result = make([]map[string]interface{}, 0)
	for rows.Next() {
		var dest = make(map[string]interface{})
		if err = rows.MapScan(dest); err != nil {
			return nil, err
		}
		result = append(result, dest)
	}
	return result, nil
}

// SelectRowsOffset select rows to struct slice
func (dbo *DBO) SelectRowsOffset(ctx context.Context, table string, cols []string, where []Where, order *Order, offset, limit uint64, result interface{}) error {
	if table == "" {
		return errors.New("table name is required")
	}
	if len(where) == 0 {
		return errors.New("full-table-select is not allowed")
	}
	if order != nil && (len(order.Feilds) == 0 || order.Type == "") {
		return errors.New("order type and fields is required")
	}

	values := make([]interface{}, len(where))
	for i, v := range where {
		values[i] = v.Value
	}

	var sqlBuff bytes.Buffer
	if len(cols) > 0 {
		sqlBuff.WriteString(fmt.Sprintf("select %s from %s where 1 = 1", strings.Join(cols, ","), table))
	} else {
		sqlBuff.WriteString(fmt.Sprintf("select * from %s where 1 = 1", table))
	}
	for i := 0; i < len(where); i++ {
		sqlBuff.WriteString(fmt.Sprintf(" and %s %s ? ", where[i].Name, where[i].GetOp()))
	}
	if order != nil {
		// append order by clause for ordering
		sqlBuff.WriteString(fmt.Sprintf(" order by %s ", order.Feilds[0]))
		for i := 1; i < len(order.Feilds); i++ {
			sqlBuff.WriteString(fmt.Sprintf(" , %s ", order.Feilds[i]))
		}
		sqlBuff.WriteString(order.Type)

		// append limit clause for paging
		sqlBuff.WriteString(fmt.Sprintf(" limit %d offset %d ", limit, offset))
	}

	dbo.logger.Debug("SelectRowsOffset", zap.String("sql", sqlBuff.String()), zap.Any("values", values))

	// execute
	return dbo.conn.SelectContext(ctx, result, sqlBuff.String(), values...)
}

// SelectRawSQL query useing raw sql
func (dbo *DBO) SelectRawSQL(ctx context.Context, table string, sqlStr string, values []interface{}, result interface{}) error {
	if table == "" {
		return errors.New("table name is required")
	}
	dbo.logger.Debug("selectRawSQL", zap.String("sql", sqlStr), zap.Any("values", values))
	return dbo.conn.SelectContext(ctx, result, sqlStr, values...)
}

func (dbo *DBO) Excute(ctx context.Context, stmt *sql.Stmt, fields []Feild) (sql.Result, error) {
	values := make([]interface{}, len(fields))
	for i, v := range fields {
		values[i] = v.Value
	}
	return stmt.ExecContext(ctx, values...)
}

func (dbo *DBO) Prepare(ctx context.Context, tx *sql.Tx, table string, fields []Feild) (*sql.Stmt, error) {
	var sqlBuff bytes.Buffer
	sqlBuff.WriteString(fmt.Sprintf("insert into %s (", table))
	for i := 0; i < len(fields)-1; i++ {
		sqlBuff.WriteString(fmt.Sprintf("%s,", fields[i].Name))
	}
	sqlBuff.WriteString(fmt.Sprintf("%s) values (", fields[len(fields)-1].Name))

	for i := 1; i < len(fields); i++ {
		sqlBuff.WriteString(fmt.Sprintf("?,"))
	}
	sqlBuff.WriteString(fmt.Sprintf("?);"))
	dbo.logger.Debug("Prepare", zap.String("sql", sqlBuff.String()))
	if tx != nil {
		return tx.PrepareContext(ctx, sqlBuff.String())
	} else {
		return dbo.conn.PrepareContext(ctx, sqlBuff.String())
	}
}
