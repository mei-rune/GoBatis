package generator

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/aryann/difflib"
	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2"
	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

var (
	check       = os.Getenv("gobatis_check") == "true"
	beforeClean = os.Getenv("gobatis_clean_before_check") == "true"
	afterClean  = os.Getenv("gobatis_clean_after_check") == "true"
)

type Generator struct {
	annotationPrefix string
	tagName          string
}

func (cmd *Generator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.annotationPrefix, "annotation_prefix", "", "")
	fs.StringVar(&cmd.tagName, "tag", "xorm", "")
	return fs
}

func (cmd *Generator) Run(args []string) error {
	for _, file := range args {
		if err := cmd.runFile(file); err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (cmd *Generator) runFile(filename string) error {
	pa, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	// dir := filepath.Dir(pa)

	ctx := &goparser2.ParseContext{
		Context: astutil.NewContext(nil),
		Mapper: goparser2.TypeMapper{
			TagName: cmd.tagName,
		},
		AnnotationPrefix: cmd.annotationPrefix,
	}
	if cmd.tagName == "xorm" {
		ctx.Mapper.TagSplit = gobatis.TagSplitForXORM
	}

	file, err := goparser2.Parse(ctx, pa)
	if err != nil {
		return err
	}

	targetFile := strings.TrimSuffix(pa, ".go") + ".gobatis.go"

	if len(file.Interfaces) == 0 {
		err = os.Remove(targetFile)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	out, err := os.Create(targetFile + ".tmp")
	if err != nil {
		return err
	}
	defer func() {
		out.Close()
	}()

	if err = cmd.generateHeader(out, file); err != nil {
		return err
	}

	for _, itf := range file.Interfaces {
		if err := cmd.generateInterface(out, file, itf); err != nil {
			return err
		}
	}

	if err = out.Close(); err != nil {
		os.Remove(targetFile + ".tmp")
		return err
	}

	if beforeClean {
		exists := false
		if _, err := os.Stat(targetFile + ".old"); err == nil {
			exists = true
		}

		if exists && beforeClean {
			os.Remove(targetFile + ".old")
			exists = false
		}

		if !exists {
			if _, err := os.Stat(targetFile); err == nil {
				err = os.Rename(targetFile, targetFile+".old")
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}

	// 不知为什么，有时运行两次 goimports 才起效
	exec.Command("goimports", "-w", targetFile+".tmp").Run()
	goImports(targetFile + ".tmp")

	if check { // nolint: nestif
		if _, err := os.Stat(targetFile + ".old"); err == nil {
			actual := readFile(targetFile+".tmp", false)
			excepted := readFile(targetFile+".old", false)
			if !reflect.DeepEqual(actual, excepted) {
				fmt.Println("[ERROR]", targetFile, "failure......")
				results := difflib.Diff(excepted, actual)
				for _, result := range results {
					if result.Delta == difflib.Common {
						continue
					}

					fmt.Println(result)
				}
			} else {
				fmt.Println("[SUCC]", targetFile, " ok......")
				if afterClean {
					os.Remove(targetFile + ".old")
				}
			}
		} else {
			copyFile(targetFile+".tmp", targetFile+".old")
		}
	}

	return os.Rename(targetFile+".tmp", targetFile)
}

func copyFile(src, dst string) error {
	bs, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, bs, 0666)
}

func readFile(pa string, trimSpace bool) []string {
	bs, err := ioutil.ReadFile(pa)
	if err != nil {
		return []string{}
	}

	return splitLines(bs, trimSpace)
}

func splitLines(txt []byte, trimSpace bool) []string {
	// r := bufio.NewReader(strings.NewReader(s))
	s := bufio.NewScanner(bytes.NewReader(txt))
	var ss []string
	for s.Scan() {
		if trimSpace {
			ss = append(ss, strings.TrimSpace(s.Text()))
		} else {
			ss = append(ss, s.Text())
		}
	}
	return ss
}

func goImports(src string) error {
	cmd := exec.Command("goimports", "-w", src)
	cmd.Dir = filepath.Dir(src)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		fmt.Println("goimports -w", src)
		fmt.Println(string(out))
	}
	if err != nil {
		fmt.Println(err)
	} else if len(out) == 0 {
		fmt.Println("run `" + cmd.Path + " -w " + src + "` ok")
	}
	return err
}

func (cmd *Generator) generateHeader(out io.Writer, file *goparser2.File) error {
	io.WriteString(out, "// Please don't edit this file!\r\npackage ")
	io.WriteString(out, file.Package)
	io.WriteString(out, "\r\n\r\nimport (")
	io.WriteString(out, "\r\n\t\"errors\"")
	io.WriteString(out, "\r\n\t\"reflect\"")
	io.WriteString(out, "\r\n\t\"strings\"")
	for _, pa := range file.Imports {
		if pa == `github.com/runner-mei/GoBatis` {
			continue
		}

		if strings.HasSuffix(pa, "/errors") {
			continue
		}

		io.WriteString(out, "\r\n\t\"")
		io.WriteString(out, pa)
		io.WriteString(out, "\"")
	}
	io.WriteString(out, "\r\n\tgobatis \"github.com/runner-mei/GoBatis\"")
	io.WriteString(out, "\r\n)")
	return nil
}

func (cmd *Generator) generateInterface(out io.Writer, file *goparser2.File, itf *goparser2.Interface) error {
	if err := cmd.generateInterfaceInit(out, file, itf); err != nil {
		return err
	}

	return cmd.generateInterfaceImpl(out, file, itf)
}

func (cmd *Generator) generateInterfaceInit(out io.Writer, file *goparser2.File, itf *goparser2.Interface) error {
	io.WriteString(out, "\r\n\r\n"+`func init() {
	gobatis.Init(func(ctx *gobatis.InitContext) error {`)
	for _, m := range itf.Methods {
		if m.Config != nil && m.Config.Reference != nil {
			continue
		}

		id := itf.Name + "." + m.Name
		if itf.UseNamespace {
			id = itf.Namespace + "." + itf.Name + "." + m.Name
		}

		io.WriteString(out, "\r\n"+`	{ //// `+id)

		io.WriteString(out, "\r\n"+`		stmt, exists := ctx.Statements["`+id+`"]
		if exists {
			if stmt.IsGenerated() {
					return gobatis.ErrStatementAlreadyExists("`+id+`")
			}
		} else {`)

		//  这个函数没有什么用，只是为了分隔代码
		err := func() error {
			var recordTypeName string

			if m.Config != nil && m.Config.RecordType != "" {
				recordTypeName = m.Config.RecordType
			} else {
				recordType := itf.DetectRecordType(m, false)
				if recordType != nil {
					recordTypeName = recordType.ToLiteral()
				}
			}

			if m.Config.DefaultSQL != "" || len(m.Config.Dialects) != 0 {
				io.WriteString(out, "\r\n		")
				io.WriteString(out, preprocessingSQL("sqlStr", true, m.Config.DefaultSQL, recordTypeName))

				if len(m.Config.Dialects) > 0 {
					io.WriteString(out, "\r\n		switch ctx.Dialect {")
					for _, dialect := range m.Config.Dialects {
						io.WriteString(out, "\r\n		case "+dialect.ToGoLiteral()+":\r\n")
						io.WriteString(out, preprocessingSQL("sqlStr", false, dialect.SQL, recordTypeName))
					}
					io.WriteString(out, "\r\n}")
				}
				if m.Config.DefaultSQL == "" {
					io.WriteString(out, "\r\n		if sqlStr == \"\" {")
				}
			} else {
				if recordTypeName == "" {
					io.WriteString(out, "\r\n"+`		return errors.New("sql '`+id+`' error : statement not found - Generate SQL fail: recordType is unknown")`)
					return nil
				}
			}

			if m.Config.DefaultSQL == "" { // nolint: nestif
				args := map[string]interface{}{
					"recordTypeName": recordTypeName,
					"file":           file,
					"itf":            itf,
					"method":         m,
					"printContext":   &goparser2.PrintContext{File: file, Interface: itf},
				}
				if m.Config.DefaultSQL == "" && len(m.Config.Dialects) == 0 {
					args["var_undefined"] = true
				}

				var templateFunc *template.Template
				switch m.StatementTypeName() {
				case "insert":
					templateFunc = insertFunc
				case "upsert":
					templateFunc = insertFunc
					args["var_isUpsert"] = true
				case "update":
					templateFunc = updateFunc
				case "delete":
					templateFunc = deleteFunc
				case "select":
					if strings.Contains(m.Name, "Count") {
						templateFunc = countFunc
					} else {
						if m.Results.List[0].Type().IsExceptedType("underlyingStruct") {
							templateFunc = selectFunc
						} else if m.Results.List[0].Type().IsExceptedType("func") {
							templateFunc = selectFunc
						} else {
							io.WriteString(out, "\r\n				return errors.New(\"sql '"+id+"' error : statement not found - Generate SQL fail: sql is undefined\")")
						}
					}
				default:
					io.WriteString(out, "\r\n				return errors.New(\"sql '"+id+"' error : statement not found\")")
				}

				if templateFunc == nil {
					return nil
				}

				io.WriteString(out, "			\r\n")
				err := templateFunc.Execute(out, args)
				if err != nil {
					return errors.New("generate inteface '" + itf.Name + "' fail, " + err.Error())
				}

				if m.Config.DefaultSQL == "" && len(m.Config.Dialects) != 0 {
					io.WriteString(out, "\r\n}")
				}
			}

			io.WriteString(out, "\r\n"+`		stmt, err := gobatis.NewMapppedStatement(ctx, "`+id+`", 
				`+m.StatementGoTypeName()+`, 
				gobatis.ResultStruct, 
				sqlStr)
			if err != nil {
				return err
			}
			ctx.Statements["`+id+`"] = stmt`)
			return nil
		}()
		if err != nil {
			return err
		}

		io.WriteString(out, "\r\n    }")
		io.WriteString(out, "\r\n  }")
	}

	io.WriteString(out, "\r\n"+`return nil
		})
	}`)

	return nil
}

func (cmd *Generator) generateInterfaceImpl(out io.Writer, file *goparser2.File, itf *goparser2.Interface) error {
	printContext := &goparser2.PrintContext{File: file, Interface: itf}

	io.WriteString(out, "\r\n")
	args := map[string]interface{}{
		"file":         file,
		"itf":          itf,
		"printContext": printContext,
	}
	err := newFunc.Execute(out, args)
	if err != nil {
		return errors.New("generate inteface '" + itf.Name + "' fail, " + err.Error())
	}

	io.WriteString(out, "\r\n\r\ntype "+itf.Name+"Impl struct {")
	for _, name := range itf.EmbeddedInterfaces {
		io.WriteString(out, "\r\n  "+name)
	}
	for _, name := range itf.ReferenceInterfaces() {
		io.WriteString(out, "\r\n  "+Goify(name, false)+" "+name)
	}
	io.WriteString(out, "\r\n  session gobatis.SqlSession")
	io.WriteString(out, "\r\n}")

	for _, m := range itf.Methods {
		io.WriteString(out, "\r\n\r\nfunc (impl *"+itf.Name+"Impl) ")
		io.WriteString(out, m.MethodSignature(printContext))
		io.WriteString(out, " {")

		if m.Config != nil && m.Config.Reference != nil { // nolint: nestif
			io.WriteString(out, "\r\nreturn impl.")
			io.WriteString(out, Goify(m.Config.Reference.Interface, false))
			io.WriteString(out, ".")
			io.WriteString(out, m.Config.Reference.Method)
			io.WriteString(out, "(")

			for idx, param := range m.Params.List {
				if idx > 0 {
					io.WriteString(out, ",")
				}
				io.WriteString(out, param.Name)
				if param.IsVariadic {
					io.WriteString(out, "...")
				}
			}
			io.WriteString(out, ")")
		} else {
			args := map[string]interface{}{
				"file":         file,
				"itf":          itf,
				"printContext": &goparser2.PrintContext{File: file, Interface: itf},
				"method":       m,
			}

			foundIndex := -1
			for idx, param := range m.Params.List {
				if param.Type().IsContextType() {
					foundIndex = idx
					break
				}
			}
			if foundIndex >= 0 {
				args["contextArg"] = m.Params.List[foundIndex].Name + ","
			} else {
				args["contextArg"] = "context.Background(),"
			}

			var templateFunc *template.Template

			statementType := m.StatementTypeName()
			switch statementType {
			case "insert":
				if m.IsNotInsertID() {
					args["statementType"] = statementType
					templateFunc = selectOneImplFunc
					// {{- template "selectOne" $ | arg "method" $m | arg "statementType" $statementType}}
				} else {
					templateFunc = insertImplFunc
					// {{- template "insert" $ | arg "method" $m }}
				}
			case "upsert":
				templateFunc = insertImplFunc
				// {{- template "insert" $ | arg "method" $m }}
			case "update":
				templateFunc = updateImplFunc
				// {{- template "update" $ | arg "method" $m }}
			case "delete":
				templateFunc = deleteImplFunc
				// {{- template "delete" $ | arg "method" $m }}
			case "select":
				templateFunc = selectImplFunc(out, file, itf, m, args)

				// {{- template "select" $ | arg "method" $m }}
			default:
				io.WriteString(out, "\r\n	unknown statement type - '{{$statementType}}'")
			}

			if templateFunc != nil {
				// io.WriteString(out,"\r\n")
				err = templateFunc.Execute(out, args)
				if err != nil {
					return errors.New("generate impl for '" + itf.Name + "' fail, " + err.Error())
				}
			}
		}
		io.WriteString(out, "\r\n}")
	}
	return nil
}

func selectImplFunc(out io.Writer, file *goparser2.File, itf *goparser2.Interface, method *goparser2.Method, args map[string]interface{}) *template.Template {
	if len(method.Results.List) <= 0 {
		io.WriteString(out, "\r\n	results is empty?")
		return nil
	}

	if len(method.Results.List) == 1 { // nolint: nestif
		r1 := method.Results.List[0]

		if r1.IsCallback() || r1.IsBatchCallback() {
			return selectCallbackImplFunc
		} else if r1.IsFuncType() {
			io.WriteString(out, "\r\n	result is func, but func signature is unsupported:")
			io.WriteString(out, "\r\n	if result is batch result, then type is func(*XXX) (bool, error)")
			io.WriteString(out, "\r\n	if result is one result, then type is func(*XXX) (error)")
		} else {
			io.WriteString(out, "\r\n	results is unsupported")
		}
	} else if len(method.Results.List) == 2 {
		r1 := method.Results.List[0]
		if r1.IsBatchCallback() {
			r2 := method.Results.List[1]
			if !r2.IsCloser() {
				io.WriteString(out, "\r\n	callback results must is (func(*XXX) (bool, error), io.Closer)")
			} else {
				return selectCallbackImplFunc
			}
		} else if r1.IsFuncType() {
			io.WriteString(out, "\r\n	result is func, but func signature is unsupported:")
			io.WriteString(out, "\r\n	if result is batch result, then type is func(*XXX) (bool, error)")
			io.WriteString(out, "\r\n	if result is one result, then type is func(*XXX) (error)")
		} else {
			if r1.Type().IsMapType() {
				// recordType := itf.DetectRecordType(method, false)
				if r1.Type().IsBasicMap() {
					// {{-     template "selectBasicMap" $ | arg "scanMethod" "ScanBasicMap"}}

					args["scanMethod"] = "ScanBasicMap"
					return selectBasicMapImplFunc
				} else if strings.Contains(r1.Type().ToLiteral(), "string]interface{}") {
					// {{-     template "selectOne" $}}

					return selectOneImplFunc
				} else {
					// {{-     template "selectArray" $ | arg "scanMethod" "ScanResults"}}

					args["scanMethod"] = "ScanResults"
					return selectArrayImplFunc
				}
			} else if r1.Type().IsSliceOrArrayType() {
				// {{-   template "selectArray" $ | arg "scanMethod" "ScanSlice"  }}

				args["scanMethod"] = "ScanSlice"
				return selectArrayImplFunc
			} else {
				// {{-   template "selectOne" $}}
				return selectOneImplFunc
			}
		}
	} else if len(method.Results.List) > 2 {
		r1 := method.Results.List[0]

		errorType := false
		for i, result := range method.Results.List {
			if i == (len(method.Results.List) - 1) {
				// last
			} else if result.Type().IsExceptedType("slice") {
				if !r1.Type().IsExceptedType("slice") {
					errorType = true
					io.WriteString(out, "\r\n	"+result.Name+" isnot slice, but "+r1.Name+" is slice.")
				}
			} else {
				if r1.Type().IsExceptedType("slice") {
					errorType = true
					io.WriteString(out, "\r\n	"+result.Name+" is slice, but "+r1.Name+" isnot slice.")
				}
			}
		}

		if errorType {
			io.WriteString(out, "\r\n	results is unsupported")
		} else {
			if r1.Type().IsSliceOrArrayType() {
				// {{-   template "selectArrayForMutiObject" $}}

				return selectArrayForMutiObjectImplFunc
			} else {
				// {{-   template "selectOneForMutiObject" $}}
				return selectOneForMutiObjectImplFunc
			}
		}
	} else {
		io.WriteString(out, "\r\n	results is unsupported")
	}
	return nil
}

var funcs = template.FuncMap{
	"concat":            strings.Join,
	"containSubstr":     strings.Contains,
	"startWith":         strings.HasPrefix,
	"endWith":           strings.HasSuffix,
	"trimPrefix":        strings.TrimPrefix,
	"trimSuffix":        strings.TrimSuffix,
	"goify":             Goify,
	"underscore":        Underscore,
	"tableize":          Tableize,
	"singularize":       Singularize,
	"pluralize":         Pluralize,
	"camelizeDownFirst": CamelizeDownFirst,
	"isType": func(typ goparser2.Type, excepted string, or ...string) bool {
		return typ.IsExceptedType(excepted, or...)
	},
	// "isStructType":      goparser2.IsStructType,
	// "underlyingType":    goparser2.GetElemType,
	"argFromFunc": goparser2.ArgFromFunc,
	"typePrint": func(ctx *goparser2.PrintContext, typ goparser2.Type) string {
		return typ.ToLiteral()
	},
	"detectRecordType": func(itf *goparser2.Interface, method *goparser2.Method) *goparser2.Type {
		return itf.DetectRecordType(method, false)
	},
	"isBasicMap": func(recordType, returnType goparser2.Type) bool {
		return returnType.IsBasicMap()
	},
	"isTypeLiteral": func(name string) bool {
		return name == "_type" // || name == "typeStr"
	},
	"sub": func(a, b int) int {
		return a - b
	},
	"sum": func(a, b int) int {
		return a + b
	},
	"default": func(value, defvalue interface{}) interface{} {
		if nil == value {
			return defvalue
		}
		if s, ok := value.(string); ok && "" == s {
			return defvalue
		}
		return value
	},
	"set": func(args map[string]interface{}, name string, value interface{}) string {
		args[name] = value
		return ""
	},
	"arg": func(name string, value interface{}, args map[string]interface{}) map[string]interface{} {
		args[name] = value
		return args
	},
	"last": func(objects interface{}) interface{} {
		if objects == nil {
			return nil
		}

		rv := reflect.ValueOf(objects)
		if rv.Kind() == reflect.Array {
			if rv.Len() == 0 {
				panic(fmt.Errorf("last args is empty - %T - %v", objects, objects))
			}

			return rv.Index(rv.Len() - 1).Interface()
		}
		if rv.Kind() == reflect.Slice {
			if rv.Len() == 0 {
				panic(fmt.Errorf("last args is empty - %T - %v", objects, objects))
			}
			return rv.Index(rv.Len() - 1).Interface()
		}
		return nil
	},
	"preprocessingSQL": preprocessingSQL,
}

var insertFunc,
	updateFunc,
	deleteFunc,
	selectFunc,
	countFunc,
	newFunc,
	selectArrayForMutiObjectImplFunc,
	selectOneForMutiObjectImplFunc,
	selectBasicMapImplFunc,
	selectArrayImplFunc,
	selectOneImplFunc,
	selectCallbackImplFunc,
	deleteImplFunc,
	updateImplFunc,
	insertImplFunc *template.Template

func init() {
	for k, v := range gobatis.TemplateFuncs {
		funcs[k] = v
	}

	initInsertFunc()
	initUpdateFunc()
	initDeleteFunc()
	initCountFunc()
	initNewFunc()
	initSelectFunc()

	initSelectArrayForMutiObjectImplFunc()
	initSelectOneForMutiObjectImplFunc()
	initSelectBasicMapImplFunc()
	initSelectArrayImplFunc()
	initSelectOneImplFunc()
	initSelectCallbackImplFunc()
	initDeleteImplFunc()
	initUpdateImplFunc()
	initInsertImplFunc()
}

func initInsertFunc() {
	insertFunc = template.Must(template.New("insert").Funcs(funcs).Parse(`
	{{- set . "var_has_context" false}}
	{{- set . "var_contains_struct" false}}

	{{- range $idx, $param := .method.Params.List}}
		{{- if isType $param.Type "context"}}
			{{- set $ "var_has_context" true}}
		{{- else if and (isType $param.Type "struct") (isType $param.Type "ignoreStructs" | not)}}
			{{- set $ "var_contains_struct" true}}
		{{- end}}
	{{- end}}
	{{- if .var_has_context}}
		{{- set . "var_param_length" (sub (len .method.Params.List) 1) }}
	{{- else}}
		{{- set . "var_param_length" (len .method.Params.List) }}
	{{- end }}


  {{- /*   
    var_style  值如下
    1 为一个标准的 insertXXX(x XXX)
    2 为一个 insertXXX(f1, f2, f3, f4, f5, ...) 
    3 为一个 upsertXXX(x XXX)
  */}}
	{{- set . "var_style" ""}}
	{{- if eq .var_param_length 0 }}
	  {{- set . "var_style" "error_param_empty" }}
	{{- else if eq .var_param_length 1 }}
		{{- if or (containSubstr .method.Name "Upsert") .var_isUpsert }}
			{{- set . "var_style" "upsert"}}
	  {{- else if .var_contains_struct}}
			{{- set . "var_style" "by_struct"}}
		{{- else}}
			{{- set . "var_style" "by_arguments"}}
		{{- end}}
	{{- else }}
		{{/* - if containSubstr .method.Name "Upsert" */}}
			{{/* - set . "var_style" "upsert" */}}
    {{- if .var_contains_struct}}
	    {{- set . "var_style" "error_param_more_than_one" }}
	  {{- else}}
		  {{- if or (containSubstr .method.Name "Upsert") .var_isUpsert }}
			  {{- set . "var_style" "upsert"}}
      {{- else}}
			  {{- set . "var_style" "by_arguments"}}
      {{- end}}
		{{- end}}
	{{- end }}

	{{- if startWith .var_style "error" | not }}
		{{- $var_undefined := default .var_undefined false}}
		{{- if $var_undefined -}}
		sqlStr
		{{- else -}}
		s
		{{- end}}, err := gobatis.Generate{{if eq .var_style "upsert"}}Upsert{{else}}Insert{{end}}SQL(ctx.Dialect, ctx.Mapper, 
    reflect.TypeOf(&{{.recordTypeName}}{}),
    {{- if eq .var_style "upsert"}}
    []string{
        {{- $upsertKeys := $.method.ReadFieldNames "On" }}
        {{- /* $upsertKeys := $.method.UpsertKeys */}}
        {{- /* if eq $.var_param_length 1 */}}
          {{- /* $upsertKeys = $.method.ReadFieldNames "On" */}}
  			{{- /* end */}}
      
			{{- range $idx, $paramName := $upsertKeys}}
		       "{{$paramName}}",
			{{- end}}
		},
    {{- end}}
		[]string{
			{{- range $idx, $param := .method.Params.List}}
      
        {{- $upsertKeys := $.method.ReadFieldNames "On" }}
        {{- /* $upsertKeys := $.method.UpsertKeys */}}
        {{- /* if eq $.var_param_length 1 */}}
          {{- /* $upsertKeys = $.method.ReadFieldNames "On" */}}
  			{{- /* end */}}
      
        {{- $exists := false}}
        {{- range $idx, $a := $upsertKeys }}
            {{- if eq $a $param.Name}}
              {{- $exists = true}}
            {{- end}}
        {{- end}}
				{{- if and (isType $param.Type "context" | not) (not $exists) -}}
				   {{- if isTypeLiteral $param.Name}}
				   	"type",
				   {{- else}}
		       "{{$param.Name}}",
		       {{- end}}
			 	{{- end}}
			{{- end}}
		},
    []reflect.Type{
			{{- range $idx, $param := .method.Params.List}}
      
        {{- $upsertKeys := $.method.ReadFieldNames "On" }}
        {{- /* $upsertKeys := $.method.UpsertKeys */}}
        {{- /* if eq $.var_param_length 1 */}}
          {{- /* $upsertKeys = $.method.ReadFieldNames "On" */}}
  			{{- /* end */}}
      
        {{- $exists := false}}
        {{- range $idx, $a := $upsertKeys }}
            {{- if eq $a $param.Name}}
              {{- $exists = true}}
            {{- end}}
        {{- end}}
        
        {{- if $exists }}
        {{- else}}
    			{{- if isType $param.Type "context" | not }}
      				{{- if isType $param.Type "slice"}}
      				reflect.TypeOf({{typePrint $.printContext $param.Type}}{}),
						  {{- else if $param.IsEllipsis}}
							  reflect.TypeOf([]{{typePrint $.printContext $param.Type}}{}),
      				{{- else if isType $param.Type "ptr"}}
      				reflect.TypeOf(({{typePrint $.printContext $param.Type}})(nil)),
      				{{- else if isType $param.Type "basic"}}
      				reflect.TypeOf(new({{typePrint $.printContext $param.Type}})).Elem(),
      				{{- else if isType $param.Type "interface"}}
      				nil,
      				{{- else}}
      				reflect.TypeOf(&{{typePrint $.printContext $param.Type}}{}).Elem(),
      				{{- end}}
      		{{- end}}
    		{{- end}}
			{{- end}}
		},
		{{- if eq (len .method.Results.List) 2 -}}
	    	false
	  {{- else -}}
	    	true
	  {{- end}})
		if err != nil {
			return gobatis.ErrForGenerateStmt(err, "generate {{.itf.Name}}.{{.method.Name}} error")
		}
		{{- if not $var_undefined}}
		sqlStr = s
		{{- end}}
	{{- else}}
		{{- if eq .var_style "error_param_empty"}}
	    Please set default sql statement, param is empty!
		{{- else if eq .var_style "error_param_more_than_one"}}
	    Please set default sql statement, param is more than one!
		{{- else}}
	    Please set default sql statement!
		{{- end}}
	{{- end}}`))
}

func initUpdateFunc() {
	updateFunc = template.Must(template.New("update").Funcs(funcs).Parse(`
	{{- set . "var_first_is_context" false}}
	{{- set . "var_contains_struct" false}}
  {{- if eq (len .method.Params.List) 0 -}}
     {{.method.Name}} params is empty
  {{- end -}}
	{{- $lastParam := last .method.Params.List}}
	{{- $var_undefined := default .var_undefined false}}
	{{- $var_style_1  := and (isType $lastParam.Type "struct") (isType $lastParam.Type "ignoreStructs" | not) -}}

	{{- range $idx, $param := .method.Params.List}}
		{{- if and (isType $param.Type "context") (eq $idx 0)}}
			{{- set $ "var_first_is_context" true}}
		{{- else if and (and (isType $param.Type "struct") (isType $param.Type "ignoreStructs" | not)) (isNotLast $.method.Params.List $idx)}}
			{{- set $ "var_contains_struct" true}}

			 ignoreStructs = {{ $param.Type }} {{isType $param.Type "ignoreStructs"}}
			 isNotLast = {{ $param.Type }} {{isNotLast $.method.Params.List $idx}}
		{{- end}}
	{{- end}}

	{{- if .var_contains_struct}}
	generate update statement fail, please ....
	{{- else}}

		{{-   if $var_undefined -}}
		sqlStr
		{{- else -}}
		s
		{{- end}}, err := 

		{{- if $var_style_1 -}}
				gobatis.GenerateUpdateSQL(ctx.Dialect, ctx.Mapper, 
					"{{$lastParam.Name}}.", reflect.TypeOf(&{{.recordTypeName}}{}), 
					[]string{
					{{- range $idx, $param := .method.Params.List}}
						{{- if isType $param.Type "context" | not -}}
							{{- if lt $idx ( sub (len $.method.Params.List) 1) }}
				   		{{- if isTypeLiteral $param.Name}}
				   		"type",
				   		{{- else}}
							"{{$param.Name}}",
							{{- end}}
							{{- end}}
					 	{{- end}}
					{{- end}}
				},
				[]reflect.Type{
					{{- range $idx, $param := .method.Params.List}}
					{{- if isType $param.Type "context" | not }}
					{{- if lt $idx ( sub (len $.method.Params.List) 1) }}
						{{- if isType $param.Type "slice"}}
						reflect.TypeOf({{typePrint $.printContext $param.Type}}{}),
					  {{- else if $param.IsEllipsis}}
						reflect.TypeOf([]{{typePrint $.printContext $param.Type}}{}),
						{{- else if isType $param.Type "ptr"}}
						reflect.TypeOf(({{typePrint $.printContext $param.Type}})(nil)),
						{{- else if isType $param.Type "basic"}}
						reflect.TypeOf(new({{typePrint $.printContext $param.Type}})).Elem(),
						{{- else}}
						reflect.TypeOf(&{{typePrint $.printContext $param.Type}}{}).Elem(),
						{{- end}}
					{{- end}}
					{{- end}}
					{{- end}}
				})
		{{-  else -}}
				gobatis.GenerateUpdateSQL2(ctx.Dialect, ctx.Mapper, 
					reflect.TypeOf(&{{.recordTypeName}}{}), 
					{{- if .var_first_is_context -}}
						{{- $firstParam := index .method.Params.List 1 -}}
						reflect.TypeOf(new({{typePrint .printContext $firstParam.Type}})),
					{{- else -}}
						{{- $firstParam := index .method.Params.List 0 -}}
						reflect.TypeOf(new({{typePrint .printContext $firstParam.Type}})),
					{{- end -}}
					
					{{- if .var_first_is_context -}}
						{{- $firstParam := index .method.Params.List 1}}"{{$firstParam.Name}}",
					{{- else -}}
						{{- $firstParam := index .method.Params.List 0}}"{{$firstParam.Name}}",
					{{- end -}}
					[]string{
					{{- range $idx, $param := .method.Params.List}}
						{{- if isType $param.Type "context" | not -}}
							{{- if eq $idx 0 -}}
							{{/* 第一个是查询参数 */}}
							{{- else if and (eq $idx 1) $.var_first_is_context -}}
							{{/* 第一个是查询参数 */}}
							{{- else}}
										{{- if isTypeLiteral $param.Name}}
							   		"type",
							   		{{- else}}
										"{{$param.Name}}",
										{{- end}}
							{{- end -}}
						{{- end}}
					{{- end}}
				})
		{{-  end}}
		if err != nil {
			return gobatis.ErrForGenerateStmt(err, "generate {{.itf.Name}}.{{.method.Name}} error")
		}
		{{- if not $var_undefined}}
		sqlStr = s
		{{- end}}
	{{- end}}`))
}

func initDeleteFunc() {
	deleteFunc = template.Must(template.New("delete").Funcs(funcs).Parse(`
	{{-   $var_undefined := default .var_undefined false}}
	{{-   if $var_undefined -}}
	sqlStr
	{{- else -}}
	s
	{{- end}}, err := gobatis.GenerateDeleteSQL(ctx.Dialect, ctx.Mapper, 
	reflect.TypeOf(&{{.recordTypeName}}{}), 
		[]string{
	{{-     range $idx, $param := .method.Params.List}}
	{{-       if isType $param.Type "context" | not}}
		  {{- if isTypeLiteral $param.Name}}
   		"type",
   		{{- else}}
			"{{$param.Name}}",
			{{- end}}
	{{-       end}}
	{{-     end}}
		},
		[]reflect.Type{
	{{-     range $idx, $param := .method.Params.List}}
	{{-       if isType $param.Type "context" | not }}
	  {{- if isType $param.Type "slice"}}
		  reflect.TypeOf({{typePrint $.printContext $param.Type}}{}),
	  {{- else if $param.IsEllipsis}}
		  reflect.TypeOf([]{{typePrint $.printContext $param.Type}}{}),
	  {{- else if isType $param.Type "ptr"}}
		  reflect.TypeOf(({{typePrint $.printContext $param.Type}})(nil)),
	  {{- else if isType $param.Type "basic"}}
		  reflect.TypeOf(new({{typePrint $.printContext $param.Type}})).Elem(),
		{{- else}}
		  reflect.TypeOf(&{{typePrint $.printContext $param.Type}}{}).Elem(),
		{{- end}}
	{{-       end}}
	{{-     end}}
		},
		[]gobatis.Filter{ 
		{{- range $param := .method.Config.SQL.Filters}}
		{Expression: "{{$param.Expression}}"{{if $param.Dialect}}, Dialect: "{{$param.Dialect}}"{{end}}},
		{{- end}}
		})
	if err != nil {
		return gobatis.ErrForGenerateStmt(err, "generate {{.itf.Name}}.{{.method.Name}} error")
	}
	{{- if not $var_undefined }}
	sqlStr = s
	{{- end}}`))
}

func initCountFunc() {
	countFunc = template.Must(template.New("count").Funcs(funcs).Parse(`
	{{-   $var_undefined := default .var_undefined false}}
	{{-   if $var_undefined -}}
	sqlStr
	{{- else -}}
	s
	{{- end}}, err := gobatis.GenerateCountSQL(ctx.Dialect, ctx.Mapper, 
	reflect.TypeOf(&{{.recordTypeName}}{}), 
		[]string{
	{{-     range $idx, $param := .method.Params.List}}
	{{-       if isType $param.Type "context" | not }}
		  {{- if isTypeLiteral $param.Name}}
   		"type",
   		{{- else}}
			"{{$param.Name}}",
			{{- end}}
	{{-       end}}
	{{-     end}}
		},
		[]reflect.Type{
	{{-     range $idx, $param := .method.Params.List}}
	{{-       if isType $param.Type "context" | not }}
	  {{- if isType $param.Type "slice"}}
		  reflect.TypeOf({{typePrint $.printContext $param.Type}}{}),
	  {{- else if $param.IsEllipsis}}
		  reflect.TypeOf([]{{typePrint $.printContext $param.Type}}{}),
	  {{- else if isType $param.Type "ptr"}}
		  reflect.TypeOf(({{typePrint $.printContext $param.Type}})(nil)),
	  {{- else if isType $param.Type "basic"}}
		  reflect.TypeOf(new({{typePrint $.printContext $param.Type}})).Elem(),
		{{- else}}
		  reflect.TypeOf(&{{typePrint $.printContext $param.Type}}{}).Elem(),
		{{- end}}
	{{-       end}}
	{{-     end}}
		},
		[]gobatis.Filter{ 
		{{- range $param := .method.Config.SQL.Filters}}
		{Expression: "{{$param.Expression}}"{{if $param.Dialect}}, Dialect: "{{$param.Dialect}}"{{end}}},
		{{- end}}
		})
	if err != nil {
		return gobatis.ErrForGenerateStmt(err, "generate {{.itf.Name}}.{{.method.Name}} error")
	}
	{{- if not $var_undefined }}
	sqlStr = s
	{{- end}}`))
}

func initSelectFunc() {
	selectFunc = template.Must(template.New("select").Funcs(funcs).Parse(`
	{{-   $var_undefined := default .var_undefined false}}
	{{-   if $var_undefined -}}
	sqlStr
	{{- else -}}
	s
	{{- end}}, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper, 
	reflect.TypeOf(&{{.recordTypeName}}{}), 
		[]string{
	{{-     range $idx, $param := .method.Params.List}}
	{{-       if isType $param.Type "context" | not }}
		  {{- if isTypeLiteral $param.Name}}
   		"type",
   		{{- else}}
			"{{$param.Name}}",
			{{- end}}
	{{-       end}}
	{{-     end}}
		},
		[]reflect.Type{
	{{-     range $idx, $param := .method.Params.List}}
	{{-       if isType $param.Type "context" | not }}
	  {{- if isType $param.Type "slice"}}
		  reflect.TypeOf({{typePrint $.printContext $param.Type}}{}),
	  {{- else if $param.IsEllipsis}}
		  reflect.TypeOf([]{{typePrint $.printContext $param.Type}}{}),
	  {{- else if isType $param.Type "ptr"}}
		  reflect.TypeOf(({{typePrint $.printContext $param.Type}})(nil)),
	  {{- else if isType $param.Type "basic"}}
		  reflect.TypeOf(new({{typePrint $.printContext $param.Type}})).Elem(),
		{{- else}}
		  reflect.TypeOf(&{{typePrint $.printContext $param.Type}}{}).Elem(),
		{{- end}}
	{{-       end}}
	{{-     end}}
		},
		[]gobatis.Filter{ 
		{{- range $param := .method.Config.SQL.Filters}}
		{Expression: "{{$param.Expression}}"{{if $param.Dialect}}, Dialect: "{{$param.Dialect}}"{{end}}},
		{{- end}}
		})
	if err != nil {
		return gobatis.ErrForGenerateStmt(err, "generate {{.itf.Name}}.{{.method.Name}} error")
	}
	{{- if not $var_undefined }}
	sqlStr = s
	{{- end}}`))
}

func initNewFunc() {
	newFunc = template.Must(template.New("NewFunc").Funcs(funcs).Parse(`
	func New{{.itf.Name}}(ref gobatis.SqlSession
	{{- range $if := .itf.EmbeddedInterfaces -}}
	  , {{- goify $if false}} {{$if -}}
	{{- end -}}
	{{- range $if := .itf.ReferenceInterfaces -}}
	  , {{- goify $if false}} {{$if -}}
	{{- end -}}
		) {{.itf.Name}} {
		if ref == nil {
			panic(errors.New("param 'ref' is nil"))
		}
		if reference, ok := ref.(*gobatis.Reference); ok {
			if reference.SqlSession == nil {
				panic(errors.New("param 'ref.SqlSession' is nil"))
			} 
		} else if valueReference, ok := ref.(gobatis.Reference); ok {
			if valueReference.SqlSession == nil {
				panic(errors.New("param 'ref.SqlSession' is nil"))
			} 
		}
		return &{{.itf.Name}}Impl{	
	    {{- range $if := .itf.EmbeddedInterfaces -}}
	  		{{$if}}: {{goify $if false}},
		{{end -}}
		session: ref,
	    {{- range $if := .itf.ReferenceInterfaces}}
	  		{{goify $if false}}: {{goify $if false}}, 
		{{- end}}}
	}`))
}

// func initImpl() {
// 	implFunc = template.Must(template.New("ImplFunc").Funcs(funcs).Parse(`
// {{- define "printContext"}}
// 	{{- if .method.Params.List -}}
// 		{{- set $ "hasContextInParams" false -}}
// 	  	{{- range $param := .method.Params.List}}
// 	  	  {{- if isType $param.Type "context" -}}
// 	   		{{- $param.Name -}},
// 			{{- set $ "hasContextInParams" true -}}
// 	   	   {{- end -}}
// 	   	{{- end -}}
// 	   	{{- if not .hasContextInParams -}}
//   			context.Background(),
// 	   	{{- end -}}
//   	{{- else -}}
//   	context.Background(),
//   	{{- end -}}
// {{- end -}}
// `))
// }

func initInsertImplFunc() {
	insertImplFunc = template.Must(template.New("insertImplFunc").Funcs(funcs).Parse(`
  {{- if eq (len .method.Results.List) 2}}
  return
  {{- else -}}
	{{- $rerr := index .method.Results.List 0}}
	{{- $errName := default $rerr.Name "err"}}
	_, {{$errName}} {{if not $rerr.Name -}}:{{- end -}}=
  {{- end}} impl.session.Insert(
  	{{- .contextArg -}}
  	{{if .itf.UseNamespace }}
  	"{{.itf.Namespace}}.{{.itf.Name}}.{{.method.Name}}",
		{{- else -}}
  	"{{.itf.Name}}.{{.method.Name}}",
  	{{- end -}}
		{{- if .method.Params.List}}
		[]string{
  	{{- range $param := .method.Params.List}}
	  {{-   if isType $param.Type "context" | not }}
		  {{- if isTypeLiteral $param.Name}}
   		"type",
   		{{- else}}
			"{{$param.Name}}",
			{{- end}}
		{{-   end}}
		{{- end}}
		},
		{{- else -}}
		nil,
		{{- end -}}
		{{- if .method.Params.List}}
		[]interface{}{
			{{- range $param := .method.Params.List}}
	       {{-   if isType $param.Type "context" | not }}
				 {{$param.Name}},
			   {{- end}}
			{{- end}}
		}
		{{- else -}}
		nil
		{{- end -}}
	  {{- if ne (len .method.Results.List) 2 -}}
	  ,
	  true
	  {{- end -}}
    )

  {{- if ne (len .method.Results.List) 2}}
	{{- $rerr := index .method.Results.List 0}}
	{{- $errName := default $rerr.Name "err"}}
	return {{$errName}}
  {{- end}}`))
}

func initUpdateImplFunc() {
	updateImplFunc = template.Must(template.New("updateImplFunc").Funcs(funcs).Parse(`
	{{- if eq (len .method.Results.List) 2}}
  return
  {{- else -}}
	{{- $rerr := index .method.Results.List 0}}
	{{- $errName := default $rerr.Name "err"}}
	_, {{$errName}} {{if not $rerr.Name -}}:{{- end -}}=
  {{- end}} impl.session.Update(
  	{{- .contextArg -}}
  	{{if .itf.UseNamespace }}
  	"{{.itf.Namespace}}.{{.itf.Name}}.{{.method.Name}}",
		{{- else -}}
  	"{{.itf.Name}}.{{.method.Name}}",
  	{{- end -}}
	{{- if .method.Params.List}}
	[]string{
	{{- range $param := .method.Params.List}}
	 {{-   if isType $param.Type "context" | not }}
		  {{- if isTypeLiteral $param.Name}}
   		"type",
   		{{- else}}
			"{{$param.Name}}",
			{{- end}}
	 {{-   end -}}
	{{- end}}
	},
	{{- else -}}
	nil,
	{{- end -}}
	{{- if .method.Params.List}}
	[]interface{}{
		{{- range $param := .method.Params.List}}
	     {{-   if isType $param.Type "context" | not }}
			 {{$param.Name}},
		   {{- end}}
		{{- end}}
	}
	{{- else -}}
	nil
	{{- end -}}
	)


  {{- if ne (len .method.Results.List) 2}}
	{{- $rerr := index .method.Results.List 0}}
	{{- $errName := default $rerr.Name "err"}}
	return {{$errName}}
  {{- end}}`))
}

func initDeleteImplFunc() {
	deleteImplFunc = template.Must(template.New("deleteImplFunc").Funcs(funcs).Parse(`
	{{- if eq (len .method.Results.List) 2}}
  return
  {{- else -}}
	{{- $rerr := index .method.Results.List 0}}
	{{- $errName := default $rerr.Name "err"}}
	_, {{$errName}} {{if not $rerr.Name -}}:{{- end -}}=
  {{- end}} impl.session.Delete(
  	{{- .contextArg -}}
  	{{if .itf.UseNamespace }}
  	"{{.itf.Namespace}}.{{.itf.Name}}.{{.method.Name}}",
		{{- else -}}
  	"{{.itf.Name}}.{{.method.Name}}",
  	{{- end -}}
	{{- if .method.Params.List}}
	[]string{
	{{- range $param := .method.Params.List}}
	  {{-   if isType $param.Type "context" | not }}
		  {{- if isTypeLiteral $param.Name}}
   		"type",
   		{{- else}}
			"{{$param.Name}}",
			{{- end}}
	  {{- end}}
	{{- end}}
	},
	{{- else -}}
	nil,
	{{- end -}}
	{{- if .method.Params.List}}
	[]interface{}{
		{{- range $param := .method.Params.List}}
	    {{-   if isType $param.Type "context" | not }}
			 {{$param.Name}},
		  {{- end}}
		{{- end}}
	}
	{{- else -}}
	nil
	{{- end -}}
  )


  {{- if ne (len .method.Results.List) 2}}
	{{- $rerr := index .method.Results.List 0}}
	{{- $errName := default $rerr.Name "err"}}
	return {{$errName}}
  {{- end}}`))
}

func initSelectCallbackImplFunc() {
	selectCallbackImplFunc = template.Must(template.New("selectCallbackImplFunc").Funcs(funcs).Parse(`
    {{- $r1 := index .method.Results.List 0}}
  	{{- $isBatch := $r1.IsBatchCallback}}
  	{{- $selectMethodName := "SelectOne"}}
  	{{- $resultName := "result"}}
  	{{- if $isBatch}}
  	  {{- $selectMethodName = "Select"}}
  		{{- $resultName = "results"}}
  	{{- end}}
	{{$resultName}} := impl.session.{{$selectMethodName}}(
	  	{{- .contextArg -}}
	  	{{if .itf.UseNamespace }}
  	"{{.itf.Namespace}}.{{.itf.Name}}.{{.method.Name}}",
		{{- else -}}
  	"{{.itf.Name}}.{{.method.Name}}",
  	{{- end -}}
		{{- if .method.Params.List}}
		[]string{
		{{- range $param := .method.Params.List}}
	    {{-   if isType $param.Type "context" | not }}
			  {{- if isTypeLiteral $param.Name}}
	   		"type",
	   		{{- else}}
				"{{$param.Name}}",
				{{- end}}
		  {{- end}}
		{{- end}}
		},
		{{- else -}}
		nil,
		{{- end -}}
		{{- if .method.Params.List}}
		[]interface{}{
			{{- range $param := .method.Params.List}}
	       {{-   if isType $param.Type "context" | not }}
				 {{$param.Name}},
		     {{- end}}
			{{- end}}
		}
		{{- else -}}
		nil
		{{- end -}})

	{{- $arg := argFromFunc $r1.Type }}
	{{- $argName := default $arg.Name "value"}}
  	{{- if $isBatch}}
		return func({{$argName}} {{typePrint .printContext $arg.Type}}) (bool, error) {
			if !results.Next() {
				 if results.Err() == sql.ErrNoRows {
				 	  return false, nil
				 }
			   return false, results.Err()
			}
			return true, results.Scan({{$argName}})
	  	}{{if eq (len .method.Results.List) 2}}, results{{- end}}
  	{{- else}}
		return func({{$argName}} {{typePrint .printContext $arg.Type}}) error {
			return result.Scan({{$argName}})
	  	}
  	{{- end}}`))
}

func initSelectOneImplFunc() {
	selectOneImplFunc = template.Must(template.New("selectOneImplFunc").Funcs(funcs).Parse(`
	{{- $r1 := index .method.Results.List 0}}
	{{- $rerr := index .method.Results.List 1}}
	  
	{{- $r1Name := default $r1.Name "instance"}}
	{{- $errName := default $rerr.Name "err"}}

	{{- if not $r1.Name }}
    {{- if startWith $r1.Type.String "*"}}
		  {{- if $r1.Type.IsBasicType}}
	var instance = new({{typePrint $.printContext $r1.Type.ElemType}})
		  {{- else}}
	var instance = &{{trimPrefix ($r1.ToTypeLiteral) "*"}}{}
		  {{- end}}
    {{- else}}
	var instance {{$r1.ToTypeLiteral}}
    {{- end}}
  {{- end}}

  {{- if startWith $r1.Type.String "*" }}
		{{- if $r1.Type.ElemType.IsBasicType }}
		var nullable gobatis.Nullable
		nullable.Value = {{$r1Name}}
		{{- end }}
	{{- else }}
		{{- if isType $r1.Type "basic" }}
		var nullable gobatis.Nullable
		nullable.Value = &{{$r1Name}}
		{{- end }}
	{{- end}}

	{{- $SelectOne := "SelectOne"}}
    {{- if .method.IsNotInsertID}}
      {{- $statementType := default .statementType ""}}
      {{- if eq $statementType "insert" }} 
	    {{- $SelectOne = "InsertQuery"}}
      {{- end}}
    {{- end}}

	{{$errName}} {{if not $rerr.Name -}}:{{- end -}}= impl.session.{{$SelectOne}}(
	  	{{- .contextArg -}}
	  	{{if .itf.UseNamespace }}
  	"{{.itf.Namespace}}.{{.itf.Name}}.{{.method.Name}}",
		{{- else -}}
  	"{{.itf.Name}}.{{.method.Name}}",
  	{{- end -}}
		{{- if .method.Params.List}}
		[]string{
		{{- range $param := .method.Params.List}}
	    {{-   if isType $param.Type "context" | not }}
			  {{- if isTypeLiteral $param.Name}}
	   		"type",
	   		{{- else}}
				"{{$param.Name}}",
				{{- end}}
		  {{- end}}
		{{- end}}
		},
		{{- else -}}
		nil,
		{{- end -}}
		{{- if .method.Params.List}}
		[]interface{}{
			{{- range $param := .method.Params.List}}
	       {{-   if isType $param.Type "context" | not }}
				 {{$param.Name}},
		     {{- end}}
			{{- end}}
		}
		{{- else -}}
		nil
		{{- end -}}
	{{- if startWith $r1.Type.String "*" -}}
		{{- if $r1.Type.ElemType.IsBasicType -}}
		   ).Scan(&nullable)
		{{- else -}}
		   ).Scan({{$r1Name}})
		{{- end -}}
	{{- else -}}
		{{- if isType $r1.Type "basic" -}}
			).Scan(&nullable)
		{{- else -}}
			).Scan(&{{$r1Name}})
		{{- end -}}
	{{- end}}
  if {{$errName}} != nil {
	  {{- if isType $r1.Type "ptr"}}
    return nil, {{$errName}}
  	{{- else if isType $r1.Type "bool"}}
		  	{{- if or (startWith $.method.Name "Has") (startWith $.method.Name "Exist") (endWith $.method.Name "Exist") (endWith $.method.Name "Exists") }}
    		if {{$errName}} == sql.ErrNoRows {
    			return false, nil	
    		}
  	    {{- end}}
    		return false, {{$errName}}
  	{{- else if isType $r1.Type "numeric"}}
    return 0, {{$errName}}
  	{{- else if isType $r1.Type "string"}}
    return "", {{$errName}}
  	{{- else if isType $r1.Type "struct"}}
    return instance, {{$errName}}
  	{{- else}}
    return nil, {{$errName}}
  	{{- end}}
  }

  {{- if startWith $r1.Type.String "*"}}
		{{- if $r1.Type.ElemType.IsBasicType}}
		if !nullable.Valid {
	      {{- if isType $r1.Type "ptr"}}
		    return nil, sql.ErrNoRows
		  	{{- else if isType $r1.Type "bool"}}
		  			{{- if or (startWith $.method.Name "Has") (startWith $.method.Name "Exist") (endWith $.method.Name "Exist") (endWith $.method.Name "Exists") }}
		    		return false, nil
		  	    {{- else}}
		    		return false, sql.ErrNoRows
		        {{- end}}
		  	{{- else if isType $r1.Type "numeric"}}
		    return 0, sql.ErrNoRows
		  	{{- else if isType $r1.Type "string"}}
		    return "", sql.ErrNoRows
		  	{{- else if isType $r1.Type "struct"}}
		    return instance, sql.ErrNoRows
		  	{{- else}}
		    return nil, sql.ErrNoRows
		  	{{- end}}
		}
		{{- end}}
	{{- else}}
		{{- if isType $r1.Type "basic"}}
		if !nullable.Valid {
			  {{- if startWith $r1.Type.String "*"}}
		    return nil, sql.ErrNoRows
		  	{{- else if isType $r1.Type "bool"}}
		  	    {{- if or (startWith $.method.Name "Has") (startWith $.method.Name "Exist") (endWith $.method.Name "Exist") (endWith $.method.Name "Exists") }}
		    		return false, nil
		  	    {{- else}}
		    		return false, sql.ErrNoRows
		        {{- end}}
		  	{{- else if isType $r1.Type "numeric"}}
		    return 0, sql.ErrNoRows
		  	{{- else if isType $r1.Type "string"}}
		    return "", sql.ErrNoRows
		  	{{- else}}
		    return nil, sql.ErrNoRows
		  	{{- end}}
		}
		{{- end}}
	{{- end}}

  return {{$r1Name}}, nil`))
}

func initSelectArrayImplFunc() {
	selectArrayImplFunc = template.Must(template.New("selectArrayImplFunc").Funcs(funcs).Parse(`
  {{- $scanMethod := default .scanMethod "ScanSlice"}}
	{{- $r1 := index .method.Results.List 0}}
	{{- $rerr := index .method.Results.List 1}}

	{{- $r1Name := default $r1.Name "instances"}}
	{{- $errName := default $rerr.Name "err"}}

  {{- if not $r1.Name }}
	var instances {{$r1.ToTypeLiteral}}
	{{- end}}
    results := impl.session.Select(
	  	{{- .contextArg -}}
	  	{{if .itf.UseNamespace }}
  	"{{.itf.Namespace}}.{{.itf.Name}}.{{.method.Name}}",
		{{- else -}}
  	"{{.itf.Name}}.{{.method.Name}}",
  	{{- end -}}
		{{- if .method.Params.List}}
		[]string{
		{{- range $param := .method.Params.List}}
	    {{-   if isType $param.Type "context" | not }}
			  {{- if isTypeLiteral $param.Name}}
	   		"type",
	   		{{- else}}
				"{{$param.Name}}",
				{{- end}}
		  {{- end}}
		{{- end}}
		},
		{{- else -}}
		nil,
		{{- end -}}
		{{- if .method.Params.List}}
		[]interface{}{
			{{- range $param := .method.Params.List}}
	       {{-   if isType $param.Type "context" | not }}
				 {{$param.Name}},
		     {{- end}}
			{{- end}}
		}
		{{- else -}}
		nil
		{{- end -}}
		)
  {{$errName}} {{if not $rerr.Name -}}:{{- end -}}= results.{{$scanMethod}}(&{{$r1Name}})
  if {{$errName}} != nil {
    return nil, {{$errName}}
  }
  return {{$r1Name}}, nil`))
}

func initSelectBasicMapImplFunc() {
	selectBasicMapImplFunc = template.Must(template.New("selectBasicMapImplFunc").Funcs(funcs).Parse(`
  {{- $scanMethod := default .scanMethod "ScanBasicMap"}}
	{{- $r1 := index .method.Results.List 0}}
	{{- $rerr := index .method.Results.List 1}}

	{{- $r1Name := default $r1.Name "instances"}}
	{{- $errName := default $rerr.Name "err"}}

  {{- if not $r1.Name }}
	var {{$r1Name}} = {{$r1.ToTypeLiteral}}{}
	{{- else}}
	{{$r1Name}} = {{$r1.ToTypeLiteral}}{}
	{{- end}}

    results := impl.session.Select(
	  	{{- .contextArg -}}
	  	{{if .itf.UseNamespace }}
  	"{{.itf.Namespace}}.{{.itf.Name}}.{{.method.Name}}",
		{{- else -}}
  	"{{.itf.Name}}.{{.method.Name}}",
  	{{- end -}}
		{{- if .method.Params.List}}
		[]string{
		{{- range $param := .method.Params.List}}
	    {{-   if isType $param.Type "context" | not }}
			  {{- if isTypeLiteral $param.Name}}
	   		"type",
	   		{{- else}}
				"{{$param.Name}}",
				{{- end}}
		  {{- end}}
		{{- end}}
		},
		{{- else -}}
		nil,
		{{- end -}}
		{{- if .method.Params.List}}
		[]interface{}{
			{{- range $param := .method.Params.List}}
	      {{-   if isType $param.Type "context" | not }}
				 {{$param.Name}},
			  {{- end}}
			{{- end}}
		}
		{{- else -}}
		nil
		{{- end -}}
		)
  {{$errName}} {{if not $rerr.Name -}}:{{- end -}}= results.{{$scanMethod}}(&{{$r1Name}})
  if {{$errName}} != nil {
    return nil, {{$errName}}
  }
  return {{$r1Name}}, nil`))
}

func initSelectOneForMutiObjectImplFunc() {
	selectOneForMutiObjectImplFunc = template.Must(template.New("selectOneForMutiObjectImplFunc").Funcs(funcs).Parse(`
  {{- if eq (len .method.Results.List) 0 -}}
  {{.method.Name}} results is empty
  {{- end -}}
  
	{{- $rerr := last .method.Results.List}}
	var instance = gobatis.NewMultiple()

	{{- if and .method.Config .method.Config.Options .method.Config.Options.default_return_name}}
	instance.SetDefaultReturnName("{{.method.Config.Options.default_return_name}}")
	{{- end}}
	{{- if and .method.Config .method.Config.Options .method.Config.Options.field_delimiter}}
	instance.SetDelimiter("{{.method.Config.Options.field_delimiter}}")
	{{- end}}
	{{- range $i, $r := .method.Results.List}}
		{{- if eq $i (sub (len $.method.Results.List) 1) -}}
		{{- else}}
		{{-   if isType $r.Type "ptr"}}
		    	{{- if $r.Type.ElemType.IsBasicType}}
				  {{$r.Name}} = new({{typePrint $.printContext $r.Type.ElemType}})
				{{- else}}
				  {{$r.Name}} = &{{typePrint $.printContext $r.Type.ElemType}}{}
				{{- end}}
				instance.Set("{{$r.Name}}", {{$r.Name}}, func(ok bool) {
					if !ok {
						{{$r.Name}} = nil
					}
				})
		{{-   else}}
				instance.Set("{{$r.Name}}", &{{$r.Name}})
		{{-   end}}
		{{- end -}}
	{{- end}}

	{{$rerr.Name}} = impl.session.SelectOne(
	  	{{- .contextArg -}}
	  	{{if .itf.UseNamespace }}
  	"{{.itf.Namespace}}.{{.itf.Name}}.{{.method.Name}}",
		{{- else -}}
  	"{{.itf.Name}}.{{.method.Name}}",
  	{{- end -}}
		{{- if .method.Params.List}}
		[]string{
		{{- range $param := .method.Params.List}}
	    {{-   if isType $param.Type "context" | not }}
			  {{- if isTypeLiteral $param.Name}}
	   		"type",
	   		{{- else}}
				"{{$param.Name}}",
				{{- end}}
		  {{- end}}
		{{- end}}
		},
		{{- else -}}
		nil,
		{{- end -}}
		{{- if .method.Params.List}}
		[]interface{}{
			{{- range $param := .method.Params.List}}
	      {{-   if isType $param.Type "context" | not }}
				 {{$param.Name}},
			  {{- end}}
			{{- end}}
		}
		{{- else -}}
		nil
		{{- end -}}
		).ScanMultiple(instance)
	if {{$rerr.Name}} != nil {
		return {{range $i, $r := .method.Results.List -}}
					{{- if eq $i (sub (len $.method.Results.List) 1) -}}
						{{$rerr.Name}}
					{{- else -}}
						{{-   if isType $r.Type "ptr" -}}
						  nil,
						{{- else -}}
						  {{$r.Name}},
						{{- end -}}
					{{- end -}}
				{{- end}}
		}
		return {{range $i, $r := .method.Results.List -}}
				{{- if eq $i (sub (len $.method.Results.List) 1) -}}
						nil
				{{- else -}}
					{{$r.Name}},
				{{- end -}}
			{{- end}}`))
}

func initSelectArrayForMutiObjectImplFunc() {
	selectArrayForMutiObjectImplFunc = template.Must(template.New("selectArrayForMutiObjectImplFunc").Funcs(funcs).Parse(`
  {{- if eq (len .method.Results.List) 0 -}}
  results is empty
  {{- end -}}
  
	{{- $rerr := last .method.Results.List}}
	var instance = gobatis.NewMultipleArray()
	{{- if and .method.Config .method.Config.Options .method.Config.Options.default_return_name}}
	instance.SetDefaultReturnName("{{.method.Config.Options.default_return_name}}")
	{{- end}}
	{{- if and .method.Config .method.Config.Options .method.Config.Options.field_delimiter}}
	instance.SetDelimiter("{{.method.Config.Options.field_delimiter}}")
	{{- end}}
	{{- range $i, $r := .method.Results.List}}
		{{- if eq $i (sub (len $.method.Results.List) 1) -}}
		{{- else}}
		  {{-   if $r.Type.ElemType.IsPtrType}}
		    instance.Set("{{$r.Name}}", func(idx int) (interface{}, func(bool)) {
		    	{{- if $r.Type.ElemType.ElemType.IsBasicType }}
					newInstance := new({{typePrint $.printContext $r.Type.ElemType.ElemType}})
		    	{{- else}}
					newInstance := &{{typePrint $.printContext $r.Type.ElemType.ElemType}}{}
					{{- end}}
					return newInstance, func(ok bool) {
						if ok {
						{{$r.Name}} = append({{$r.Name}}, newInstance)
						} else {
						{{$r.Name}} = append({{$r.Name}}, nil)
						}
					}
				})
		  {{-   else}}
		    instance.Set("{{$r.Name}}", func(idx int) (interface{}, func(bool)) {
		    	{{- if $r.Type.ElemType.IsStringType}}
					{{$r.Name}} = append({{$r.Name}}, "")
		    	{{- else if  $r.Type.ElemType.IsBasicType}}
					{{$r.Name}} = append({{$r.Name}}, 0)
		    	{{- else}}
					{{$r.Name}} = append({{$r.Name}}, {{typePrint $.printContext $r.Type.ElemType}}{})
					{{- end}}
					return &{{$r.Name}}[len({{$r.Name}})-1], nil
				})
		  {{-   end}}
		{{- end -}}
	{{- end}}

	{{$rerr.Name}} = impl.session.Select(
	  	{{- .contextArg -}}
	  	{{if .itf.UseNamespace }}
  	"{{.itf.Namespace}}.{{.itf.Name}}.{{.method.Name}}",
		{{- else -}}
  	"{{.itf.Name}}.{{.method.Name}}",
  	{{- end -}}
		{{- if .method.Params.List}}
		[]string{
		{{- range $param := .method.Params.List}}
	      {{-   if isType $param.Type "context" | not }}
				  {{- if isTypeLiteral $param.Name}}
		   		"type",
		   		{{- else}}
					"{{$param.Name}}",
					{{- end}}
		    {{- end}}
		{{- end}}
		},
		{{- else -}}
		nil,
		{{- end -}}
		{{- if .method.Params.List}}
		[]interface{}{
			{{- range $param := .method.Params.List}}
	       {{-   if isType $param.Type "context" | not }}
				 {{$param.Name}},
			   {{- end}}
			{{- end}}
		}
		{{- else -}}
		nil
		{{- end -}}
		).ScanMultipleArray(instance)
  if {{$rerr.Name}} != nil {
    return {{range $i, $r := .method.Results.List -}}
				{{- if eq $i (sub (len $.method.Results.List) 1) -}}
					{{$rerr.Name}}
				{{- else -}}
					nil,
				{{- end -}}
			{{- end}}
  }
  return {{range $i, $r := .method.Results.List -}}
			{{- if eq $i (sub (len $.method.Results.List) 1) -}}
  				nil
			{{- else -}}
			    {{$r.Name}},
			{{- end -}}
		{{- end}}`))
}
