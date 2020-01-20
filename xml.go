package gobatis

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

	return NewMapppedStatement(ctx, stmt.ID, sqlType, resultType, stmt.SQL)
}

func loadDynamicSQLFromXML(sqlStr string) (DynamicSQL, error) {
	segements, err := readSQLStatementForXML(sqlStr)
	if err != nil {
		return nil, err
	}
	return expressionArray(segements), nil
}

func readSQLStatementForXML(sqlStr string) ([]sqlExpression, error) {
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
				return readElementForXML(decoder, "")
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

func readElementForXML(decoder *xml.Decoder, tag string) ([]sqlExpression, error) {
	var sb strings.Builder
	var expressions []sqlExpression
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
				contents, err := readElementForXML(decoder, tag+"/if")
				if err != nil {
					return nil, err
				}
				if len(contents) == 0 {
					break
				}

				var content sqlExpression
				if len(contents) == 1 {
					content = contents[0]
				} else if len(contents) > 1 {
					content = expressionArray(contents)
				}
				segement, err := newIFExpression(readElementAttrForXML(el.Attr, "test"), content)
				if err != nil {
					return nil, err
				}
				expressions = append(expressions, segement)
			case "foreach":
				contents, err := readElementForXML(decoder, tag+"/foreach")
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
			case "chose":
				choseEl, err := loadChoseElementForXML(decoder, tag+"/chose")
				if err != nil {
					return nil, err
				}
				chose, err := newChoseExpression(*choseEl)
				if err != nil {
					return nil, err
				}
				expressions = append(expressions, chose)
			case "where":
				array, err := readElementForXML(decoder, tag+"/where")
				if err != nil {
					return nil, err
				}

				expressions = append(expressions, &whereExpression{expressions: array})
			case "set":
				array, err := readElementForXML(decoder, tag+"/set")
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
				printExpr := &printExpression{
					prefix: prefix,
					value:  value,
					fmt:    readElementAttrForXML(el.Attr, "fmt")}
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
				likeExpr := &likeExpression{
					prefix: prefix,
					value:  value}
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
				pagination := &paginationExpression{
					offset: readElementAttrForXML(el.Attr, "offset"),
					limit:  readElementAttrForXML(el.Attr, "limit")}
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
					sort: readElementAttrForXML(el.Attr, "sort")}
				if orderBy.sort == "" {
					orderBy.sort = readElementAttrForXML(el.Attr, "by")
				}
				expressions = append(expressions, orderBy)
			case "trim":
				array, err := readElementForXML(decoder, tag+"/trim")
				if err != nil {
					return nil, err
				}
				if len(array) == 0 {
					return nil, errors.New("element trim must isnot empty element")
				}

				trimExpr := &trimExpression{
					prefix:      readElementAttrForXML(el.Attr, "prefix"),
					suffix:      readElementAttrForXML(el.Attr, "suffix"),
					expressions: array}
				if prefixoverride := readElementAttrForXML(el.Attr, "prefixOverrides"); prefixoverride != "" {
					expr, err := newRawExpression(prefixoverride)
					if err != nil {
						return nil, errors.New("element trim.prefixOverrides is invalid - '" + prefixoverride + "'")
					}
					trimExpr.prefixoverride = expr
				}

				if suffixoverride := readElementAttrForXML(el.Attr, "suffixOverrides"); suffixoverride != "" {
					expr, err := newRawExpression(suffixoverride)
					if err != nil {
						return nil, errors.New("element trim.suffixOverrides is invalid - '" + suffixoverride + "'")
					}
					trimExpr.suffixoverride = expr
				}
				expressions = append(expressions, trimExpr)
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
			lastPrint = nil

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

func loadChoseElementForXML(decoder *xml.Decoder, tag string) (*xmlChoseElement, error) {
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
				contents, err := readElementForXML(decoder, "when")
				if err != nil {
					return nil, err
				}

				if len(contents) == 0 {
					segement.when = append(segement.when, xmlWhenElement{content: rawString(""),
						test: readElementAttrForXML(el.Attr, "test")})
					break
				}

				var content sqlExpression
				if len(contents) == 1 {
					content = contents[0]
				} else if len(contents) > 1 {
					content = expressionArray(contents)
				}

				segement.when = append(segement.when, xmlWhenElement{content: content,
					test: readElementAttrForXML(el.Attr, "test")})
				break
			}

			if el.Name.Local == "otherwise" {
				contents, err := readElementForXML(decoder, "otherwise")
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
	content sqlExpression
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
	otherwise sqlExpression
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
	contents                        []sqlExpression
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
	for _, tag := range []string{"<where>", "<set>", "<chose>", "<if>", "<foreach>"} {
		if strings.Contains(sqlStr, tag) {
			return true
		}
	}

	for _, tag := range []string{"<if", "<foreach", "<print", "<pagination", "<order_by", "<like", "<trim"} {
		idx := strings.Index(sqlStr, tag)
		exceptIndex := idx + len(tag)
		if idx >= 0 && len(sqlStr) > exceptIndex && unicode.IsSpace(rune(sqlStr[exceptIndex])) {
			return true
		}
	}

	return false
}
