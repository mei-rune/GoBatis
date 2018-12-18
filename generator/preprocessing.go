package generator

import (
	"fmt"
	"strings"
	"unicode"
)

func preprocessingSQL(name string, isNew bool, sqlStr, defaultRecordType string) string {
	startIdx, endIdx, recordType, alias := readTablenameToken(sqlStr)
	if startIdx < 0 {
		if isNew {
			return name + " := " + fmt.Sprintf("%q", sqlStr)
		}
		return name + " = " + fmt.Sprintf("%q", sqlStr)
	}

	var sb strings.Builder
	sb.WriteString("      var sb strings.Builder\r\n")

	for startIdx >= 0 {
		sb.WriteString("      sb.WriteString(")
		sb.WriteString(fmt.Sprintf("%q", sqlStr[:startIdx]))
		sb.WriteString(")\r\n")

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

		sqlStr = sqlStr[endIdx:]
		startIdx, endIdx, recordType, alias = readTablenameToken(sqlStr)
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

func readTablenameToken(sqlStr string) (int, int, string, string) {
	return readXMLToken(sqlStr, "<tablename")
}

func readXMLToken(sqlStr string, tokenType string) (startIdx, endIdx int, recordType, alias string) {
	lowerSqlStr := strings.ToLower(sqlStr)

	startIdx = strings.Index(lowerSqlStr, tokenType+"/>")
	if startIdx >= 0 {
		endIdx = startIdx + len(tokenType) + len("/>")
		return
	}

	startIdx = strings.Index(lowerSqlStr, tokenType)
	exceptIndex := startIdx + len(tokenType)
	if startIdx < 0 || len(sqlStr) <= exceptIndex || !unicode.IsSpace(rune(sqlStr[exceptIndex])) {
		startIdx = -1
		return
	}

	endIdx = strings.Index(sqlStr[exceptIndex:], "/>")
	if endIdx < 0 {
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
	for _, field := range fields {
		kv := strings.Split(field, "=")
		if len(kv) != 2 {
			startIdx = -1
			return
		}
		switch kv[0] {
		case "type":
			recordType = strings.Trim(kv[1], "\"")
		case "alias", "as":
			alias = strings.Trim(kv[1], "\"")
		}
	}

	return
}
