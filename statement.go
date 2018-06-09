package gobatis

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/runner-mei/GoBatis/reflectx"
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
	id          string
	sqlTemplate *template.Template
	sql         string
	sqlCompiled *struct {
		dollarSQL  string
		questSQL   string
		bindParams Params
	}
	sqlType StatementType
	result  ResultType
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

	if strings.Contains(sqlTemp, "{{") {
		tpl, err := template.New(id).Parse(sqlTemp)
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
			stmt.sqlCompiled = &struct {
				dollarSQL  string
				questSQL   string
				bindParams Params
			}{dollarSQL: concatFragments(DOLLAR, fragments, bindParams),
				questSQL:   concatFragments(QUESTION, fragments, bindParams),
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
		argments = append(argments, Param{Name: s[:end]})

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

func concatFragments(bindType int, fragments []string, names Params) string {
	if bindType == QUESTION {
		return strings.Join(fragments, "?")
	}
	var sb strings.Builder
	sb.WriteString(fragments[0])
	for i := 1; i < len(fragments); i++ {
		sb.WriteString("$")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(fragments[i])
	}
	return sb.String()
}

func bindNamedQuery(bindParams Params, paramNames []string, paramValues []interface{}, mapper *reflectx.Mapper) ([]interface{}, error) {
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
			return bindMapArgs(bindParams, mapArgs)
		}
		return bindStruct(bindParams, paramValues[0], mapper)
	}

	if len(bindParams) != len(paramValues) && len(paramValues) == 1 {
		if mapArgs, ok := paramValues[0].(map[string]interface{}); ok {
			return bindMapArgs(bindParams, mapArgs)
		}
		return bindStruct(bindParams, paramValues[0], mapper)
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
			sqlValue, err := toSQLType(&bindParams[idx], paramValues[foundIndex])
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
					rv := mapper.FieldByName(reflect.ValueOf(v), bindParams[idx].Name[dotIndex+1:])
					if rv.IsValid() {
						sqlValue, err := toSQLTypeWith(&bindParams[idx], rv)
						if err != nil {
							return nil, err
						}
						bindValues[idx] = sqlValue
						continue
					}
				}
			}
		}

		sqlValue, err := toSQLType(&bindParams[idx], nil)
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
func bindStruct(params Params, arg interface{}, m *reflectx.Mapper) ([]interface{}, error) {
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

	err := m.TraversalsByNameFunc(v.Type(), params.toNames(), func(i int, t []int) error {
		if len(t) == 0 {
			return fmt.Errorf("could not find argument '%s' in %#v", params[i].Name, arg)
		}

		val := reflectx.FieldByIndexesReadOnly(v, t)
		sqlValue, err := toSQLTypeWith(&params[i], val)
		if err != nil {
			return err
		}
		arglist = append(arglist, sqlValue)
		return nil
	})

	return arglist, err
}

// like bindParams, but for maps.
func bindMapArgs(params Params, arg map[string]interface{}) ([]interface{}, error) {
	arglist := make([]interface{}, 0, len(params))

	for idx := range params {
		val, ok := arg[params[idx].Name]
		if !ok {
			return arglist, fmt.Errorf("could not find argument '%s' in %#v", params[idx].Name, arg)
		}
		sqlValue, err := toSQLType(&params[idx], val)
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
