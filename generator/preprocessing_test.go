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
      if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&abc{})); err != nil {
        return err
      } else {
        sb.WriteString(tablename)
      }
      s := sb.String()
`},
		{text: "a<tablename/>", name: "s", isNew: true, recordType: "abc", result: `      var sb strings.Builder
      sb.WriteString("a")
      if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&abc{})); err != nil {
        return err
      } else {
        sb.WriteString(tablename)
      }
      s := sb.String()
`},

		{text: "<tablename />", name: "s", isNew: true, recordType: "abc", result: `      var sb strings.Builder
      if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&abc{})); err != nil {
        return err
      } else {
        sb.WriteString(tablename)
      }
      s := sb.String()
`},

		{text: "<tablename  alias=\"att\" />", name: "s", isNew: true, recordType: "abc", result: `      var sb strings.Builder
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
      if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&abcd{})); err != nil {
        return err
      } else {
        sb.WriteString(tablename)
      }
      sb.WriteString(" AS ")
      sb.WriteString("att")
      s := sb.String()
`},

		{text: "<tablename type=\"abcd\" alias=\"att\" /><tablename type=\"efgh\" alias=\"ef\" />", name: "s", isNew: true, recordType: "abc", result: `      var sb strings.Builder
      if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&abcd{})); err != nil {
        return err
      } else {
        sb.WriteString(tablename)
      }
      sb.WriteString(" AS ")
      sb.WriteString("att")
      if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&efgh{})); err != nil {
        return err
      } else {
        sb.WriteString(tablename)
      }
      sb.WriteString(" AS ")
      sb.WriteString("ef")
      s := sb.String()
`},
		{text: "<constant name=\"aa\" />", name: "s", isNew: true, recordType: "abc", result: `      var sb strings.Builder
      if cValue, ok := ctx.Config.Constants["aa"]; !ok {
        return errors.New("constant 'aa' is missing!")
      } else {
        sb.WriteString(gobatis.SqlValuePrint(cValue))
      }
      s := sb.String()
`},
	} {
		// fmt.Println("================", idx)
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
