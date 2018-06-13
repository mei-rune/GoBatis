package gobatis

import (
	"database/sql"
	"errors"
)

type Result struct {
	o         *Connection
	id        string
	sql       string
	sqlParams []interface{}
	err       error
}

func (result Result) Scan(value interface{}) error {
	if result.err != nil {
		return result.err
	}

	if result.o.showSQL {
		result.o.logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, result.id, result.sql, result.sqlParams)
	}

	rows, err := result.o.db.Query(result.sql, result.sqlParams...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	return scanAny(result.o.dbType, result.o.mapper, rows, value, false, result.o.isUnsafe)
}

type Results struct {
	o         *Connection
	id        string
	sql       string
	sqlParams []interface{}
	rows      *sql.Rows
	err       error
}

func (results *Results) Close() error {
	if results.rows != nil {
		return results.rows.Close()
	}
	return nil
}

func (results *Results) Err() error {
	return results.err
}

func (results *Results) Next() bool {
	if results.err != nil {
		return false
	}

	if results.rows == nil {
		if results.o.showSQL {
			results.o.logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, results.id, results.sql, results.sqlParams)
		}

		results.rows, results.err = results.o.db.Query(results.sql, results.sqlParams...)
		if results.err != nil {
			return false
		}
	}

	return results.rows.Next()
}

func (results *Results) Scan(value interface{}) error {
	if results.err != nil {
		return results.err
	}

	if results.rows == nil {
		return errors.New("please first invoke Next()")
	}
	return scanAny(results.o.dbType, results.o.mapper, results.rows, value, false, results.o.isUnsafe)
}

func (results *Results) ScanSlice(value interface{}) error {
	return results.scanAll(value)
}

func (results *Results) ScanResults(value interface{}) error {
	return results.scanAll(value)
}

func (results *Results) scanAll(value interface{}) error {
	if results.err != nil {
		return results.err
	}

	if results.rows != nil {
		return errors.New("please not invoke Next()")
	}

	if results.o.showSQL {
		results.o.logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, results.id, results.sql, results.sqlParams)
	}

	rows, err := results.o.db.Query(results.sql, results.sqlParams...)
	if err != nil {
		return err
	}
	defer rows.Close()

	err = scanAll(results.o.dbType, results.o.mapper, rows, value, false, results.o.isUnsafe)
	if err != nil {
		return err
	}

	return rows.Close()
}

func (results *Results) ScanBasicMap(value interface{}) error {
	if results.err != nil {
		return results.err
	}

	if results.rows != nil {
		return errors.New("please not invoke Next()")
	}

	if results.o.showSQL {
		results.o.logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, results.id, results.sql, results.sqlParams)
	}

	rows, err := results.o.db.Query(results.sql, results.sqlParams...)
	if err != nil {
		return err
	}
	defer rows.Close()

	err = scanBasicMap(results.o.dbType, results.o.mapper, rows, value)
	if err != nil {
		return err
	}

	return rows.Close()
}
