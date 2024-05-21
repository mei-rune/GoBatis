package core

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

type stmtXML struct {
	ID     string `xml:"id,attr"`
	Result string `xml:"result,attr,omitempty"`
	SQL    string `xml:",innerxml"` // nolint
}

type sqlFragmentXML struct {
	ID     string `xml:"id,attr"`
	SQL    string `xml:",innerxml"` // nolint
}

type xmlConfig struct {
	XMLName xml.Name  `xml:"gobatis"` // nolint
	Selects []stmtXML `xml:"select"`  // nolint
	Deletes []stmtXML `xml:"delete"`  // nolint
	Updates []stmtXML `xml:"update"`  // nolint
	Inserts []stmtXML `xml:"insert"`  // nolint
	SqlFragments []sqlFragmentXML  `xml:"sql"`  // nolint
}

func readMappedStatementsFromXMLFile(ctx *InitContext, filename string) ([]*MappedStatement, error) {
	statements := make([]*MappedStatement, 0)

	xmlFile, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("Error opening file: " + err.Error())
	}
	defer xmlFile.Close()

	xmlObj := xmlConfig{}
	decoder := xml.NewDecoder(xmlFile)
	if err = decoder.Decode(&xmlObj); err != nil {
		return nil, errors.New("Error decode file '" + filename + "': " + err.Error())
	}

	var sqlFragments = map[string]SqlExpression{}

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

			sf := sqlFragments[id]
			if sf != nil {
				return sf, nil
			}

			var sqlStr string
			for _, stmt := range xmlObj.SqlFragments {
				if stmt.ID == id {
					sqlStr = stmt.SQL
					if sqlStr == "" {
						return nil, errors.New("sql '"+ id +"' is empty in file '"+filename+"'")
					}
					break
				}
			}

			if sqlStr == "" {				
				return nil, errors.New("sql '"+id+"' missing")
			}
			segements, err := readSQLStatementForXML(stmtctx, sqlStr)
			if err != nil {
				return nil, err
			}
			sf = expressionArray(segements)
			sqlFragments[id] = sf
			return sf, nil
		}

	for _, deleteStmt := range xmlObj.Deletes {
		stmt, err := newMapppedStatementFromXML(stmtctx, deleteStmt, StatementTypeDelete)
		if err != nil {
			return nil, errors.New("Error parse file '" + filename + "' on '" + deleteStmt.ID + "': " + err.Error())
		}
		statements = append(statements, stmt)
	}

	for _, insertStmt := range xmlObj.Inserts {
		stmt, err := newMapppedStatementFromXML(stmtctx, insertStmt, StatementTypeInsert)
		if err != nil {
			return nil, errors.New("Error parse file '" + filename + "' on '" + insertStmt.ID + "': " + err.Error())
		}
		statements = append(statements, stmt)
	}
	for _, selectStmt := range xmlObj.Selects {
		stmt, err := newMapppedStatementFromXML(stmtctx, selectStmt, StatementTypeSelect)
		if err != nil {
			return nil, errors.New("Error parse file '" + filename + "' on '" + selectStmt.ID + "': " + err.Error())
		}
		statements = append(statements, stmt)
	}
	for _, updateStmt := range xmlObj.Updates {
		stmt, err := newMapppedStatementFromXML(stmtctx, updateStmt, StatementTypeUpdate)
		if err != nil {
			return nil, errors.New("Error parse file '" + filename + "' on '" + updateStmt.ID + "': " + err.Error())
		}
		statements = append(statements, stmt)
	}
	return statements, nil
}

type StmtContext struct {
	*InitContext

	FindSqlFragment func(string) (SqlExpression, error)
}

func newMapppedStatementFromXML(ctx *StmtContext, stmt stmtXML, sqlType StatementType) (*MappedStatement, error) {
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

	s, err := NewMapppedStatement(ctx, stmt.ID, sqlType, resultType, stmt.SQL)
	if err != nil {
		return nil, err
	}
	s.source = "xml"
	return s, nil
}

func loadDynamicSQLFromXML(ctx *StmtContext, sqlStr string) (DynamicSQL, error) {
	segements, err := readSQLStatementForXML(ctx, sqlStr)
	if err != nil {
		return nil, err
	}
	return expressionArray(segements), nil
}

