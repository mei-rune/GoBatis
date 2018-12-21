package generator

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

func doTablenameToken(defaultRecordType string, attrs [][2]string, sb *strings.Builder) {
	var recordType, alias string
	for _, kv := range attrs {
		switch kv[0] {
		case "type":
			recordType = kv[1]
		case "alias", "as":
			alias = kv[1]
		}
	}

	sb.WriteString(`      if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&`)
	if recordType != "" {
		sb.WriteString(recordType)
	} else if defaultRecordType != "" {
		sb.WriteString(defaultRecordType)
	} else {
		sb.WriteString("XXX")
	}
	sb.WriteString(`{})); err != nil {
        return err
      } else {
        sb.WriteString(tablename)
      }` + "\r\n")
	if alias != "" {
		sb.WriteString("      sb.WriteString(\" AS \")\r\n")
		sb.WriteString("      sb.WriteString(")
		sb.WriteString(fmt.Sprintf("%q", alias))
		sb.WriteString(")\r\n")
	}
}

func doConstantToken(attrs [][2]string, sb *strings.Builder) {
	var id string
	for _, kv := range attrs {
		switch kv[0] {
		case "id", "name", "value":
			id = kv[1]
		}
	}

	if id == "" {
		panic(errors.New("name attribute is missing in <constant />"))
	}

	sb.WriteString(`      if cValue, ok := ctx.Config.Constants[`)
	sb.WriteString(fmt.Sprintf("%q", id))
	sb.WriteString(`]); !ok {
        return errors.New("constant '` + id + `' is missing!")
      }
      sb.WriteString(fmt.Sprint(cValue))
`)
}

func preprocessingSQL(name string, isNew bool, sqlStr, defaultRecordType string) string {
	tokenIndex, startIdx, endIdx, attrs := readXMLToken(sqlStr, []string{"<tablename", "<constant"})
	if tokenIndex < 0 {
		if isNew {
			return name + " := " + fmt.Sprintf("%q", sqlStr)
		}
		return name + " = " + fmt.Sprintf("%q", sqlStr)
	}

	var sb strings.Builder
	sb.WriteString("      var sb strings.Builder\r\n")

	for tokenIndex >= 0 {
		if startIdx > 0 {
			sb.WriteString("      sb.WriteString(")
			sb.WriteString(fmt.Sprintf("%q", sqlStr[:startIdx]))
			sb.WriteString(")\r\n")
		}

		switch tokenIndex {
		case 0:
			doTablenameToken(defaultRecordType, attrs, &sb)
		case 1:
			doConstantToken(attrs, &sb)
		default:
			panic(errors.New("tokenIndex is unknown"))
		}

		sqlStr = sqlStr[endIdx:]
		tokenIndex, startIdx, endIdx, attrs = readXMLToken(sqlStr, []string{"<tablename", "<constant"})
	}

	if sqlStr != "" {
		sb.WriteString("      sb.WriteString(")
		sb.WriteString(fmt.Sprintf("%q", sqlStr))
		sb.WriteString(")\r\n")
	}
	sb.WriteString("      ")
	sb.WriteString(name)
	if isNew {
		sb.WriteString(" := sb.String()\r\n")
	} else {
		sb.WriteString(" = sb.String()\r\n")
	}
	return sb.String()
}

func indexTokens(s string, tokens []string) (int, int) {
	var minIndex = -1
	var tokenIndex = -1
	for pos, token := range tokens {
		idx := strings.Index(s, token)
		if idx < 0 {
			continue
		}

		if minIndex < 0 || minIndex > idx {
			minIndex = idx
			tokenIndex = pos
		}
	}
	return tokenIndex, minIndex
}

func readXMLToken(sqlStr string, tokenTypes []string) (tokenIndex, startIdx, endIdx int, attributes [][2]string) {
	lowerSqlStr := strings.ToLower(sqlStr)

	tokenIndex, startIdx = indexTokens(lowerSqlStr, tokenTypes)
	if tokenIndex < 0 {
		startIdx = -1
		return
	}

	exceptIndex := startIdx + len(tokenTypes[tokenIndex])
	if len(sqlStr) <= exceptIndex {
		startIdx = -1
		return
	}

	endIdx = strings.Index(sqlStr[exceptIndex:], "/>")
	if endIdx < 0 {
		startIdx = -1
		return
	}

	if endIdx == 0 {
		endIdx = exceptIndex + len("/>")
		return
	}

	if !unicode.IsSpace(rune(sqlStr[exceptIndex])) {
		startIdx = -1
		return
	}

	endIdx += exceptIndex
	endIdx += len("/>")
	attrText := sqlStr[exceptIndex : endIdx-len("/>")]

	if strings.Contains(attrText, "<") || strings.Contains(attrText, ">") {
		startIdx = -1
		return
	}
	fields := strings.Fields(attrText)
	attributes = make([][2]string, 0, len(fields))
	for _, field := range fields {
		kv := strings.Split(field, "=")
		if len(kv) != 2 {
			startIdx = -1
			return
		}
		attributes = append(attributes, [2]string{kv[0], strings.Trim(kv[1], "\"")})
	}

	return
}
