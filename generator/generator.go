package generator

import (
	"errors"
	"flag"
	"fmt"
	"go/types"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/goparser"
)

type Generator struct {
}

func (cmd *Generator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return fs
}

func (cmd *Generator) Run(args []string) error {
	for _, file := range flag.Args() {
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

	file, err := goparser.Parse(pa)
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
		os.Remove(targetFile + ".tmp")
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
	err = os.Rename(targetFile+".tmp", targetFile)
	if err != nil {
		return err
	}

	// 不知为什么，有时运行两次 goimports 才起效
	exec.Command("goimports", "-w", targetFile).Run()
	return goImports(targetFile)
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
	for _, pa := range file.Imports {
		if pa == `github.com/runner-mei/GoBatis` {
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
			return rv.Index(rv.Len() - 1).Interface()
		}
		if rv.Kind() == reflect.Slice {
			return rv.Index(rv.Len() - 1).Interface()
		}
		return nil
	},
}

var newFunc, implFunc *template.Template

func init() {
	for k, v := range gobatis.TemplateFuncs {
		funcs[k] = v
	}

	newFunc = template.Must(template.New("NewFunc").Funcs(funcs).Parse(`
{{- define "insert"}}
	{{-   $var_undefined := default .var_undefined "false"}}
	{{-   if eq $var_undefined "true"}}
	sqlStr
	{{- else}}
	s
	{{- end}}, err := gobatis.GenerateInsertSQL(ctx.Dialect, ctx.Mapper, 
	reflect.TypeOf(&{{typePrint .printContext .recordType}}{}), 
	{{- if eq (len .method.Results.List) 2 -}}
    	false
    	{{- else -}}
    	true
    	{{- end}})
	if err != nil {
		return gobatis.ErrForGenerateStmt(err, "generate {{.itf.Name}}.{{.method.Name}} error")
	}
	{{- if ne $var_undefined "true"}}
	sqlStr = s
	{{- end}}
{{- end}}

{{- define "update"}}

	{{- set . "var_first_is_context" false}}
	{{- set . "var_contains_struct" false}}
	{{- $lastParam := last .method.Params.List}}
	{{- $var_undefined := default .var_undefined "false"}}
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

		{{-   if eq $var_undefined "true"}}
		sqlStr
		{{- else}}
		s
		{{- end}}, err := 

		{{- if $var_style_1 -}}
				gobatis.GenerateUpdateSQL(ctx.Dialect, ctx.Mapper, 
					"{{$lastParam.Name}}.", reflect.TypeOf(&{{typePrint .printContext .recordType}}{}), []string{
					{{- range $idx, $param := .method.Params.List}}
						{{- if isType $param.Type "context" | not -}}
							{{- if lt $idx ( sub (len $.method.Params.List) 1) }}
				        	"{{$param.Name}}",
							{{- end}}
					 	{{- end}}
					{{- end}}
				})
		{{-  else -}}
				gobatis.GenerateUpdateSQL2(ctx.Dialect, ctx.Mapper, 
					reflect.TypeOf(&{{typePrint .printContext .recordType}}{}), 
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
			        		"{{$param.Name}}",
							{{- end -}}
						{{- end}}
					{{- end}}
				})
		{{-  end}}
		if err != nil {
			return gobatis.ErrForGenerateStmt(err, "generate {{.itf.Name}}.{{.method.Name}} error")
		}
		{{- if ne $var_undefined "true"}}
		sqlStr = s
		{{- end}}
	{{- end}}
{{- end}}

{{- define "delete"}}
	{{-   $var_undefined := default .var_undefined "false"}}
	{{-   if eq $var_undefined "true"}}
	sqlStr
	{{- else}}
	s
	{{- end}}, err := gobatis.GenerateDeleteSQL(ctx.Dialect, ctx.Mapper, 
	reflect.TypeOf(&{{typePrint .printContext .recordType}}{}), []string{
	{{-     range $idx, $param := .method.Params.List}}
	{{-       if isType $param.Type "context" | not}}
		"{{$param.Name}}",
	{{-       end}}
	{{-     end}}
		})
	if err != nil {
		return gobatis.ErrForGenerateStmt(err, "generate {{.itf.Name}}.{{.method.Name}} error")
	}
	{{- if ne $var_undefined "true"}}
	sqlStr = s
	{{- end}}
{{- end}}

{{- define "count"}}
	{{-   $var_undefined := default .var_undefined "false"}}
	{{-   if eq $var_undefined "true"}}
	sqlStr
	{{- else}}
	s
	{{- end}}, err := gobatis.GenerateCountSQL(ctx.Dialect, ctx.Mapper, 
	reflect.TypeOf(&{{typePrint .printContext .recordType}}{}), []string{
	{{-     range $idx, $param := .method.Params.List}}
	{{-       if isType $param.Type "context" | not }}
		"{{$param.Name}}",
	{{-       end}}
	{{-     end}}
		})
	if err != nil {
		return gobatis.ErrForGenerateStmt(err, "generate {{.itf.Name}}.{{.method.Name}} error")
	}
	{{- if ne $var_undefined "true"}}
	sqlStr = s
	{{- end}}
{{- end}}


{{- define "select"}}
	{{-   $var_undefined := default .var_undefined "false"}}
	{{-   if eq $var_undefined "true"}}
	sqlStr
	{{- else}}
	s
	{{- end}}, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper, 
	reflect.TypeOf(&{{typePrint .printContext .recordType}}{}), []string{
	{{-     range $idx, $param := .method.Params.List}}
	{{-       if isType $param.Type "context" | not }}
		"{{$param.Name}}",
	{{-       end}}
	{{-     end}}
		})
	if err != nil {
		return gobatis.ErrForGenerateStmt(err, "generate {{.itf.Name}}.{{.method.Name}} error")
	}
	{{- if ne $var_undefined "true"}}
	sqlStr = s
	{{- end}}
{{- end}}

{{- define "genSQL"}}

{{-   $var_undefined := default .var_undefined "false"}}
  {{- $recordType := detectRecordType .itf .method}}
  {{- if $recordType}}
    {{- $statementType := .method.StatementTypeName}}
	  {{- if eq $statementType "insert"}}
	  {{-   template "insert" . | arg "recordType" $recordType}}
	  {{-   if eq $var_undefined "true"}}
	  {{-     template "registerStmt" $ }}
	  {{-   end}}
	  {{- else if eq $statementType "update"}}
	  {{-   template "update" . | arg "recordType" $recordType}}
	  {{-   if eq $var_undefined "true"}}
	  {{-     template "registerStmt" $ }}
	  {{-   end}}
	  {{- else if eq $statementType "delete"}}
    {{-   template "delete" . | arg "recordType" $recordType}}
	  {{-   if eq $var_undefined "true"}}
	  {{-     template "registerStmt" $ }}
	  {{-   end}}
	  {{- else if eq $statementType "select"}}
	  {{-   if containSubstr .method.Name "Count" }}
	  {{-     template "count" . | arg "recordType" $recordType}}
	  {{-     if eq $var_undefined "true"}}
	  {{-       template "registerStmt" $ }}
	  {{-     end}}
	  {{-   else}}
	  {{-     template "select" . | arg "recordType" $recordType}}
	  {{-     if eq $var_undefined "true"}}
	  {{-       template "registerStmt" $ }}
	  {{-     end}}
	  {{-   end}}
	  {{- else}}
	  return errors.New("sql '{{.itf.Name}}.{{.method.Name}}' error : statement not found ")
	  {{- end}}
  {{- else}}
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
		if _, exists := ctx.Statements["{{$.itf.Name}}.{{$m.Name}}"]; !exists {
		{{-   if or $m.Config.DefaultSQL  $m.Config.Dialects}}
			sqlStr := {{printf "%q" $m.Config.DefaultSQL}}
			{{-     if $m.Config.Dialects}}
			switch ctx.Dialect {
				{{-    range $typ, $dialect := $m.Config.Dialects}}
			case gobatis.ToDbType("{{$typ}}"):
				sqlStr = {{printf "%q" $dialect}}
				{{-    end}}
			}

			{{-     end}}
			{{- if not $m.Config.DefaultSQL}}
			if sqlStr == "" {	
			   {{- template "genSQL" $ | arg "method" $m }}
			}
			{{- end}}
			{{- template "registerStmt" $ | arg "method" $m}}
		{{- else}}
			{{- template "genSQL" $ | arg "method" $m | arg "var_undefined" "true"}}
		{{-   end}}
		}
	}
	{{-   end}}
	{{- end}}
	return nil
	})
}

func New{{.itf.Name}}(ref *gobatis.Reference
{{- range $if := .itf.ReferenceInterfaces -}}
  , {{- goify $if false}} {{$if -}}
{{- end -}}
	) {{.itf.Name}} {
	return &{{.itf.Name}}Impl{session: ref,
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
   	"{{$param.Name}}",
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
	 "{{$param.Name}}",
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
	    "{{$param.Name}}",
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
	var instance = &{{trimPrefix ($r1.Print .printContext) "*"}}{}
		  {{- end}}
    {{- else}}
	var instance {{$r1.Print .printContext}}
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


	{{$errName}} {{if not $rerr.Name -}}:{{- end -}}= impl.session.SelectOne(
	  	{{- template "printContext" . -}}
	  	"{{.itf.Name}}.{{.method.Name}}",
		{{- if .method.Params.List}}
		[]string{
		{{- range $param := .method.Params.List}}
	    {{-   if isType $param.Type "context" | not }}
		  "{{$param.Name}}",
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
	var instances {{$r1.Print .printContext}}
	{{- end}}
    results := impl.session.Select(
	  	{{- template "printContext" . -}}
	  	"{{.itf.Name}}.{{.method.Name}}",
		{{- if .method.Params.List}}
		[]string{
		{{- range $param := .method.Params.List}}
	    {{-   if isType $param.Type "context" | not }}
		  "{{$param.Name}}",
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
	var {{$r1Name}} = {{$r1.Print .printContext}}{}
	{{- else}}
	{{$r1Name}} = {{$r1.Print .printContext}}{}
	{{- end}}

    results := impl.session.Select(
	  	{{- template "printContext" . -}}
	  	"{{.itf.Name}}.{{.method.Name}}",
		{{- if .method.Params.List}}
		[]string{
		{{- range $param := .method.Params.List}}
	    {{-   if isType $param.Type "context" | not }}
		   "{{$param.Name}}",
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
				instance.Set("{{$r.Name}}", {{$r.Name}})
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
		  "{{$param.Name}}",
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
		    instance.Set("{{$r.Name}}", func(idx int) interface{} {
		    	{{- if isType $r.Type.Elem.Elem "basic"}}
					var newInstance {{typePrint $.printContext $r.Type.Elem.Elem}}
					{{$r.Name}} = append({{$r.Name}}, &newInstance)
					return &newInstance
		    	{{- else}}
					newInstance := &{{typePrint $.printContext $r.Type.Elem.Elem}}{}
					{{$r.Name}} = append({{$r.Name}}, newInstance)
					return newInstance
					{{- end}}
				})
		  {{-   else}}
		    instance.Set("{{$r.Name}}", func(idx int) interface{} {
		    	{{- if isType $r.Type.Elem "string"}}
					{{$r.Name}} = append({{$r.Name}}, "")
		    	{{- else if isType $r.Type.Elem "basic"}}
					{{$r.Name}} = append({{$r.Name}}, 0)
		    	{{- else}}
					{{$r.Name}} = append({{$r.Name}}, {{typePrint $.printContext $r.Type.Elem}}{})
					{{- end}}
					return &{{$r.Name}}[len({{$r.Name}})-1]
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
		 "{{$param.Name}}",
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
    {{- if eq (len .method.Results.List) 2}}
	    {{- $r1 := index .method.Results.List 0}}
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
{{- range $if := .itf.ReferenceInterfaces}}
  {{goify $if false}} {{$if}}
{{- end}}
	session *gobatis.Reference
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
		default:
			panic(errors.New("unknown type - " + name + "," + strings.Join(or, ",")))
		}
	}

	return false
}
