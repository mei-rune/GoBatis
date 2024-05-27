package core

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

type (
	StatementType int
	ResultType    int
)

func (t StatementType) String() string {
	if int(t) < 0 || int(t) > len(statementTypeNames) {
		return ""
	}
	return statementTypeNames[int(t)]
}

func (t ResultType) String() string {
	if int(t) < 0 || int(t) > len(resultTypeNames) {
		return ""
	}
	return resultTypeNames[int(t)]
}

const (
	StatementTypeNone   StatementType = -1
	StatementTypeSelect StatementType = 0
	StatementTypeUpdate StatementType = 1
	StatementTypeInsert StatementType = 2
	StatementTypeDelete StatementType = 3

	ResultUnknown ResultType = 0
	ResultMap     ResultType = 1
	ResultStruct  ResultType = 2
)

var (
	statementTypeNames = [...]string{
		"select",
		"update",
		"insert",
		"delete",
	}

	resultTypeNames = [...]string{
		"unknown",
		"map",
		"struct",
	}
)

type Param struct {
	Name    string
	Type    string
	Null    sql.NullBool
	NotNull sql.NullBool
}

type Params []Param

func (params Params) toNames() []string {
	names := make([]string, len(params))
	for idx := range params {
		names[idx] = params[idx].Name
	}
	return names
}

func parseParam(s string) (Param, error) {
	ss := strings.Split(s, ",")
	if len(ss) == 0 {
		return Param{Name: s}, errors.New("param '" + s + "' is syntex error")
	}
	if len(ss) == 1 {
		return Param{Name: ss[0]}, nil
	}
	param := Param{Name: ss[0]}
	for _, a := range ss[1:] {
		kv := strings.SplitN(a, "=", 2)
		var key, value string
		if len(kv) == 1 {
			key = kv[0]
		} else if len(kv) == 2 {
			key = kv[0]
			value = kv[1]
		}

		if key == "" {
			continue
		}

		value = strings.ToLower(strings.TrimSpace(value))
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "type":
			param.Type = value
		case "null":
			param.Null.Valid = true
			param.Null.Bool = value == "true"
		case "notnull":
			param.NotNull.Valid = true
			param.NotNull.Bool = value == "true"
		default:
			return Param{Name: s}, errors.New("param '" + s + "' is syntex error - " + key + " is unsupported")
		}
	}
	return param, nil
}

type sqlAndParam struct {
	SQL    string
	Params []interface{}
}

type DynamicSQL interface {
	GenerateSQL(*Context) (string, []interface{}, error)
}

type MappedStatement struct {
	id          string
	source      string // xml or generate
	sqlType     StatementType
	result      ResultType
	rawSQL      string
	dynamicSQLs []DynamicSQL
}

func (stmt *MappedStatement) IsGenerated() bool {
	return stmt.source != "xml"
}

func (stmt *MappedStatement) SQLStrings() []string {
	if len(stmt.dynamicSQLs) == 0 {
		return []string{fmt.Sprint(stmt.dynamicSQLs[0])}
	}
	res := make([]string, len(stmt.dynamicSQLs))
	for idx := range stmt.dynamicSQLs {
		res[idx] = fmt.Sprint(stmt.dynamicSQLs[idx])
	}
	return res
}

func (stmt *MappedStatement) GenerateSQLs(ctx *Context) ([]sqlAndParam, error) {
	sqlAndParams := make([]sqlAndParam, len(stmt.dynamicSQLs))
	for idx := range stmt.dynamicSQLs {
		sql, params, err := stmt.dynamicSQLs[idx].GenerateSQL(ctx)
		if err != nil {
			return nil, err
		}
		sqlAndParams[idx].SQL = sql
		sqlAndParams[idx].Params = params
	}
	return sqlAndParams, nil
}

