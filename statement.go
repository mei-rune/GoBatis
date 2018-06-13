package gobatis

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"
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

type SQLWithParams struct {
	dollarSQL  string
	questSQL   string
	bindParams Params
}

type MappedStatement struct {
	id          string
	sqlTemplate *template.Template
	sql         string
	sqlCompiled *SQLWithParams
	sqlType     StatementType
	result      ResultType
}

type stmtXML struct {
	ID     string `xml:"id,attr"`
	Result string `xml:"result,attr"`
	SQL    string `xml:",chardata"`
}

type xmlConfig struct {
	Selects []stmtXML `xml:"select"`
	Deletes []stmtXML `xml:"delete"`
	Updates []stmtXML `xml:"update"`
	Inserts []stmtXML `xml:"insert"`
}

func readMappedStatements(path string) ([]*MappedStatement, error) {
	statements := make([]*MappedStatement, 0)

	xmlFile, err := os.Open(path)
	if err != nil {
		return nil, errors.New("Error opening file: " + err.Error())
	}
	defer xmlFile.Close()

	xmlObj := xmlConfig{}
	decoder := xml.NewDecoder(xmlFile)
	if err = decoder.Decode(&xmlObj); err != nil {
		return nil, errors.New("Error decode file '" + path + "': " + err.Error())
	}

	for _, deleteStmt := range xmlObj.Deletes {
		mapper, err := newMapppedStatement(deleteStmt, StatementTypeDelete)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + deleteStmt.ID + "': " + err.Error())
		}
		statements = append(statements, mapper)
	}
	for _, insertStmt := range xmlObj.Inserts {
		mapper, err := newMapppedStatement(insertStmt, StatementTypeInsert)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + insertStmt.ID + "': " + err.Error())
		}
		statements = append(statements, mapper)
	}
	for _, selectStmt := range xmlObj.Selects {
		mapper, err := newMapppedStatement(selectStmt, StatementTypeSelect)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + selectStmt.ID + "': " + err.Error())
		}
		statements = append(statements, mapper)
	}
	for _, updateStmt := range xmlObj.Updates {
		mapper, err := newMapppedStatement(updateStmt, StatementTypeUpdate)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + updateStmt.ID + "': " + err.Error())
		}
		statements = append(statements, mapper)
	}
	return statements, nil
}

func newMapppedStatement(stmt stmtXML, sqlType StatementType) (*MappedStatement, error) {
	var resultType ResultType
	switch strings.ToLower(stmt.Result) {
	case "":
		resultType = ResultUnknown
	case "struct", "type", "resultstruct", "resulttype":
		resultType = ResultStruct
	case "map", "resultmap":
		// sqlMapperObj.result = resultMap
		return nil, errors.New("result '" + stmt.Result + "' of '" + stmt.ID + "' is unsupported")
	default:
		return nil, errors.New("result '" + stmt.Result + "' of '" + stmt.ID + "' is unsupported")
	}
	return NewMapppedStatement(stmt.ID, sqlType, resultType, stmt.SQL)
}