func readSQLStatementForXML(ctx *StmtContext, sqlStr string) ([]SqlExpression, error) {
	txtBegin := `<?xml version="1.0" encoding="utf-8"?>
<statement>`
	txtEnd := `</statement>`

	decoder := xml.NewDecoder(strings.NewReader(txtBegin + sqlStr + txtEnd))
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("EOF isnot except in the root element")
			}
			return nil, err
		}

		switch el := token.(type) {
		case xml.StartElement:
			if el.Name.Local == "statement" {
				return readElementForXML(ctx, decoder, "")
			}
		case xml.Directive, xml.ProcInst, xml.Comment:
		case xml.CharData:
			if len(bytes.TrimSpace([]byte(el))) != 0 {
				return nil, fmt.Errorf("CharData isnot except element - %s", el)
			}
		default:
			return nil, fmt.Errorf("%T isnot except element in the element", token)
		}
	}
}

func readElementForXML(ctx *StmtContext, decoder *xml.Decoder, tag string) ([]SqlExpression, error) {
	var sb strings.Builder
	var expressions []SqlExpression
	var lastPrint *string

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("EOF isnot except in the '" + tag + "' element")
			}
			return nil, err
		}

		switch el := token.(type) {
		case xml.StartElement:
			var prefix string
			if s := sb.String(); strings.TrimSpace(s) != "" {
				segement, err := newRawExpression(s)
				if err != nil {
					return nil, err
				}

				expressions = append(expressions, segement)
			} else if lastPrint != nil {
				*lastPrint = sb.String()
			} else {
				prefix = sb.String()
			}
			lastPrint = nil
			sb.Reset()

			switch el.Name.Local {
			case "if":
				contents, err := readElementForXML(ctx, decoder, tag+"/if")
				if err != nil {
					return nil, err
				}
				if len(contents) == 0 {
					break
				}

				segement, err := newIFExpression(readElementAttrForXML(el.Attr, "test"), contents)
				if err != nil {
					return nil, err
				}
				expressions = append(expressions, segement)
			case "else":
				if !strings.HasSuffix(tag, "/if") {
					return nil, errors.New("element else only exist in the <if>")
				}

				txt, err := readElementTextForXML(decoder, tag+"/else")
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(txt) != "" {
					return nil, errors.New("element else must has empty")
				}

				test := readElementAttrForXML(el.Attr, "if-test")
				if strings.TrimSpace(test) == "" {
					expressions = append(expressions, elseExpr)
					continue
				}

				expressions = append(expressions, elseExpression{test: test})
			case "foreach":
				contents, err := readElementForXML(ctx, decoder, tag+"/foreach")
				if err != nil {
					return nil, err
				}

				foreach, err := newForEachExpression(xmlForEachElement{
					item:         readElementAttrForXML(el.Attr, "item"),
					index:        readElementAttrForXML(el.Attr, "index"),
					collection:   readElementAttrForXML(el.Attr, "collection"),
					openTag:      readElementAttrForXML(el.Attr, "open"),
					separatorTag: readElementAttrForXML(el.Attr, "separator"),
					closeTag:     readElementAttrForXML(el.Attr, "close"),
					contents:     contents,
				})
				if err != nil {
					return nil, err
				}

				expressions = append(expressions, foreach)
			case "chose", "choose":
				choseEl, err := loadChoseElementForXML(ctx, decoder, tag+"/"+el.Name.Local)
				if err != nil {
					return nil, err
				}
				chose, err := newChoseExpression(*choseEl)
				if err != nil {
					return nil, err
				}
				expressions = append(expressions, chose)
			case "where":
				array, err := readElementForXML(ctx, decoder, tag+"/where")
				if err != nil {
					return nil, err
				}

				expressions = append(expressions, &whereExpression{expressions: array})
			case "set":
				array, err := readElementForXML(ctx, decoder, tag+"/set")
				if err != nil {
					return nil, err
				}

				expressions = append(expressions, &setExpression{expressions: array})
			case "print":
				content, err := readElementTextForXML(decoder, tag+"/print")
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(content) != "" {
					return nil, errors.New("element print must is empty element")
				}
				value := readElementAttrForXML(el.Attr, "value")
				if value == "" {
					return nil, errors.New("element print must has a 'value' notempty attribute")
				}
				inStr := readElementAttrForXML(el.Attr, "inStr")
				if inStr == "" {
					inStr = readElementAttrForXML(el.Attr, "in_str")
				}
				printExpr := &printExpression{
					prefix: prefix,
					value:  value,
					fmt:    readElementAttrForXML(el.Attr, "fmt"),
					inStr:  inStr,
				}
				lastPrint = &printExpr.suffix
				expressions = append(expressions, printExpr)
			case "like":
				content, err := readElementTextForXML(decoder, tag+"/like")
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(content) != "" {
					return nil, errors.New("element like must is empty element")
				}
				value := readElementAttrForXML(el.Attr, "value")
				if value == "" {
					return nil, errors.New("element like must has a 'value' notempty attribute")
				}

				isPrefix := readElementAttrForXML(el.Attr, "isPrefix")
				if isPrefix == "" {
					isPrefix = readElementAttrForXML(el.Attr, "is_prefix")
				}
				isSuffix := readElementAttrForXML(el.Attr, "isSuffix")
				if isSuffix == "" {
					isSuffix = readElementAttrForXML(el.Attr, "is_suffix")
				}

				likeExpr := &likeExpression{
					prefix:   prefix,
					value:    value,
					isPrefix: isPrefix,
					isSuffix: isSuffix,
				}
				lastPrint = &likeExpr.suffix
				expressions = append(expressions, likeExpr)
			case "pagination":
				content, err := readElementTextForXML(decoder, tag+"/pagination")
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(content) != "" {
					return nil, errors.New("element pagination must is empty element")
				}

				page := readElementAttrForXML(el.Attr, "page")
				size := readElementAttrForXML(el.Attr, "size")
				if page != "" || size != "" {
					pagination := &pageExpression{
						page: page,
						size: size,
					}
					if pagination.page == "" {
						pagination.page = "page"
					}
					if pagination.size == "" {
						pagination.size = "size"
					}
					expressions = append(expressions, pagination)
					break
				}

				pagination := &limitExpression{
					offset: readElementAttrForXML(el.Attr, "offset"),
					limit:  readElementAttrForXML(el.Attr, "limit"),
				}
				if pagination.offset == "" {
					pagination.offset = "offset"
				}
				if pagination.limit == "" {
					pagination.limit = "limit"
				}
				expressions = append(expressions, pagination)
			case "order_by", "orderBy", "sort_by", "sortBy":
				content, err := readElementTextForXML(decoder, tag+"/"+el.Name.Local)
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(content) != "" {
					return nil, errors.New("element order_by must is empty element")
				}
				orderBy := &orderByExpression{
					sort: readElementAttrForXML(el.Attr, "sort"),
				}
				if orderBy.sort == "" {
					orderBy.sort = readElementAttrForXML(el.Attr, "by")
				}
				orderBy.prefix = readElementAttrForXML(el.Attr, "prefix")
				orderBy.direction = readElementAttrForXML(el.Attr, "direction")

				expressions = append(expressions, orderBy)
			case "trim":
				array, err := readElementForXML(ctx, decoder, tag+"/trim")
				if err != nil {
					return nil, err
				}
				if len(array) == 0 {
					return nil, errors.New("element trim must isnot empty element")
				}

				trimExpr := &trimExpression{
					prefixoverride: splitTrimStrings(readElementAttrForXML(el.Attr, "prefixOverrides")),
					suffixoverride: splitTrimStrings(readElementAttrForXML(el.Attr, "suffixOverrides")),
					expressions:    array,
				}
				if prefix := readElementAttrForXML(el.Attr, "prefix"); prefix != "" {
					expr, err := newRawExpression(prefix)
					if err != nil {
						return nil, errors.New("element trim.prefix is invalid - '" + prefix + "', " + err.Error())
					}
					trimExpr.prefix = expr
				}

				if suffix := readElementAttrForXML(el.Attr, "suffix"); suffix != "" {
					expr, err := newRawExpression(suffix)
					if err != nil {
						return nil, errors.New("element trim.suffix is invalid - '" + suffix + "'")
					}
					trimExpr.suffix = expr
				}
				expressions = append(expressions, trimExpr)
			case "include":
				array, err := readElementForXML(ctx, decoder, tag+"/include")
				if err != nil {
					return nil, err
				}
				if len(array) != 0 {
					return nil, errors.New("element include must is empty element")
				}

				refid := readElementAttrForXML(el.Attr, "refid")
				if refid == "" {
						return nil, errors.New("element include.refid is missing")
				}

				expr, err := ctx.FindSqlFragment(refid)
				if err != nil {
					return nil, errors.New("element include.refid '"+refid+"' invalid, " + err.Error())
				}

				expressions = append(expressions, expr)
			case "value-range", "value_range", "valuerange":
				content, err := readElementTextForXML(decoder, tag+"/"+el.Name.Local)
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(content) != "" {
					return nil, errors.New("element '" + tag + "' must is empty element")
				}
				valueRange := &valueRangeExpression{
					field: readElementAttrForXML(el.Attr, "field"),
					value: readElementAttrForXML(el.Attr, "value"),
				}

				if prefix := readElementAttrForXML(el.Attr, "prefix"); prefix != "" {
					expr, err := newRawExpression(prefix)
					if err != nil {
						return nil, errors.New("element " + tag + ".prefix is invalid - '" + prefix + "'")
					}
					valueRange.prefix = expr
				}
				if suffix := readElementAttrForXML(el.Attr, "suffix"); suffix != "" {
					expr, err := newRawExpression(suffix)
					if err != nil {
						return nil, errors.New("element " + tag + ".suffix is invalid - '" + suffix + "'")
					}
					valueRange.suffix = expr
				}

				if valueRange.field == "" {
					return nil, errors.New("element '" + tag + ".field' is missing")
				}
				if valueRange.value == "" {
					return nil, errors.New("element '" + tag + ".value' is missing")
				}
				expressions = append(expressions, valueRange)
			default:
				if tag == "" {
					return nil, fmt.Errorf("StartElement(" + el.Name.Local + ") isnot except element in the root element")
				}
				return nil, fmt.Errorf("StartElement(" + el.Name.Local + ") isnot except element in the '" + tag + "' element")
			}

		case xml.EndElement:
			if s := sb.String(); strings.TrimSpace(s) != "" {
				segement, err := newRawExpression(s)
				if err != nil {
					return nil, err
				}

				expressions = append(expressions, segement)
			} else if lastPrint != nil {
				*lastPrint = sb.String()
			}
			sb.Reset()
			lastPrint = nil // nolint: ineffassign

			return expressions, nil
		case xml.CharData:
			sb.Write(el)
		case xml.Directive, xml.ProcInst, xml.Comment:
			sb.WriteString(" ")
		default:
			return nil, fmt.Errorf("%T isnot except element in the '"+tag+"'", token)
		}
	}
}

