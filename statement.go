package gobatis

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"unicode"
)

type StatementType int
type ResultType int

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
	Name string
	Type string
}

type Params []Param

func (params Params) toNames() []string {
	names := make([]string, len(params))
	for idx := range params {
		names[idx] = params[idx].Name
	}
	return names
}

type MappedStatement struct {
	id         string
	sqlType    StatementType
	result     ResultType
	rawSql     string
	dynamicSQL DynamicSQL
}

type DynamicSQL interface {
	GenerateSQL(*Context) (string, []interface{}, error)
}

func (stmt *MappedStatement) GenerateSQL(ctx *Context) (sql string, sqlParams []interface{}, err error) {
	if stmt.dynamicSQL == nil {
		return stmt.rawSql, ctx.ParamValues, nil
	}
	return stmt.dynamicSQL.GenerateSQL(ctx)
}

func NewMapppedStatement(ctx *InitContext, id string, statementType StatementType, resultType ResultType, sqlStr string) (*MappedStatement, error) {
	stmt := &MappedStatement{
		id:      id,
		sqlType: statementType,
		result:  resultType,
	}

	sqlTemp := strings.Replace(strings.TrimSpace(sqlStr), "\r\n", " ", -1)
	sqlTemp = strings.Replace(sqlTemp, "\n", " ", -1)
	sqlTemp = strings.Replace(sqlTemp, "\t", " ", -1)
	for strings.Contains(sqlTemp, "  ") {
		sqlTemp = strings.Replace(sqlTemp, "  ", " ", -1)
	}
	sqlTemp = strings.TrimSpace(sqlTemp)
	stmt.rawSql = sqlTemp

	if strings.Contains(sqlTemp, "{{") {
		funcMap := ctx.Config.TemplateFuncs
		tpl, err := template.New(id).Funcs(funcMap).Parse(sqlTemp)
		if err != nil {
			return nil, errors.New("sql is invalid go template of '" + id + "', " + err.Error() + "\r\n\t" + sqlTemp)
		}
		stmt.dynamicSQL = &templateSQL{sqlTemplate: tpl}
		return stmt, nil
	}

	// http://www.mybatis.org/mybatis-3/dynamic-sql.html
	for _, tag := range []string{"<where>", "<set>", "<chose>"} {
		if strings.Contains(sqlTemp, tag) {
			dynamicSQL, err := loadDynamicSQLFromXML(sqlTemp)
			if err != nil {
				return nil, errors.New("sql is invalid dynamic sql of '" + id + "', " + err.Error() + "\r\n\t" + sqlTemp)
			}

			stmt.dynamicSQL = dynamicSQL
			return stmt, nil
		}
	}
	for _, tag := range []string{"<if", "<foreach"} {
		idx := strings.Index(sqlTemp, tag)
		exceptIndex := idx + len(tag) + 1
		if idx >= 0 && len(sqlTemp) > exceptIndex && unicode.IsSpace(rune(sqlTemp[exceptIndex])) {
			dynamicSQL, err := loadDynamicSQLFromXML(sqlTemp)
			if err != nil {
				return nil, errors.New("sql is invalid dynamic sql of '" + id + "', " + err.Error() + "\r\n\t" + sqlTemp)
			}

			stmt.dynamicSQL = dynamicSQL
			return stmt, nil
		}
	}

	if strings.Contains(sqlTemp, "${") {
		ctx.Logger.Println("WARN: sql statement contains ${}, replace it with #{}?")
	}

	fragments, bindParams, err := compileNamedQuery(sqlTemp)
	if err != nil {
		return nil, errors.New("sql is invalid named sql of '" + id + "', " + err.Error())
	}
	if len(bindParams) != 0 {
		stmt.dynamicSQL = &sqlWithParams{
			rawSQL:     sqlTemp,
			dollarSQL:  Dollar.Concat(fragments, bindParams),
			questSQL:   Question.Concat(fragments, bindParams),
			bindParams: bindParams,
		}
	}
	return stmt, nil
}

func compileNamedQuery(txt string) ([]string, Params, error) {
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
		default:
			return Param{Name: s}, errors.New("param '" + s + "' is syntex error - " + key + " is unsupported")
		}
	}
	return param, nil
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
		var args = map[string]interface{}{}
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

	fragments, nameArgs, err := compileNamedQuery(sql)
	if err != nil {
		return "", nil, err
	}
	if len(nameArgs) == 0 {
		return sql, nil, nil
	}
	sql = ctx.Dialect.Placeholder().Concat(fragments, nameArgs)
	sqlParams, err := bindNamedQuery(nameArgs, ctx)
	return sql, sqlParams, err
}

type sqlWithParams struct {
	rawSQL     string
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

func (stmt *sqlWithParams) String() string {
	return stmt.rawSQL
}

func (stmt *sqlWithParams) GenerateSQL(ctx *Context) (string, []interface{}, error) {
	sql := ctx.Dialect.Placeholder().Get(stmt)
	sqlParams, err := bindNamedQuery(stmt.bindParams, ctx)
	return sql, sqlParams, err
}

func (stmt *sqlWithParams) writeTo(printer *sqlPrinter) {
	sql := printer.ctx.Dialect.Placeholder().Get(stmt)
	sqlParams, err := bindNamedQuery(stmt.bindParams, printer.ctx)
	if err != nil {
		printer.err = err
		return
	}
	printer.sb.WriteString(sql)
	if len(sqlParams) != 0 {
		printer.params = append(printer.params, sqlParams...)
	}
}
