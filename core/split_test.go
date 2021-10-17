package core

import (
	"strings"
	"testing"
)

func TestSemicolons(t *testing.T) {
	type testData struct {
		line   string
		result bool
	}

	tests := []testData{
		{
			line:   "END;",
			result: true,
		},
		{
			line:   "END; -- comment",
			result: true,
		},
		{
			line:   "END   ; -- comment",
			result: true,
		},
		{
			line:   "END -- comment",
			result: false,
		},
		{
			line:   "END -- comment ;",
			result: false,
		},
		{
			line:   "END \" ; \" -- comment",
			result: false,
		},
	}

	for _, test := range tests {
		r := endsWithSemicolon(test.line)
		if r != test.result {
			t.Errorf("incorrect semicolon. got %v, want %v", r, test.result)
		}
	}
}

func TestSplitStatements(t *testing.T) {
	type testData struct {
		sql       string
		direction bool
		count     int
	}

	tests := []testData{
		{
			sql: `
CREATE TABLE IF NOT EXISTS histories (
  id                BIGSERIAL  PRIMARY KEY,
  current_value     varchar(2000) NOT NULL,
  created_at      timestamp with time zone  NOT NULL
);

-- +gobatis StatementBegin
CREATE OR REPLACE FUNCTION histories_partition_creation( DATE, DATE )
returns void AS $$
DECLARE
  create_query text;
BEGIN
  FOR create_query IN SELECT
      'CREATE TABLE IF NOT EXISTS histories_'
      || TO_CHAR( d, 'YYYY_MM' )
      || ' ( CHECK( created_at >= timestamp '''
      || TO_CHAR( d, 'YYYY-MM-DD 00:00:00' )
      || ''' AND created_at < timestamp '''
      || TO_CHAR( d + INTERVAL '1 month', 'YYYY-MM-DD 00:00:00' )
      || ''' ) ) inherits ( histories );'
    FROM generate_series( $1, $2, '1 month' ) AS d
  LOOP
    EXECUTE create_query;
  END LOOP;  -- LOOP END
END;         -- FUNCTION END
$$
language plpgsql;
-- +gobatis StatementEnd

`,
			count: 2,
		},
		{
			sql: `SELECT 1;
      SELECT 1;`,
			count: 2,
		},
		{
			sql: `SELECT 1; -- a
      SELECT 1; -- b`,
			count: 2,
		},
		{
			sql: `SELECT 1; -- a
       -- b`,
			count: 1,
		},
		{
			sql:   `SELECT 1;`,
			count: 1,
		},
		{
			sql:   `SELECT 1`,
			count: 1,
		},
	}

	for _, test := range tests {
		stmts := splitSQLStatements(strings.NewReader(test.sql))
		if len(stmts) != test.count {
			t.Errorf("incorrect number of stmts. got %v, want %v", len(stmts), test.count)
		}
	}
}