func readElementTextForXML(decoder *xml.Decoder, tag string) (string, error) {
	var sb strings.Builder
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return "", fmt.Errorf("EOF isnot except in the " + tag + " element")
			}
			return "", err
		}

		switch el := token.(type) {
		case xml.StartElement:
			return "", fmt.Errorf("StartElement(" + el.Name.Local + ") isnot except in the " + tag + " element")
		case xml.EndElement:
			return sb.String(), nil
		case xml.CharData:
			sb.Write(el)
		case xml.Directive, xml.ProcInst, xml.Comment:
			sb.WriteString(" ")
		default:
			return "", fmt.Errorf("%T isnot except element in the '"+tag+"'", token)
		}
	}
}

func readElementAttrForXML(attrs []xml.Attr, name string) string {
	for idx := range attrs {
		if attrs[idx].Name.Local == name {
			return attrs[idx].Value
		}
	}
	return ""
}

func loadChoseElementForXML(ctx *StmtContext, decoder *xml.Decoder, tag string) (*xmlChoseElement, error) {
	var segement xmlChoseElement
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("EOF isnot except in the '" + tag + "' element")
			}
			return nil, err
		}

		switch el := token.(type) {
		case xml.StartElement:
			if el.Name.Local == "when" {
				contents, err := readElementForXML(ctx, decoder, "when")
				if err != nil {
					return nil, err
				}

				if len(contents) == 0 {
					segement.when = append(segement.when, xmlWhenElement{
						content: rawString(""),
						test:    readElementAttrForXML(el.Attr, "test"),
					})
					break
				}

				var content SqlExpression
				if len(contents) == 1 {
					content = contents[0]
				} else if len(contents) > 1 {
					content = expressionArray(contents)
				}

				segement.when = append(segement.when, xmlWhenElement{
					content: content,
					test:    readElementAttrForXML(el.Attr, "test"),
				})
				break
			}

			if el.Name.Local == "otherwise" {
				contents, err := readElementForXML(ctx, decoder, "otherwise")
				if err != nil {
					return nil, err
				}
				if len(contents) == 0 {
					break
				}

				if len(contents) == 1 {
					segement.otherwise = contents[0]
				} else if len(contents) > 1 {
					segement.otherwise = expressionArray(contents)
				}
				break
			}

			return nil, fmt.Errorf("StartElement(" + el.Name.Local + ") isnot except in the '" + tag + "' element")
		case xml.EndElement:
			return &segement, nil
		case xml.CharData:
			if len(bytes.TrimSpace(el)) != 0 {
				return nil, fmt.Errorf("CharData(" + string(el) + ") isnot except in the '" + tag + "' element")
			}
		case xml.Directive, xml.ProcInst, xml.Comment:
		default:
			return nil, fmt.Errorf("%T isnot except element in the '"+tag+"'", token)
		}
	}
}

