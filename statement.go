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

type MappedStatement struct {
	id          string
	sqlTemplate *template.Template
	sql         string
	sqlCompiled *struct {
		dollarSQL string
		questSQL  string
		bindNames []string
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
		fragments, bindArgs, err := compileNamedQuery(sqlTemp)
		if err != nil {
			return nil, errors.New("sql is invalid named sql of '" + id + "', " + err.Error())
		}
		if len(bindArgs) != 0 {
			stmt.sqlCompiled = &struct {
				dollarSQL string
				questSQL  string
				bindNames []string
			}{dollarSQL: concatFragments(DOLLAR, fragments, bindArgs),
				questSQL:  concatFragments(QUESTION, fragments, bindArgs),
				bindNames: bindArgs}
		}
	}
	return stmt, nil
}

func compileNamedQuery(s string) ([]string, []string, error) {
	idx := strings.Index(s, "#{")
	if idx < 0 {
		return []string{s}, nil, nil
	}

	var fragments []string
	var argments []string
	for {
		fragments = append(fragments, s[:idx])
		s = s[idx+len("#{"):]
		end := strings.Index(s, "}")
		if end < 0 {
			return nil, nil, errors.New(MarkSQLError(s, idx))
		}
		argments = append(argments, s[:end])

		s = s[end+len("}"):]
		idx = strings.Index(s, "#{")
		if idx < 0 {
			fragments = append(fragments, s)
			return fragments, argments, nil
		}
	}
}

func concatFragments(bindType int, fragments, names []string) string {
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

func bindNamedQuery(bindArgs, paramNames []string, paramValues []interface{}, mapper *reflectx.Mapper) ([]interface{}, error) {
	if len(bindArgs) == 0 {
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
			return bindMapArgs(bindArgs, mapArgs)
		}
		return bindStruct(bindArgs, paramValues[0], mapper)
	}

	if len(bindArgs) != len(paramValues) && len(paramValues) == 1 {
		if mapArgs, ok := paramValues[0].(map[string]interface{}); ok {
			return bindMapArgs(bindArgs, mapArgs)
		}
		return bindStruct(bindArgs, paramValues[0], mapper)
	}

	bindValues := make([]interface{}, len(bindArgs))
	for idx := range bindArgs {
		bindName := bindArgs[idx]
		foundIndex := -1
		for nidx := range paramNames {
			if paramNames[nidx] == bindName {
				bindValues[idx] = paramValues[nidx]
				foundIndex = nidx
				break
			}
		}
		if foundIndex >= 0 {
			continue
		}
		dotIndex := strings.IndexByte(bindArgs[idx], '.')
		if dotIndex < 0 {
			continue
		}
		variableName := bindName[:dotIndex]
		for nidx := range paramNames {
			if paramNames[nidx] == variableName {
				v := paramValues[nidx]
				if v != nil {
					rv := mapper.FieldByName(reflect.ValueOf(v), bindName[dotIndex+1:])
					if rv.IsValid() {
						bindValues[idx] = rv.Interface()
					}
				}
				break
			}
		}
	}
	return bindValues, nil
}

// private interface to generate a list of interfaces from a given struct
// type, given a list of names to pull out of the struct.  Used by public
// BindStruct interface.
func bindStruct(names []string, arg interface{}, m *reflectx.Mapper) ([]interface{}, error) {
	arglist := make([]interface{}, 0, len(names))

	// grab the indirected value of arg
	v := reflect.ValueOf(arg)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		if len(names) <= 1 {
			return []interface{}{arg}, nil
		}
		return arglist, fmt.Errorf("could not find %v in %#v", names, arg)
	}

	err := m.TraversalsByNameFunc(v.Type(), names, func(i int, t []int) error {
		if len(t) == 0 {
			return fmt.Errorf("could not find name %s in %#v", names[i], arg)
		}

		val := reflectx.FieldByIndexesReadOnly(v, t)
		arglist = append(arglist, val.Interface())

		return nil
	})

	return arglist, err
}

// like bindArgs, but for maps.
func bindMapArgs(names []string, arg map[string]interface{}) ([]interface{}, error) {
	arglist := make([]interface{}, 0, len(names))

	for _, name := range names {
		val, ok := arg[name]
		if !ok {
			return arglist, fmt.Errorf("could not find name %s in %#v", name, arg)
		}
		arglist = append(arglist, val)
	}
	return arglist, nil
}

func MarkSQLError(sql string, index int) string {
	result := fmt.Sprintf("%s[****ERROR****]->%s", sql[0:index], sql[index:])
	return result
}
