package core

import (
	"fmt"
	"strings"
)

const tagPrefix = "db"

var xormkeyTags = map[string]struct{}{
	"pk":         {},
	"autoincr":   {},
	"null":       {},
	"notnull":    {},
	"unique":     {},
	"extends":    {},
	"index":      {},
	"<-":         {},
	"->":         {},
	"created":    {},
	"updated":    {},
	"deleted":    {},
	"version":    {},
	"default":    {},
	"json":       {},
	"jsonarray":  {},
	"jsonb":      {},
	"bit":        {},
	"tinyint":    {},
	"smallint":   {},
	"mediumint":  {},
	"int":        {},
	"integer":    {},
	"bigint":     {},
	"str":        {},
	"char":       {},
	"varchar":    {},
	"tinytext":   {},
	"text":       {},
	"mediumtext": {},
	"longtext":   {},
	"binary":     {},
	"varbinary":  {},
	"date":       {},
	"datetime":   {},
	"time":       {},
	"timestamp":  {},
	"timestampz": {},
	"real":       {},
	"float":      {},
	"double":     {},
	"decimal":    {},
	"numeric":    {},
	"tinyblob":   {},
	"clob":       {},
	"blob":       {},
	"mediumblob": {},
	"longblob":   {},
	"bytea":      {},
	"bool":       {},
	"serial":     {},
}

func TagSplitForXORM(s string, fieldName string) []string {
	parts, err := splitXormTag(s)
	if err != nil {
		panic(err)
	}
	if len(parts) == 0 {
		return []string{fieldName}
	}

	copyed := make([]string, 1, len(parts))
	for i := 0; i < len(parts); i++ {
		name := parts[i].name

		if len(parts[i].params) > 0 {
			// unique(xxxx) 改成 unique=xxxx
			copyed = append(copyed, name+"="+strings.Join(parts[i].params, ","))
			continue
		}

		if _, ok := xormkeyTags[name]; !ok {
			if copyed[0] == "" {
				copyed[0] = strings.Trim(name, "'")
				continue
			}
		}

		copyed = append(copyed, name)
	}
	if copyed[0] == "" {
		copyed[0] = fieldName
	}
	return copyed
}

var TagSplitForDb = func(s string, fieldName string) []string {
	return strings.Split(s, ",")
}

type tag struct {
	name   string
	params []string
}

func splitXormTag(tagStr string) ([]tag, error) {
	tagStr = strings.TrimSpace(tagStr)
	var (
		inQuote    bool
		inBigQuote bool
		lastIdx    int
		curTag     tag
		paramStart int
		tags       []tag
	)
	for i, t := range tagStr {
		switch t {
		case '\'':
			inQuote = !inQuote
		case ' ':
			if !inQuote && !inBigQuote {
				if lastIdx < i {
					if curTag.name == "" {
						curTag.name = tagStr[lastIdx:i]
					}
					tags = append(tags, curTag)
					lastIdx = i + 1
					curTag = tag{}
				} else if lastIdx == i {
					lastIdx = i + 1
				}
			} else if inBigQuote && !inQuote {
				paramStart = i + 1
			}
		case ',':
			if !inQuote && !inBigQuote {
				return nil, fmt.Errorf("comma[%d] of %s should be in quote or big quote", i, tagStr)
			}
			if !inQuote && inBigQuote {
				curTag.params = append(curTag.params, strings.TrimSpace(tagStr[paramStart:i]))
				paramStart = i + 1
			}
		case '(':
			inBigQuote = true
			if !inQuote {
				curTag.name = tagStr[lastIdx:i]
				paramStart = i + 1
			}
		case ')':
			inBigQuote = false
			if !inQuote {
				curTag.params = append(curTag.params, tagStr[paramStart:i])
			}
		}
	}
	if lastIdx < len(tagStr) {
		if curTag.name == "" {
			curTag.name = tagStr[lastIdx:]
		}
		tags = append(tags, curTag)
	}
	return tags, nil
}
