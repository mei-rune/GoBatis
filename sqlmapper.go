package osm

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"
)

type statementType int
type resultType int

const (
	typeSelect statementType = 0
	typeUpdate statementType = 1
	typeInsert statementType = 2
	typeDelete statementType = 3

	resultUnknown resultType = 0
	resultMap     resultType = 1
	resultStruct  resultType = 2
)

type sqlMapper struct {
	id          string
	sqlTemplate *template.Template
	sql         string
	sqlType     statementType
	result      resultType
}

type stmtXML struct {
	ID     string `xml:"id,attr"`
	Result string `xml:"result,attr"`
	SQL    string `xml:",chardata"`
}

type osmXML struct {
	Selects []stmtXML `xml:"select"`
	Deletes []stmtXML `xml:"delete"`
	Updates []stmtXML `xml:"update"`
	Inserts []stmtXML `xml:"insert"`
}

func readMappers(path string) ([]*sqlMapper, error) {
	sqlMappers := make([]*sqlMapper, 0)

	xmlFile, err := os.Open(path)
	if err != nil {
		return nil, errors.New("Error opening file: " + err.Error())
	}
	defer xmlFile.Close()

	osmXMLObj := osmXML{}

	decoder := xml.NewDecoder(xmlFile)
	if err = decoder.Decode(&osmXMLObj); err != nil {
		return nil, errors.New("Error decode file '" + path + "': " + err.Error())
	}

	for _, deleteStmt := range osmXMLObj.Deletes {
		mapper, err := newMapper(deleteStmt, typeDelete)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + deleteStmt.ID + "': " + err.Error())
		}
		sqlMappers = append(sqlMappers, mapper)
	}
	for _, insertStmt := range osmXMLObj.Inserts {
		mapper, err := newMapper(insertStmt, typeInsert)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + insertStmt.ID + "': " + err.Error())
		}
		sqlMappers = append(sqlMappers, mapper)
	}
	for _, selectStmt := range osmXMLObj.Selects {
		mapper, err := newMapper(selectStmt, typeSelect)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + selectStmt.ID + "': " + err.Error())
		}
		sqlMappers = append(sqlMappers, mapper)
	}
	for _, updateStmt := range osmXMLObj.Updates {
		mapper, err := newMapper(updateStmt, typeUpdate)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + updateStmt.ID + "': " + err.Error())
		}
		sqlMappers = append(sqlMappers, mapper)
	}
	return sqlMappers, nil
}

func newMapper(stmt stmtXML, sqlType statementType) (*sqlMapper, error) {
	sqlMapperObj := new(sqlMapper)
	sqlMapperObj.id = stmt.ID
	sqlMapperObj.sqlType = sqlType

	switch strings.ToLower(stmt.Result) {
	case "":
		sqlMapperObj.result = resultUnknown
	case "struct", "type", "resultstruct", "resultType":
		sqlMapperObj.result = resultStruct
	case "map", "resultmap":
		// sqlMapperObj.result = resultMap
		return nil, errors.New("result '" + stmt.Result + "' of '" + stmt.ID + "' is unsupported")
	default:
		return nil, errors.New("result '" + stmt.Result + "' of '" + stmt.ID + "' is unsupported")
	}

	sqlTemp := strings.Replace(strings.TrimSpace(stmt.SQL), "\r\n", " ", -1)
	sqlTemp = strings.Replace(sqlTemp, "\n", " ", -1)
	sqlTemp = strings.Replace(sqlTemp, "\t", " ", -1)
	for strings.Contains(sqlTemp, "  ") {
		sqlTemp = strings.Replace(sqlTemp, "  ", " ", -1)
	}
	sqlTemp = strings.Trim(sqlTemp, "\t\n ")

	sqlMapperObj.sql = sqlTemp

	var err error
	sqlMapperObj.sqlTemplate, err = template.New(stmt.ID).Parse(sqlTemp)
	if err != nil {
		return nil, errors.New("sql is invalid go template of '" + stmt.ID + "', " + err.Error() + "\r\n\t" + sqlTemp)
	}
	return sqlMapperObj, nil
}

func markSQLError(sql string, index int) string {
	result := fmt.Sprintf("%s[****ERROR****]->%s", sql[0:index], sql[index:])
	return result
}