func NewMapppedStatement(ctx *StmtContext, id string, statementType StatementType, resultType ResultType, sqlStr string) (*MappedStatement, error) {
	if ctx.FindSqlFragment == nil {
		sqlExpressions := ctx.InitContext.SqlExpressions
		ctx.FindSqlFragment = func(id string) (SqlExpression, error) {
			if sqlExpressions != nil {
				sf := sqlExpressions[id]
				if sf != nil {
					return sf, nil
				}
			}
			return nil, errors.New("sql '" + id + "' missing")
		}
	}

	stmt := &MappedStatement{
		id:      id,
		sqlType: statementType,
		result:  resultType,
	}

	stmt.rawSQL = sqlStr

	if ctx.Config.EnabledSQLCheck && strings.Contains(sqlStr, "${") {
		fmt.Println("WARN: sql statement contains ${}, replace it with #{}?") // nolint
	}

	sqlList := splitSQLStatements(strings.NewReader(sqlStr))
	for idx := range sqlList {
		sql, err := CreateSQL(ctx, id, sqlList[idx], sqlStr, len(sqlList) == 1)
		if err != nil {
			return nil, err
		}

		stmt.dynamicSQLs = append(stmt.dynamicSQLs, sql)
	}
	return stmt, nil
}

func CreateSQL(ctx *StmtContext, id, sqlStr, fullText string, one bool) (DynamicSQL, error) {
	if strings.Contains(sqlStr, "{{") {
		funcMap := ctx.Config.TemplateFuncs
		tpl, err := template.New(id).Funcs(funcMap).Parse(sqlStr)
		if err != nil {
			return nil, errors.New("sql is invalid go template of '" + id + "', " + err.Error() + "\r\n\t" + sqlStr)
		}

		if hasXMLTag(sqlStr) {
			return nil, errors.New("sql is invalid go template of '" + id + "', becase xml tag is exists in:\r\n\t" + sqlStr)
		}

		return &templateSQL{sqlTemplate: tpl}, nil
	}

	// http://www.mybatis.org/mybatis-3/dynamic-sql.html
	if hasXMLTag(sqlStr) {
		dynamicSQL, err := loadDynamicSQLFromXML(ctx, sqlStr)
		if err != nil {
			return nil, errors.New("sql is invalid dynamic sql of '" + id + "', " + err.Error() + "\r\n\t" + sqlStr)
		}
		return dynamicSQL, nil
	}

	fragments, bindParams, err := CompileNamedQuery(sqlStr)
	if err != nil {
		return nil, errors.New("sql is invalid named sql of '" + id + "', " + err.Error())
	}
	if len(bindParams) != 0 {
		return &parameterizedSQL{
			rawSQL:     sqlStr,
			dollarSQL:  Placeholders(Dollar, fragments, bindParams, 0),
			questSQL:   Placeholders(Question, fragments, bindParams, 0),
			bindParams: bindParams,
		}, nil
	}

	if !one {
		return rawSQL(sqlStr), nil
	}
	return allParamsSQL(sqlStr), nil
}

func NewSqlExpression(ctx *InitContext, sqlstr string) (SqlExpression, error) {
	stmtctx := &StmtContext{
		InitContext: ctx,
	}
	stmtctx.FindSqlFragment = func(id string) (SqlExpression, error) {
		if stmtctx.InitContext.SqlExpressions != nil {
			sf := stmtctx.InitContext.SqlExpressions[id]
			if sf != nil {
				return sf, nil
			}
		}
		return nil, errors.New("sql '" + id + "' missing")
	}
	segements, err := readSQLStatementForXML(stmtctx, sqlstr)
	if err != nil {
		return nil, err
	}
	return expressionArray(segements), nil
}

func CompileNamedQuery(txt string) ([]string, Params, error) {
	idx := strings.Index(txt, "#{")
	if idx < 0 {
		return []string{txt}, nil, nil
	}

	s := txt
	seekPos := 0
	var fragments []string
	var argments Params
	for {
		seekPos += idx
		fragments = append(fragments, s[:idx])
		s = s[idx+len("#{"):]
		end := strings.Index(s, "}")
		if end < 0 {
			return nil, nil, errors.New(MarkSQLError(txt, seekPos))
		}
		param, err := parseParam(s[:end])
		if err != nil {
			return nil, nil, err
		}
		argments = append(argments, param)

		s = s[end+len("}"):]

		seekPos += len("#{}")
		seekPos += end

		idx = strings.Index(s, "#{")
		if idx < 0 {
			fragments = append(fragments, s)
			return fragments, argments, nil
		}
	}
}

