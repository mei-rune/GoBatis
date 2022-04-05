package generator

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/types"
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
	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser"
)

var check = os.Getenv("gobatis_check") == "true"
var beforeClean = os.Getenv("gobatis_clean_before_check") == "true"
var afterClean = os.Getenv("gobatis_clean_after_check") == "true"

type Generator struct {
	tagName string
}

func (cmd *Generator) Flags(fs *flag.FlagSet) *flag.FlagSet {
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
	//dir := filepath.Dir(pa)

	ctx := &goparser.ParseContext{
		Mapper: goparser.TypeMapper{
			TagName: cmd.tagName,
		},
	}
	if cmd.tagName == "xorm" {
		ctx.Mapper.TagSplit = gobatis.TagSplitForXORM
	}

	file, err := goparser.Parse(pa, ctx)
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

	if check {
		if _, err := os.Stat(targetFile); err == nil {
			exists := false
			if _, err := os.Stat(targetFile + ".old"); err == nil {
				exists = true
			}

			if exists && beforeClean {
				os.Remove(targetFile + ".old")
				exists = false
			}

			if !exists {
				err = os.Rename(targetFile, targetFile+".old")
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}

	// 不知为什么，有时运行两次 goimports 才起效
	exec.Command("goimports", "-w", targetFile).Run()
	goImports(targetFile)

	if check {
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
			copyFile(targetFile, targetFile+".old")
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
	//r := bufio.NewReader(strings.NewReader(s))
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

func (cmd *Generator) generateHeader(out io.Writer, file *goparser.File) error {
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
	io.WriteString(out, "\r\n)\r\n")
	return nil
}

func (cmd *Generator) generateInterface(out io.Writer, file *goparser.File, itf *goparser.Interface) error {
	args := map[string]interface{}{"file": file, "itf": itf,
		"printContext": &goparser.PrintContext{File: file, Interface: itf}}
	err := newFunc.Execute(out, args)
	if err != nil {
		return errors.New("generate inteface '" + itf.Name + "' fail, " + err.Error())
	}
	err = implFunc.Execute(out, args)
	if err != nil {
		return errors.New("generate impl for '" + itf.Name + "' fail, " + err.Error())
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
	"isType":            isExceptedType,
	"isStructType":      goparser.IsStructType,
	"underlyingType":    goparser.GetElemType,
	"argFromFunc":       goparser.ArgFromFunc,
	"typePrint": func(ctx *goparser.PrintContext, typ types.Type) string {
		return goparser.PrintType(ctx, typ, false)
	},
	"detectRecordType": func(itf *goparser.Interface, method *goparser.Method) types.Type {
		return itf.DetectRecordType(method)
	},
	"isBasicMap": func(recordType, returnType types.Type) bool {
		// keyType := getKeyType(recordType)

		for {
			if ptr, ok := returnType.(*types.Pointer); !ok {
				break
			} else {
				returnType = ptr.Elem()
			}
		}

		mapType, ok := returnType.(*types.Map)
		if !ok {
			return false
		}

		elemType := mapType.Elem()
		for {
			if ptr, ok := elemType.(*types.Pointer); !ok {
				break
			} else {
				elemType = ptr.Elem()
			}
		}

		if _, ok := elemType.(*types.Basic); ok {
			return true
		}

		switch elemType.String() {
		case "time.Time", "net.IP", "net.HardwareAddr":
			return true
		}
		return false
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

var newFunc, implFunc *template.Template

func init() {
	for k, v := range gobatis.TemplateFuncs {
		funcs[k] = v
	}

	newFunc = template.Must(template.New("NewFunc").Funcs(funcs).Parse(`
{{- define "insert"}}
	{{- set . "var_has_context" false}}
	{{- set . "var_contains_struct" false}}

	{{- range $idx, $param := .method.Params.List}}
		{{- if isType $param.Type "context"}}
			{{- set $ "var_has_context" true}}
		{{- else if isType $param.Type "struct"}}
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
		{{- if $var_undefined}}
		sqlStr
		{{- else}}
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
	{{- end}}
{{- end}}

{{- define "update"}}
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
		{{- else if and (isType $param.Type "struct") (isNotLast $.method.Params.List $idx)}}
			{{- set $ "var_contains_struct" true}}
		{{- end}}
	{{- end}}

	{{- if .var_contains_struct}}
	generate update statement fail, please ....
	{{- else}}

		{{-   if $var_undefined }}
		sqlStr
		{{- else}}
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
	{{- end}}
{{- end}}

{{- define "delete"}}
	{{-   $var_undefined := default .var_undefined false}}
	{{-   if $var_undefined }}
	sqlStr
	{{- else}}
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
	{{- end}}
{{- end}}

{{- define "count"}}
	{{-   $var_undefined := default .var_undefined false}}
	{{-   if $var_undefined }}
	sqlStr
	{{- else}}
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
	{{- end}}
{{- end}}

{{- define "select"}}
	{{-   $var_undefined := default .var_undefined false}}
	{{-   if $var_undefined}}
	sqlStr
	{{- else}}
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
	{{- end}}
{{- end}}

{{- define "genSQL"}}
  {{- set . "var_isUpsert" false}}
  
  {{- if .recordTypeName}}
    {{- $statementType := .method.StatementTypeName}}
	  {{- if eq $statementType "insert"}}
	  {{-   template "insert" . | arg "recordTypeName" .recordTypeName}}
	  {{- else if eq $statementType "upsert"}}
	  {{-   template "insert" . | arg "recordTypeName" .recordTypeName | arg "var_isUpsert" true}}
	  {{- else if eq $statementType "update"}}
	  {{-   template "update" . | arg "recordTypeName" .recordTypeName}}
	  {{- else if eq $statementType "delete"}}
    {{-   template "delete" . | arg "recordTypeName" .recordTypeName}}
	  {{- else if eq $statementType "select"}}
	  {{-   if containSubstr .method.Name "Count" }}
	  {{-     template "count" . | arg "recordTypeName" .recordTypeName}}
	  {{-   else}}
		{{-     $r1 := index .method.Results.List 0}}
		{{-     if isType $r1.Type "underlyingStruct"}}
	  {{-       template "select" . | arg "recordTypeName" .recordTypeName}}
	  {{-     else if isType $r1.Type "func"}}
	  {{-       template "select" . | arg "recordTypeName" .recordTypeName}}
	  {{-     else}}
              {{- set . "genError" true}}
	  	        return errors.New("sql '{{.itf.Name}}.{{.method.Name}}' error : statement not found - Generate SQL fail: sql is undefined")
	  {{-     end}}
	  {{-   end}}
	  {{- else}}
    {{- set . "genError" true}}
	  return errors.New("sql '{{.itf.Name}}.{{.method.Name}}' error : statement not found ")
	  {{- end}}
  {{- else}}
        {{- set . "genError" true}}
        return errors.New("sql '{{.itf.Name}}.{{.method.Name}}' error : statement not found - Generate SQL fail: recordType is unknown")
  {{- end}}
{{- end}}

{{- define "registerStmt"}}
stmt, err := gobatis.NewMapppedStatement(ctx, "{{.itf.Name}}.{{.method.Name}}", 
	{{.method.StatementGoTypeName}}, 
	gobatis.ResultStruct, 
	sqlStr)
if err != nil {
	return err
}
ctx.Statements["{{.itf.Name}}.{{.method.Name}}"] = stmt
{{- end}}


func init() {
	gobatis.Init(func(ctx *gobatis.InitContext) error {
	{{- range $m := .itf.Methods}}
	{{-   if and $m.Config $m.Config.Reference}}
	{{-   else}}
	{ //// {{$.itf.Name}}.{{$m.Name}}
		stmt, exists := ctx.Statements["{{$.itf.Name}}.{{$m.Name}}"]
		if exists {
			if stmt.IsGenerated() {
					return gobatis.ErrStatementAlreadyExists("{{$.itf.Name}}.{{$m.Name}}")
			}
		} else {
    {{- set $ "genError" false}}
    {{- set $ "var_undefined" false}}
	  {{- set $ "recordTypeName" ""}}
  
	  {{- if and $m.Config $m.Config.RecordType}}
      {{- set $ "recordTypeName" $m.Config.RecordType}}
    {{- else}}
    	  {{- $recordType := detectRecordType $.itf $m}}
    	  {{- if $recordType}}
    	    {{- set $ "recordTypeName" (typePrint $.printContext $recordType)}}
    	  {{- end}}
	  {{- end}}

		{{-   if or $m.Config.DefaultSQL  $m.Config.Dialects}}
		  {{preprocessingSQL "sqlStr" true $m.Config.DefaultSQL $.recordTypeName }}
			{{-     if $m.Config.Dialects}}
			switch ctx.Dialect {
				{{-    range $typ, $dialect := $m.Config.Dialects}}
			case gobatis.NewDialect("{{$typ}}"):
		  	{{preprocessingSQL "sqlStr" false $dialect $.recordTypeName }}
				{{-    end}}
			}

			{{-     end}}
			{{- if not $m.Config.DefaultSQL}}
			if sqlStr == "" {	
			   {{- template "genSQL" $ | arg "method" $m }}
			}
			{{- end}}
		{{- else}}
			{{- template "genSQL" $ | arg "method" $m | arg "var_undefined" true}}
		{{- end}}
		{{- $genError := default $.genError false}}
		{{- if not $genError }}
			{{- template "registerStmt" $ | arg "method" $m}}
		{{- end}}
		}
	}
	{{-   end}}
	{{- end}}
	return nil
	})
}

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

	implFunc = template.Must(template.New("ImplFunc").Funcs(funcs).Parse(`
{{- define "printContext"}}
	{{- if .method.Params.List -}}
		{{- set $ "hasContextInParams" false -}}
	  	{{- range $param := .method.Params.List}}
	  	  {{- if isType $param.Type "context" -}}
	   		{{- $param.Name -}},
			{{- set $ "hasContextInParams" true -}}
	   	   {{- end -}}
	   	{{- end -}}
	   	{{- if not .hasContextInParams -}}
  			context.Background(),
	   	{{- end -}}
  	{{- else -}}
  	context.Background(),
  	{{- end -}}
{{- end -}}
{{- define "insert"}}
  {{- if eq (len .method.Results.List) 2}}
  return
  {{- else -}}
	{{- $rerr := index .method.Results.List 0}}
	{{- $errName := default $rerr.Name "err"}}
	_, {{$errName}} {{if not $rerr.Name -}}:{{- end -}}=
  {{- end}} impl.session.Insert(
  	{{- template "printContext" . -}}
  	"{{.itf.Name}}.{{.method.Name}}",
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
  {{- end}}

{{- end}}

{{- define "update"}}
{{- if eq (len .method.Results.List) 2}}
  return
  {{- else -}}
	{{- $rerr := index .method.Results.List 0}}
	{{- $errName := default $rerr.Name "err"}}
	_, {{$errName}} {{if not $rerr.Name -}}:{{- end -}}=
  {{- end}} impl.session.Update(
  	{{- template "printContext" . -}}
  	"{{.itf.Name}}.{{.method.Name}}",
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
  {{- end}}
{{- end}}

{{- define "delete"}}
{{- if eq (len .method.Results.List) 2}}
  return
  {{- else -}}
	{{- $rerr := index .method.Results.List 0}}
	{{- $errName := default $rerr.Name "err"}}
	_, {{$errName}} {{if not $rerr.Name -}}:{{- end -}}=
  {{- end}} impl.session.Delete(
  	{{- template "printContext" . -}}
  	"{{.itf.Name}}.{{.method.Name}}",
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
  {{- end}}
{{- end}}


{{- define "selectCallback"}}
    {{- $r1 := index .method.Results.List 0}}
  	{{- $isBatch := $r1.IsBatchCallback}}
  	{{- $selectMethodName := "SelectOne"}}
  	{{- $resultName := "result"}}
  	{{- if $isBatch}}
  	  {{- $selectMethodName = "Select"}}
  		{{- $resultName = "results"}}
  	{{- end}}
	{{$resultName}} := impl.session.{{$selectMethodName}}(
	  	{{- template "printContext" . -}}
	  	"{{.itf.Name}}.{{.method.Name}}",
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
  	{{- end}}
{{- end}}

{{- define "selectOne"}}
	{{- $r1 := index .method.Results.List 0}}
	{{- $rerr := index .method.Results.List 1}}
	  
	{{- $r1Name := default $r1.Name "instance"}}
	{{- $errName := default $rerr.Name "err"}}

	{{- if not $r1.Name }}
    {{- if startWith $r1.Type.String "*"}}
		  {{- if isType $r1.Type.Elem "basic"}}
	var instance = new({{typePrint $.printContext $r1.Type.Elem}})
		  {{- else}}
	var instance = &{{trimPrefix ($r1.PrintTypeToConsole .printContext) "*"}}{}
		  {{- end}}
    {{- else}}
	var instance {{$r1.PrintTypeToConsole .printContext}}
    {{- end}}
  {{- end}}

  {{- if startWith $r1.Type.String "*" }}
		{{- if isType $r1.Type.Elem "basic" }}
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
	  	{{- template "printContext" . -}}
	  	"{{.itf.Name}}.{{.method.Name}}",
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
		{{- if isType $r1.Type.Elem "basic" -}}
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
		{{- if isType $r1.Type.Elem "basic"}}
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

  return {{$r1Name}}, nil
{{- end}}

{{- define "selectArray"}}
  {{- $scanMethod := default .scanMethod "ScanSlice"}}
	{{- $r1 := index .method.Results.List 0}}
	{{- $rerr := index .method.Results.List 1}}

	{{- $r1Name := default $r1.Name "instances"}}
	{{- $errName := default $rerr.Name "err"}}

  {{- if not $r1.Name }}
	var instances {{$r1.PrintTypeToConsole .printContext}}
	{{- end}}
    results := impl.session.Select(
	  	{{- template "printContext" . -}}
	  	"{{.itf.Name}}.{{.method.Name}}",
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
  return {{$r1Name}}, nil
{{- end}}


{{- define "selectBasicMap"}}
  {{- $scanMethod := default .scanMethod "ScanBasicMap"}}
	{{- $r1 := index .method.Results.List 0}}
	{{- $rerr := index .method.Results.List 1}}

	{{- $r1Name := default $r1.Name "instances"}}
	{{- $errName := default $rerr.Name "err"}}

  {{- if not $r1.Name }}
	var {{$r1Name}} = {{$r1.PrintTypeToConsole .printContext}}{}
	{{- else}}
	{{$r1Name}} = {{$r1.PrintTypeToConsole .printContext}}{}
	{{- end}}

    results := impl.session.Select(
	  	{{- template "printContext" . -}}
	  	"{{.itf.Name}}.{{.method.Name}}",
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
  return {{$r1Name}}, nil
{{- end}}



{{- define "selectOneForMutiObject"}}

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
		    	{{- if isType $r.Type.Elem "basic"}}
				  {{$r.Name}} = new({{typePrint $.printContext $r.Type.Elem}})
				{{- else}}
				  {{$r.Name}} = &{{typePrint $.printContext $r.Type.Elem}}{}
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
	  	{{- template "printContext" . -}}
	  	"{{.itf.Name}}.{{.method.Name}}",
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
			{{- end}}
{{- end}}


{{- define "selectArrayForMutiObject"}}
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
		  {{-   if isType $r.Type.Elem "ptr"}}
		    instance.Set("{{$r.Name}}", func(idx int) (interface{}, func(bool)) {
		    	{{- if isType $r.Type.Elem.Elem "basic"}}
					newInstance := new({{typePrint $.printContext $r.Type.Elem.Elem}})
		    	{{- else}}
					newInstance := &{{typePrint $.printContext $r.Type.Elem.Elem}}{}
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
		    	{{- if isType $r.Type.Elem "string"}}
					{{$r.Name}} = append({{$r.Name}}, "")
		    	{{- else if isType $r.Type.Elem "basic"}}
					{{$r.Name}} = append({{$r.Name}}, 0)
		    	{{- else}}
					{{$r.Name}} = append({{$r.Name}}, {{typePrint $.printContext $r.Type.Elem}}{})
					{{- end}}
					return &{{$r.Name}}[len({{$r.Name}})-1], nil
				})
		  {{-   end}}
		{{- end -}}
	{{- end}}

	{{$rerr.Name}} = impl.session.Select(
	  	{{- template "printContext" . -}}
	  	"{{.itf.Name}}.{{.method.Name}}",
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
		{{- end}}
{{- end}}

{{- define "select"}}
  {{- if .method.Results}}
    {{- if eq (len .method.Results.List) 1}}
		{{- $r1 := index .method.Results.List 0}}
		{{- if or ($r1.IsCallback) ($r1.IsBatchCallback) }}
	    	{{- template "selectCallback" $}}
		{{- else if $r1.IsFunc }}
	result is func, but func signature is unsupported:
	if result is batch result, then type is func(*XXX) (bool, error)
	if result is one result, then type is func(*XXX) (error) 
		{{- else}}
	results is unsupported
		{{- end}}
    {{- else if eq (len .method.Results.List) 2}}
	    {{- $r1 := index .method.Results.List 0}}

		{{- if $r1.IsBatchCallback }}

	    	{{- $r2 := index .method.Results.List 1}}
	    	{{- if not $r2.IsCloser }}
				callback results must is (func(*XXX) (bool, error), io.Closer) 
			{{- else}}
	    		{{- template "selectCallback" $}}
	    	{{- end}}

		{{- else if $r1.IsFunc }}
			result is func, but func signature is unsupported:
			if result is batch result, then type is func(*XXX) (bool, error)
			if result is one result, then type is func(*XXX) (error) 
		{{- else}}

		    {{- if startWith $r1.Type.String "map["}}
	  		{{-   $recordType := detectRecordType .itf .method}}
		    {{-   if isBasicMap $recordType $r1.Type}}
		    {{-     template "selectBasicMap" $ | arg "scanMethod" "ScanBasicMap"}}
		    {{-   else if containSubstr $r1.Type.String "string]interface{}"}}
		    {{-     template "selectOne" $}}
		    {{-   else}}
		    {{-     template "selectArray" $ | arg "scanMethod" "ScanResults"}}
		    {{-   end}}
		    {{- else if containSubstr $r1.Type.String "[]"}}
		    {{-   template "selectArray" $ | arg "scanMethod" "ScanSlice"  }}
		    {{- else}}
		    {{-   template "selectOne" $}}
		    {{- end}}

		{{- end}}
	{{- else if gt (len .method.Results.List) 2}}

		{{- $r1 := index .method.Results.List 0}}
		{{- if isType $r1.Type "slice"}}
		{{- set $ "sliceInResults" true}}
		{{- end}}

		{{- set $ "errorType" false}}
		{{- range $i, $result := .method.Results.List}}
			{{- if eq $i (sub (len $.method.Results.List) 1) }}
			{{- else if isType $result.Type "slice"}}
					{{- if not $.sliceInResults }}
					{{- set $ "errorType" true}}
				  {{- $result.Name}} isnot slice, but {{$r1.Name}} is slice.
				  {{- end}}
			{{- else}}
			    {{- if $.sliceInResults }}
					{{- set $ "errorType" true}}
				  {{- $result.Name}} is slice, but {{$r1.Name}} isnot slice.
				  {{- end}}
			{{- end}}
		{{- end}}

		{{- if .errorType}}
			results is unsupported
		{{- else}}
			{{- if containSubstr $r1.Type.String "[]"}}
			{{-   template "selectArrayForMutiObject" $}}
			{{- else}}
			{{-   template "selectOneForMutiObject" $}}
			{{- end}}
		{{- end}}
    {{- else}}
    results is unsupported
    {{- end}}
  {{- else}}
    results is empty?
  {{- end}}
{{- end}}

type {{.itf.Name}}Impl struct {
{{- range $if := .itf.EmbeddedInterfaces}}
  {{$if}}
{{- end}}
{{- range $if := .itf.ReferenceInterfaces}}
  {{goify $if false}} {{$if}}
{{- end}}
	session gobatis.SqlSession
}
{{ range $m := .itf.Methods}}
func (impl *{{$.itf.Name}}Impl) {{$m.MethodSignature $.printContext}} {
	{{- if and $m.Config $m.Config.Reference}}
   return impl.{{goify $m.Config.Reference.Interface false}}.{{$m.Config.Reference.Method -}}(
   	{{- range $idx, $param := $m.Params.List -}}
   		{{- $param.Name}}{{if $param.IsVariadic}}...{{end}} {{- if ne $idx ( sub (len $m.Params.List) 1) -}},{{- end -}}
   	{{- end -}})
	{{- else}}
		{{- $statementType := $m.StatementTypeName}}
		{{- if eq $statementType "insert"}}
		  {{- if $m.IsNotInsertID}}
		    {{- template "selectOne" $ | arg "method" $m | arg "statementType" $statementType}}
		  {{- else}}
		    {{- template "insert" $ | arg "method" $m }}
		  {{- end}}
		  {{- set $ "statementType" ""}}
		{{- else if eq $statementType "upsert"}}
		{{- template "insert" $ | arg "method" $m }}
		{{- else if eq $statementType "update"}}
		{{- template "update" $ | arg "method" $m }}
		{{- else if eq $statementType "delete"}}
		{{- template "delete" $ | arg "method" $m }}
		{{- else if eq $statementType "select"}}
		{{- template "select" $ | arg "method" $m }}
		{{- else}}
		    unknown statement type - '{{$statementType}}'
		{{- end}}
	{{- end}}
}
{{end}}
`))
}

func isExceptedType(typ types.Type, excepted string, or ...string) bool {
	if ptr, ok := typ.(*types.Pointer); ok {
		if excepted == "ptr" {
			return true
		}
		for _, name := range or {
			if name == "ptr" {
				return true
			}
		}
		return isExceptedType(ptr.Elem(), excepted, or...)
	}
	for _, name := range append([]string{excepted}, or...) {
		switch name {
		case "func":
			if _, ok := typ.(*types.Signature); ok {
				return true
			}
		case "context":
			if named, ok := typ.(*types.Named); ok {
				if named.Obj().Name() == "Context" {
					return true
				}
			}
		case "ptr":
		case "error":
			if named, ok := typ.(*types.Named); ok {
				if named.Obj().Name() == "error" {
					return true
				}
			}
		case "ignoreStructs":
			if named, ok := typ.(*types.Named); ok {
				if _, ok := named.Underlying().(*types.Struct); ok {
					typName := named.Obj().Pkg().Name() + "." + named.Obj().Name()
					for _, nm := range goparser.IgnoreStructs {
						if nm == typName {
							return true
						}
					}
					return false
				}
			}
		case "underlyingStruct":
			var exp = typ
			if slice, ok := exp.(*types.Slice); ok {
				exp = slice.Elem()
			}
			if mapTyp, ok := exp.(*types.Map); ok {
				exp = mapTyp.Elem()
			}
			if ptrTyp, ok := exp.(*types.Pointer); ok {
				exp = ptrTyp.Elem()
			}

			if named, ok := exp.(*types.Named); ok {
				exp = named.Underlying()

				typName := named.Obj().Pkg().Name() + "." + named.Obj().Name()
				for _, nm := range goparser.IgnoreStructs {
					if nm == typName {
						return false
					}
				}
			}

			if _, ok := exp.(*types.Struct); ok {
				return true
			}

		case "struct":
			if named, ok := typ.(*types.Named); ok {
				if _, ok := named.Underlying().(*types.Struct); ok {
					typName := named.Obj().Pkg().Name() + "." + named.Obj().Name()
					for _, nm := range goparser.IgnoreStructs {
						if nm == typName {
							return false
						}
					}
					return true
				}
			}
		case "slice":
			if _, ok := typ.(*types.Slice); ok {
				return true
			}
		case "numeric":
			if basic, ok := typ.(*types.Basic); ok {
				if (basic.Info() & types.IsNumeric) != 0 {
					return true
				}
			} else {
				typ = typ.Underlying()
				if basic, ok := typ.(*types.Basic); ok {
					if (basic.Info() & types.IsNumeric) != 0 {
						return true
					}
				}
			}
		case "bool", "boolean":
			if basic, ok := typ.(*types.Basic); ok {
				if (basic.Info() & types.IsBoolean) != 0 {
					return true
				}
			} else {
				typ = typ.Underlying()
				if basic, ok := typ.(*types.Basic); ok {
					if (basic.Info() & types.IsBoolean) != 0 {
						return true
					}
				}
			}
		case "string":
			if basic, ok := typ.(*types.Basic); ok {
				if basic.Kind() == types.String {
					return true
				}
			} else {
				typ = typ.Underlying()
				if basic, ok := typ.(*types.Basic); ok {
					if basic.Kind() == types.String {
						return true
					}
				}
			}

		case "basic":
			if _, ok := typ.(*types.Basic); ok {
				return true
			}
			typ = typ.Underlying()
			if _, ok := typ.(*types.Basic); ok {
				return true
			}

		case "interface", "interface{}":
			if _, ok := typ.(*types.Interface); ok {
				return true
			}
			typ = typ.Underlying()
			if _, ok := typ.(*types.Interface); ok {
				return true
			}
		default:
			panic(errors.New("unknown type - " + name + "," + strings.Join(or, ",")))
		}
	}

	return false
}
