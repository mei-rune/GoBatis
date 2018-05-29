package generator

import (
	"errors"
	"flag"
	"go/types"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

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

	exec.Command("goimports", "-w", targetFile).Run()
	return nil
}

func (cmd *Generator) generateHeader(out io.Writer, file *goparser.File) error {
	io.WriteString(out, "// Please don't edit this file!\r\npackage ")
	io.WriteString(out, file.Package)
	io.WriteString(out, "\r\n\r\nimport (")
	io.WriteString(out, "\r\n\t\"errors\"")
	for _, pa := range file.Imports {
		io.WriteString(out, "\r\n\t\"")
		io.WriteString(out, pa)
		io.WriteString(out, "\"")
	}
	io.WriteString(out, "\r\n\tgobatis\"github.com/runner-mei/GoBatis\"")
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

var newFunc = template.Must(template.New("NewFunc").Funcs(funcs).Parse(`

func init() {
	gobatis.Init(func(ctx *gobatis.InitContext) error {
	{{- range $m := .itf.Methods}}
	{ //// {{$.itf.Name}}.{{$m.Name}}
		{{-   if or $m.Config.DefaultSQL  $m.Config.Dialects}} 
		var sqlStr = {{printf "%q" $m.Config.DefaultSQL}}
		{{-     if $m.Config.Dialects}}
		switch ctx.DbType {
			{{-    range $typ, $dialect := $m.Config.Dialects}}
		case gobatis.ToDbType("{{$typ}}"):
			sqlStr = {{printf "%q" $dialect}}
			{{-    end}}
		}
		{{-     end}}
		if sqlStr != "" {
			if _, exists := ctx.Statements["{{$.itf.Name}}.{{$m.Name}}"]; exists {
				return errors.New("statement '{{$.itf.Name}}.{{$m.Name}}' is already exists")
			}
			stmt, err := gobatis.NewMapppedStatement("{{$.itf.Name}}.{{$m.Name}}", 
			  {{toStatementType $m.Name $m.Config.StatementType}}, 
			  gobatis.ResultStruct, 
			  sqlStr)
			if err != nil {
				return err
			}
		  ctx.Statements["{{$.itf.Name}}.{{$m.Name}}"] = stmt
		} else {
			if _, exists := ctx.Statements["{{$.itf.Name}}.{{$m.Name}}"]; !exists {
				return errors.New("statement '{{$.itf.Name}}.{{$m.Name}}' isnot exists")
			}
		}
	{{-   end}}
  }
	{{- end}}
	return nil
	})
}

func New{{.itf.Name}}(ref *gobatis.Reference) {{.itf.Name}} {
	return &{{.itf.Name}}Impl{session: ref}
}`))

var implFunc = template.Must(template.New("ImplFunc").Funcs(funcs).Parse(`
{{- define "insert"}}
	return impl.session.Insert("{{.itf.Name}}.{{.method.Name}}",
		{{- if .method.Params.List}}
		[]string{
  	{{- range $param := .method.Params.List}}
   	"{{$param.Name}}",
		{{- end}}
		},
		{{- else -}}
		nil,
		{{- end -}}
		{{- if .method.Params.List}}
		[]interface{}{
			{{- range $param := .method.Params.List}}
				 {{$param.Name}},
			{{- end}}
		}
		{{- else -}}
		nil
		{{- end -}}
    )
{{- end}}

{{- define "update"}}
return impl.session.Update("{{.itf.Name}}.{{.method.Name}}",
	{{- if .method.Params.List}}
	[]string{
	{{- range $param := .method.Params.List}}
	 "{{$param.Name}}",
	{{- end}}
	},
	{{- else -}}
	nil,
	{{- end -}}
	{{- if .method.Params.List}}
	[]interface{}{
		{{- range $param := .method.Params.List}}
			 {{$param.Name}},
		{{- end}}
	}
	{{- else -}}
	nil
	{{- end -}}
	)
{{- end}}

{{- define "delete"}}
return impl.session.Delete("{{.itf.Name}}.{{.method.Name}}",
	{{- if .method.Params.List}}
	[]string{
	{{- range $param := .method.Params.List}}
	 "{{$param.Name}}",
	{{- end}}
	},
	{{- else -}}
	nil,
	{{- end -}}
	{{- if .method.Params.List}}
	[]interface{}{
		{{- range $param := .method.Params.List}}
			 {{$param.Name}},
		{{- end}}
	}
	{{- else -}}
	nil
	{{- end -}}
  )
{{- end}}

{{- define "selectOne"}}
	{{- $r1 := index .method.Results.List 0}}
	{{- $rerr := index .method.Results.List 1}}
	  
	{{- $r1Name := default $r1.Name "instance"}}
	{{- $errName := default $rerr.Name "err"}}

  	{{- if not $r1.Name }}
	{{- if startWith $r1.Type.String "*"}}
  	var instance = &{{trimPrefix ($r1.Print .printContext) "*"}}{}
    {{- else}}
  	var instance {{$r1.Print .printContext}}
    {{- end}}
    {{- end}}

	{{$errName}} {{if not $rerr.Name -}}:{{- end -}}= impl.session.SelectOne("{{.itf.Name}}.{{.method.Name}}",
		{{- if .method.Params.List}}
		[]string{
		{{- range $param := .method.Params.List}}
		 "{{$param.Name}}",
		{{- end}}
		},
		{{- else -}}
		nil,
		{{- end -}}
		{{- if .method.Params.List}}
		[]interface{}{
			{{- range $param := .method.Params.List}}
				 {{$param.Name}},
			{{- end}}
		}
		{{- else -}}
		nil
		{{- end -}}
	{{- if startWith $r1.Type.String "*" -}}
		).Scan({{$r1Name}})
	{{- else -}}
		).Scan(&{{$r1Name}})
	{{- end}}
  if {{$errName}} != nil {
	  {{- if startWith $r1.Type.String "*"}}
    return nil, {{$errName}}
  	{{- else if isType $r1.Type "numeric"}}
    return 0, {{$errName}}
  	{{- else if isType $r1.Type "string"}}
    return "", {{$errName}}
  	{{- else}}
    return nil, {{$errName}}
  	{{- end}}
  }
  return {{$r1Name}}, nil
{{- end}}

{{- define "selectArray"}}
	{{- $r1 := index .method.Results.List 0}}
	{{- $rerr := index .method.Results.List 1}}

	{{- $r1Name := default $r1.Name "instances"}}
	{{- $errName := default $rerr.Name "err"}}

  	{{- if not $r1.Name }}
	var instances {{$r1.Print .printContext}}
	{{- end}}
    results := impl.session.Select("{{.itf.Name}}.{{.method.Name}}",
		{{- if .method.Params.List}}
		[]string{
		{{- range $param := .method.Params.List}}
		 "{{$param.Name}}",
		{{- end}}
		},
		{{- else -}}
		nil,
		{{- end -}}
		{{- if .method.Params.List}}
		[]interface{}{
			{{- range $param := .method.Params.List}}
				 {{$param.Name}},
			{{- end}}
		}
		{{- else -}}
		nil
		{{- end -}}
		)
  {{$errName}} {{if not $rerr.Name -}}:{{- end -}}= results.ScanSlice(&{{$r1Name}})
  if {{$errName}} != nil {
    return nil, {{$errName}}
  }
  return {{$r1Name}}, nil
{{- end}}

{{- define "select"}}
  {{- if .method.Results}}
  {{- if eq (len .method.Results.List) 2}}
	  {{- $r1 := index .method.Results.List 0}}
	  {{- if containSubstr $r1.Type.String "["}}
	  {{- template "selectArray" $}}
	  {{- else}}
	  {{- template "selectOne" $}}
	  {{- end}}
  {{- else}}
  results is unsupported
  {{- end}}
  {{- else}}
  results is empty?
  {{- end}}
{{- end}}

type {{.itf.Name}}Impl struct {
	session *gobatis.Reference
}
{{ range $m := .itf.Methods}}
func (impl *{{$.itf.Name}}Impl) {{$m.MethodSignature $.printContext}} {
	{{- $statementType := toStatementTypeName $m.Name $m.Config.StatementType}}
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
}
{{end}}
`))

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
	"default": func(value, defvalue interface{}) interface{} {
		if nil == value {
			return defvalue
		}
		if s, ok := value.(string); ok && "" == s {
			return defvalue
		}
		return value
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
	"toStatementType": func(name, typ string) string {
		if typ != "" {
			switch strings.ToLower(typ) {
			case "insert":
				return "gobatis.StatementTypeInsert"
			case "update":
				return "gobatis.StatementTypeUpdate"
			case "delete":
				return "gobatis.StatementTypeDelete"
			case "select":
				return "gobatis.StatementTypeSelect"
			}
			return "gobatis.StatementType" + typ
		}
		if isInsertStatement(name) {
			return "gobatis.StatementTypeInsert"
		}
		if isUpdateStatement(name) {
			return "gobatis.StatementTypeUpdate"
		}
		if isDeleteStatement(name) {
			return "gobatis.StatementTypeDelete"
		}
		if isSelectStatement(name) {
			return "gobatis.StatementTypeSelect"
		}
		return "gobatis.StatementTypeUnknown_" + name
	},
	"toStatementTypeName": func(name, typ string) string {
		if typ != "" {
			switch strings.ToLower(typ) {
			case "insert":
				return "insert"
			case "update":
				return "update"
			case "delete":
				return "delete"
			case "select":
				return "select"
			}
			return typ
		}
		if isInsertStatement(name) {
			return "insert"
		}
		if isUpdateStatement(name) {
			return "update"
		}
		if isDeleteStatement(name) {
			return "delete"
		}
		if isSelectStatement(name) {
			return "select"
		}
		return "unknown_" + name
	},
}

func isExceptedType(typ types.Type, name string) bool {
	switch name {
	case "numeric":
		if basic, ok := typ.(*types.Basic); ok {
			return (basic.Info() & types.IsNumeric) != 0
		}
		typ = typ.Underlying()
		if basic, ok := typ.(*types.Basic); ok {
			return (basic.Info() & types.IsNumeric) != 0
		}
		return false
	case "string":
		if basic, ok := typ.(*types.Basic); ok {
			return basic.Kind() == types.String
		}
		typ = typ.Underlying()
		if basic, ok := typ.(*types.Basic); ok {
			return basic.Kind() == types.String
		}
		return false
	default:
		panic(errors.New("unknown type - " + name))
	}
}
