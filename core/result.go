package core

import (
	"context"
	"database/sql"
	"errors"
)

type SingleRowResult = Result

type Result struct {
	ctx       context.Context // nolint
	o         *connection
	tx        DBRunner
	id        string
	sql       string
	sqlParams []interface{}
	err       error
}

func (result SingleRowResult) SQL() (string, []interface{}, error) {
	return result.sql, result.sqlParams, result.err
}

func (result SingleRowResult) Scan(value interface{}) error {
	return result.RowScan(func(r *sql.Rows) error {
		return scanAny(result.o.dialect, result.o.mapper, r, value, false, result.o.isUnsafe)
	})
}

func (result SingleRowResult) RowScan(cb func(*sql.Rows) error) error {
	if result.err != nil {
		return result.err
	}

	if result.tx == nil {
		result.tx = TxFromContext(result.ctx)
		if result.tx == nil {
			result.tx = result.o.db
		}
	}

	rows, err := result.tx.QueryContext(result.ctx, result.sql, result.sqlParams...)
	result.o.tracer.Write(result.ctx, result.id, result.sql, result.sqlParams, err)
	if err != nil {
		return result.o.dialect.HandleError(err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return result.o.dialect.HandleError(err)
		}
		return sql.ErrNoRows
	}

	return cb(rows)
}

func (result SingleRowResult) ScanMultiple(multiple *Multiple) error {
	return result.RowScan(func(r *sql.Rows) error {
		return multiple.Scan(result.o.dialect, result.o.mapper, r, result.o.isUnsafe)
	})
}

type MultRowResult struct {
	ctx       context.Context // nolint
	o         *connection
	tx        DBRunner
	id        string
	sql       string
	sqlParams []interface{}
	rows      *sql.Rows
	err       error
}

func (result *MultRowResult) SQL() (string, []interface{}, error) {
	return result.sql, result.sqlParams, result.err
}

func (results *MultRowResult) Close() error {
	if results.rows != nil {
		return results.rows.Close()
	}
	return nil
}

func (results *MultRowResult) Err() error {
	return results.err
}

func (results *MultRowResult) Next() bool {
	if results.err != nil {
		return false
	}

	if results.rows == nil {
		if results.tx == nil {
			results.tx = TxFromContext(results.ctx)
			if results.tx == nil {
				results.tx = results.o.db
			}
		}

		results.rows, results.err = results.tx.QueryContext(results.ctx, results.sql, results.sqlParams...)
		results.o.tracer.Write(results.ctx, results.id, results.sql, results.sqlParams, results.err)

		if results.err != nil {
			results.err = results.o.dialect.HandleError(results.err)
			return false
		}
	}

	return results.rows.Next()
}

func (results *MultRowResult) Rows() *sql.Rows {
	return results.rows
}

func (results *MultRowResult) Scan(value interface{}) error {
	if results.err != nil {
		return results.err
	}

	if results.rows == nil {
		return errors.New("please first invoke Next()")
	}
	return scanAny(results.o.dialect, results.o.mapper, results.rows, value, false, results.o.isUnsafe)
}

func (results *MultRowResult) ScanSlice(value interface{}) error {
	return results.scanAll(func(r rowsi) error {
		return ScanAll(results.o.dialect, results.o.mapper, r, value, false, results.o.isUnsafe)
	})
}

func (results *MultRowResult) ScanResults(value interface{}) error {
	return results.scanAll(func(r rowsi) error {
		return ScanAll(results.o.dialect, results.o.mapper, r, value, false, results.o.isUnsafe)
	})
}

func (results *MultRowResult) scanAll(cb func(rowsi) error) error {
	if results.err != nil {
		return results.err
	}

	if results.rows != nil {
		return errors.New("please not invoke Next()")
	}

	if results.tx == nil {
		results.tx = TxFromContext(results.ctx)
		if results.tx == nil {
			results.tx = results.o.db
		}
	}

	rows, err := results.tx.QueryContext(results.ctx, results.sql, results.sqlParams...)
	results.o.tracer.Write(results.ctx, results.id, results.sql, results.sqlParams, err)
	if err != nil {
		return results.o.dialect.HandleError(err)
	}
	defer rows.Close()

	err = cb(rows)
	if err != nil {
		return err
	}

	return rows.Close()
}

func (results *MultRowResult) ScanBasicMap(value interface{}) error {
	return results.scanAll(func(r rowsi) error {
		return scanBasicMap(results.o.dialect, results.o.mapper, r, value)
	})
}

func (results *MultRowResult) ScanMultipleArray(multipleArray *MultipleArray) error {
	return results.scanAll(func(r rowsi) error {
		return multipleArray.Scan(results.o.dialect, results.o.mapper, r, results.o.isUnsafe)
	})
}
