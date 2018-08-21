package gobatis

import (
	"errors"
	"reflect"
	"strings"
	"text/template"
)

type templateSQL struct {
	sqlTemplate *template.Template
}

func (stmt *templateSQL) GenerateSQL(ctx *Context, paramNames []string, paramValues []interface{}) (string, []interface{}, error) {
	var tplArgs interface{}
	if len(paramNames) == 0 {
		if len(paramValues) == 0 {
			return "", nil, errors.New("arguments is missing")
		}
		if len(paramValues) > 1 {
			return "", nil, errors.New("arguments is exceed 1")
		}

		tplArgs = paramValues[0]
	} else if len(paramNames) == 1 {
		tplArgs = paramValues[0]
		if _, ok := tplArgs.(map[string]interface{}); !ok {
			paramType := reflect.TypeOf(tplArgs)
			if paramType.Kind() == reflect.Ptr {
				paramType = paramType.Elem()
			}
			if paramType.Kind() != reflect.Struct {
				tplArgs = map[string]interface{}{paramNames[0]: paramValues[0]}
			}
		}
	} else {
		var args = map[string]interface{}{}
		for idx := range paramNames {
			args[paramNames[idx]] = paramValues[idx]
		}
		tplArgs = args
	}

	var sb strings.Builder
	err := stmt.sqlTemplate.Execute(&sb, tplArgs)
	if err != nil {
		return "", nil, err
	}
	sql := sb.String()

	fragments, nameArgs, err := compileNamedQuery(sql)
	if err != nil {
		return "", nil, err
	}
	if len(nameArgs) == 0 {
		return sql, nil, nil
	}
	sql = ctx.Dialect.Placeholder().Concat(fragments, nameArgs)
	sqlParams, err := bindNamedQuery(nameArgs, paramNames, paramValues, ctx.Dialect, ctx.Mapper)
	return sql, sqlParams, err
}

type sqlWithParams struct {
	dollarSQL  string
	questSQL   string
	bindParams Params
}

func (stmt *sqlWithParams) WithQuestion() string {
	return stmt.questSQL
}

func (stmt *sqlWithParams) WithDollar() string {
	return stmt.dollarSQL
}

func (stmt *sqlWithParams) GenerateSQL(ctx *Context, paramNames []string, paramValues []interface{}) (string, []interface{}, error) {
	sql := ctx.Dialect.Placeholder().Get(stmt)
	sqlParams, err := bindNamedQuery(stmt.bindParams, paramNames, paramValues, ctx.Dialect, ctx.Mapper)
	return sql, sqlParams, err
}