func NewMapppedStatement(id string, statementType StatementType, resultType ResultType, sqlStr string) (*MappedStatement, error) {
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
	stmt.sql = sqlTemp

	funcMap := template.FuncMap{
		"isLast": func(list interface{}, idx int) bool {
			if list == nil {
				return false
			}
			rValue := reflect.ValueOf(list)
			if rValue.Kind() != reflect.Slice {
				return false
			}
			return idx == (rValue.Len() - 1)
		},
	}

	if strings.Contains(sqlTemp, "{{") {
		tpl, err := template.New(id).Funcs(funcMap).Parse(sqlTemp)
		if err != nil {
			return nil, errors.New("sql is invalid go template of '" + id + "', " + err.Error() + "\r\n\t" + sqlTemp)
		}
		stmt.sqlTemplate = tpl
	} else {
		fragments, bindParams, err := compileNamedQuery(sqlTemp)
		if err != nil {
			return nil, errors.New("sql is invalid named sql of '" + id + "', " + err.Error())
		}
		if len(bindParams) != 0 {
			stmt.sqlCompiled = &SQLWithParams{dollarSQL: Dollar.Concat(fragments, bindParams),
				questSQL:   Question.Concat(fragments, bindParams),
				bindParams: bindParams}
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

func bindNamedQuery(bindParams Params, paramNames []string, paramValues []interface{}, dialect Dialect, mapper *Mapper) ([]interface{}, error) {
	if len(bindParams) == 0 {
		return nil, nil
	}

	if len(paramNames) == 0 {
		if len(paramValues) == 0 {
			return nil, errors.New("arguments is empty")
		}
		if len(paramValues) > 1 {
			return nil, errors.New("arguments is exceed 1")
		}

		if mapArgs, ok := paramValues[0].(map[string]interface{}); ok {
			return bindMapArgs(dialect, bindParams, mapArgs)
		}
		return bindStruct(bindParams, paramValues[0], dialect, mapper)
	}

	if len(bindParams) != len(paramValues) && len(paramValues) == 1 {
		if mapArgs, ok := paramValues[0].(map[string]interface{}); ok {
			return bindMapArgs(dialect, bindParams, mapArgs)
		}
		return bindStruct(bindParams, paramValues[0], dialect, mapper)
	}

	bindValues := make([]interface{}, len(bindParams))
	for idx := range bindParams {
		foundIndex := -1
		for nidx := range paramNames {
			if paramNames[nidx] == bindParams[idx].Name {
				foundIndex = nidx
				break
			}
		}
		if foundIndex >= 0 {
			sqlValue, err := toSQLType(dialect, &bindParams[idx], paramValues[foundIndex])
			if err != nil {
				return nil, err
			}
			bindValues[idx] = sqlValue
			continue
		}
		dotIndex := strings.IndexByte(bindParams[idx].Name, '.')
		if dotIndex >= 0 {
			variableName := bindParams[idx].Name[:dotIndex]
			for nidx := range paramNames {
				if paramNames[nidx] == variableName {
					foundIndex = nidx
					break
				}
			}
			if foundIndex >= 0 {
				if v := paramValues[foundIndex]; v != nil {
					rValue := reflect.ValueOf(v)
					tm := mapper.TypeMap(rValue.Type())
					name := bindParams[idx].Name[dotIndex+1:]
					fi, ok := tm.Names[name]
					if ok {
						fvalue, err := fi.LValue(dialect, &bindParams[idx], rValue)
						if err != nil {
							return nil, err
						}
						bindValues[idx] = fvalue
						continue
					}
				}
			}
		}

		sqlValue, err := toSQLType(dialect, &bindParams[idx], nil)
		if err != nil {
			return nil, err
		}
		bindValues[idx] = sqlValue
	}
	return bindValues, nil
}

// private interface to generate a list of interfaces from a given struct
// type, given a list of names to pull out of the struct.  Used by public
// BindStruct interface.
func bindStruct(params Params, arg interface{}, dialect Dialect, m *Mapper) ([]interface{}, error) {
	arglist := make([]interface{}, 0, len(params))

	// grab the indirected value of arg
	v := reflect.ValueOf(arg)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		if len(params) <= 1 {
			return []interface{}{arg}, nil
		}
		return arglist, fmt.Errorf("could not find %#v in %#v", params, arg)
	}

	err := m.TraversalsByNameFunc(v.Type(), params.toNames(), func(i int, fi *FieldInfo) error {
		if fi == nil {
			return fmt.Errorf("could not find argument '%s' in %#v", params[i].Name, arg)
		}

		sqlValue, err := fi.LValue(dialect, &params[i], v)
		if err != nil {
			return err
		}
		arglist = append(arglist, sqlValue)
		return nil
	})

	return arglist, err
}

// like bindParams, but for maps.
func bindMapArgs(dialect Dialect, params Params, arg map[string]interface{}) ([]interface{}, error) {
	arglist := make([]interface{}, 0, len(params))

	for idx := range params {
		val, ok := arg[params[idx].Name]
		if !ok {
			return arglist, fmt.Errorf("could not find argument '%s' in %#v", params[idx].Name, arg)
		}
		sqlValue, err := toSQLType(dialect, &params[idx], val)
		if err != nil {
			return nil, err
		}
		arglist = append(arglist, sqlValue)
	}
	return arglist, nil
}

func MarkSQLError(sql string, index int) string {
	result := fmt.Sprintf("%s[****ERROR****]->%s", sql[0:index], sql[index:])
	return result
}
