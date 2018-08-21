package gobatis

import (
	"encoding/xml"
	"errors"
	"os"
	"strings"
)

type stmtXML struct {
	ID     string `xml:"id,attr"`
	Result string `xml:"result,attr"`
	SQL    string `xml:",innerxml"`
}

type xmlConfig struct {
	Selects []stmtXML `xml:"select"`
	Deletes []stmtXML `xml:"delete"`
	Updates []stmtXML `xml:"update"`
	Inserts []stmtXML `xml:"insert"`
}

func readMappedStatements(ctx *InitContext, path string) ([]*MappedStatement, error) {
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
		mapper, err := newMapppedStatement(ctx, deleteStmt, StatementTypeDelete)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + deleteStmt.ID + "': " + err.Error())
		}
		statements = append(statements, mapper)
	}
	for _, insertStmt := range xmlObj.Inserts {
		mapper, err := newMapppedStatement(ctx, insertStmt, StatementTypeInsert)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + insertStmt.ID + "': " + err.Error())
		}
		statements = append(statements, mapper)
	}
	for _, selectStmt := range xmlObj.Selects {
		mapper, err := newMapppedStatement(ctx, selectStmt, StatementTypeSelect)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + selectStmt.ID + "': " + err.Error())
		}
		statements = append(statements, mapper)
	}
	for _, updateStmt := range xmlObj.Updates {
		mapper, err := newMapppedStatement(ctx, updateStmt, StatementTypeUpdate)
		if err != nil {
			return nil, errors.New("Error parse file '" + path + "' on '" + updateStmt.ID + "': " + err.Error())
		}
		statements = append(statements, mapper)
	}
	return statements, nil
}

func newMapppedStatement(ctx *InitContext, stmt stmtXML, sqlType StatementType) (*MappedStatement, error) {
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

	// http://www.mybatis.org/mybatis-3/dynamic-sql.html
	// for _, tag := range []string{"<where>", "<set>", "<if", "<foreach", "<case"} {
	// 	if strings.Contains(stmt.SQL, tag) { }
	// }

	return NewMapppedStatement(ctx, stmt.ID, sqlType, resultType, stmt.SQL)
}