type xmlWhenElement struct {
	test    string
	content SqlExpression
}

func (when *xmlWhenElement) String() string {
	var sb strings.Builder
	sb.WriteString("<when test=\"")
	sb.WriteString(when.test)
	sb.WriteString("\">")
	sb.WriteString(when.content.String())
	sb.WriteString("</when>")
	return sb.String()
}

type xmlChoseElement struct {
	when      []xmlWhenElement
	otherwise SqlExpression
}

func (chose *xmlChoseElement) String() string {
	var sb strings.Builder
	sb.WriteString("<chose>")
	for idx := range chose.when {
		sb.WriteString(chose.when[idx].String())
	}
	if chose.otherwise != nil {
		sb.WriteString("<otherwise>")
		sb.WriteString(chose.otherwise.String())
		sb.WriteString("</otherwise>")
	}

	sb.WriteString("</chose>")
	return sb.String()
}

type xmlForEachElement struct {
	item                            string
	index                           string
	collection                      string
	openTag, separatorTag, closeTag string
	contents                        []SqlExpression
}

func (foreach *xmlForEachElement) String() string {
	var sb strings.Builder
	sb.WriteString(`<foreach collection="`)
	sb.WriteString(foreach.collection)
	sb.WriteString(`" index="`)
	sb.WriteString(foreach.index)
	sb.WriteString(`" item="`)
	sb.WriteString(foreach.item)
	sb.WriteString(`" open="`)
	sb.WriteString(foreach.openTag)
	sb.WriteString(`" separator="`)
	sb.WriteString(foreach.separatorTag)
	sb.WriteString(`" close="`)
	sb.WriteString(foreach.closeTag)
	sb.WriteString(`">`)
	for _, content := range foreach.contents {
		sb.WriteString(content.String())
	}
	sb.WriteString("</foreach>")
	return sb.String()
}

func hasXMLTag(sqlStr string) bool {
	for _, tag := range []string{"<where>", "<set>", "<chose>", "<choose>", "<if>", "<else/>", "<foreach>"} {
		if strings.Contains(sqlStr, tag) {
			return true
		}
	}

	for _, tag := range []string{
		"<if",
		"<foreach",
		"<print",
		"<pagination",
		"<else",
		"<order_by",
		"<orderBy",
		"<sort_by",
		"<sortBy",
		"<like",
		"<trim",
		"<value-range",
		"<sql",
		"<include",
	} {
		idx := strings.Index(sqlStr, tag)
		exceptIndex := idx + len(tag)
		if idx >= 0 && len(sqlStr) > exceptIndex && unicode.IsSpace(rune(sqlStr[exceptIndex])) {
			return true
		}
	}

	return false
}