func bindNamedQuery(bindParams Params, ctx *Context) ([]interface{}, error) {
	if len(bindParams) == 0 {
		return nil, nil
	}

	bindValues := make([]interface{}, len(bindParams))
	for idx := range bindParams {
		sqlValue, err := ctx.RValue(&bindParams[idx])
		if err != nil {
			if err == ErrNotFound {
				return nil, errors.New("param '" + bindParams[idx].Name + "' is missing")
			}
			return nil, err
		}

		bindValues[idx] = sqlValue
	}
	return bindValues, nil
}

func MarkSQLError(sql string, index int) string {
	result := fmt.Sprintf("%s[****ERROR****]->%s", sql[0:index], sql[index:])
	return result
}

type templateSQL struct {
	sqlTemplate *template.Template
}

func (stmt *templateSQL) GenerateSQL(ctx *Context) (string, []interface{}, error) {
	var tplArgs interface{}
	if len(ctx.ParamNames) == 0 {
		if len(ctx.ParamValues) == 0 {
			return "", nil, errors.New("arguments is missing")
		}
		if len(ctx.ParamValues) > 1 {
			return "", nil, errors.New("arguments is exceed 1")
		}

		tplArgs = ctx.ParamValues[0]
	} else if len(ctx.ParamNames) == 1 {
		tplArgs = ctx.ParamValues[0]
		if _, ok := tplArgs.(map[string]interface{}); !ok {
			paramType := reflect.TypeOf(tplArgs)
			if paramType.Kind() == reflect.Ptr {
				paramType = paramType.Elem()
			}
			if paramType.Kind() != reflect.Struct {
				tplArgs = map[string]interface{}{ctx.ParamNames[0]: ctx.ParamValues[0]}
			}
		}
	} else {
		args := map[string]interface{}{}
		for idx := range ctx.ParamNames {
			args[ctx.ParamNames[idx]] = ctx.ParamValues[idx]
		}
		tplArgs = args
	}

	var sb strings.Builder
	err := stmt.sqlTemplate.Execute(&sb, tplArgs)
	if err != nil {
		return "", nil, err
	}
	sql := sb.String()

	fragments, nameArgs, err := CompileNamedQuery(sql)
	if err != nil {
		return "", nil, err
	}
	if len(nameArgs) == 0 {
		return sql, nil, nil
	}
	sql = Placeholders(ctx.Dialect.Placeholder(), fragments, nameArgs, 0)
	sqlParams, err := bindNamedQuery(nameArgs, ctx)
	return sql, sqlParams, err
}

type parameterizedSQL struct {
	rawSQL     string
	dollarSQL  string
	questSQL   string
	bindParams Params
}

func (stmt *parameterizedSQL) WithQuestion() string {
	return stmt.questSQL
}

func (stmt *parameterizedSQL) WithDollar() string {
	return stmt.dollarSQL
}

func (stmt *parameterizedSQL) String() string {
	return stmt.rawSQL
}

func (stmt *parameterizedSQL) GenerateSQL(ctx *Context) (string, []interface{}, error) {
	sql := ctx.Dialect.Placeholder().Print(stmt)
	sqlParams, err := bindNamedQuery(stmt.bindParams, ctx)
	return sql, sqlParams, err
}

type rawSQL string

func (sql rawSQL) GenerateSQL(ctx *Context) (string, []interface{}, error) {
	return string(sql), nil, nil
}

type allParamsSQL string

func (sql allParamsSQL) GenerateSQL(ctx *Context) (string, []interface{}, error) {
	return string(sql), ctx.ParamValues, nil
}
