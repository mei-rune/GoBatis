package generator

import (
	"bufio"
	"reflect"
	"strings"
	"testing"

	"github.com/aryann/difflib"
)

func TestPreprocessingSQL(t *testing.T) {

	for _, test := range []struct {
		text       string
		name       string
		isNew      bool
		recordType string
		result     string
	}{

		{text: "<tablename/>", name: "s", isNew: true, recordType: "abc", result: `      var sb strings.Builder
      sb.WriteString("")
      if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&abc{})); err != nil {
        return err
      } else {
        sb.WriteString(tablename)
      }
      s := sb.String()
`},

		{text: "<tablename />", name: "s", isNew: true, recordType: "abc", result: `      var sb strings.Builder
      sb.WriteString("")
      if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&abc{})); err != nil {
        return err
      } else {
        sb.WriteString(tablename)
      }
      s := sb.String()
`},

		{text: "<tablename  alias=\"att\" />", name: "s", isNew: true, recordType: "abc", result: `      var sb strings.Builder
      sb.WriteString("")
      if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&abc{})); err != nil {
        return err
      } else {
        sb.WriteString(tablename)
      }
      sb.WriteString(" AS ")
      sb.WriteString("att")
      s := sb.String()
`},

		{text: "<tablename type=\"abcd\" alias=\"att\" />", name: "s", isNew: true, recordType: "abc", result: `      var sb strings.Builder
      sb.WriteString("")
      if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&abcd{})); err != nil {
        return err
      } else {
        sb.WriteString(tablename)
      }
      sb.WriteString(" AS ")
      sb.WriteString("att")
      s := sb.String()
`},
	} {
		result := preprocessingSQL(test.name, test.isNew, test.text, test.recordType)
		if result == test.result {
			continue
		}

		actual := splitLines(result)
		excepted := splitLines(test.result)

		if !reflect.DeepEqual(actual, excepted) {
			results := difflib.Diff(excepted, actual)
			for _, result := range results {
				t.Error(result)
			}
		}

	}
}

func splitLines(txt string) []string {
	//r := bufio.NewReader(strings.NewReader(s))
	s := bufio.NewScanner(strings.NewReader(txt))
	var ss []string
	for s.Scan() {
		ss = append(ss, s.Text())
	}
	return ss
}
