package gowasm

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"wasm/generator/types"
)

const typeDefineInput = `
{{define "PrimitiveType"}}
	{{.Lang}}
{{end}}

{{define "TypeNameRef"}}
	{{.Name.Public}}
{{end}}

{{define "VoidType"}}
{{end}}

`

const typeFromWasmInput = `
{{define "PrimitveType"}}
	_value . what?
{{end}}

{{define "type-callback"}}	{{.Name.Local}}FromWasm(_cb_value, _cb_args) {{end}}
{{define "type-enum"}}		{{.Name.Local}}FromWasm(_value) {{end}}

{{define "VoidType"}}
     aVoidReturnTypeShallNotBeConvertedButNeedSomeSpecialHandling()
{{end}}
`

const typeToWasmInput = `
{{define "PrimitveType"}}
	_value . what?
{{end}}

{{define "type-callback"}}	unableToConvertCallback_{{.Name.Local}}ToWasm(_cb_value, _cb_args) {{end}}
{{define "type-enum"}}      {{.Name.Local}}ToWasm(_value) {{end}}

{{define "VoidType"}}
     aVoidReturnTypeShallNotBeConvertedButNeedSomeSpecialHandling()
{{end}}
`

const typeTemplateNameInput = `
{{define "PrimitveType"}}
	primitive
{{end}}

{{define "TypeNameRef"}}
	typenameref
{{end}}

{{define "VoidType"}}
     void
{{end}}
`

var typeDefineTmpl = template.Must(template.New("type-define").Parse(typeDefineInput))
var typeTemplateNameTmpl = template.Must(template.New("type-template-name").Parse(typeTemplateNameInput))

func typeDefine(value types.TypeRef) string {
	return convertType(value, value, typeDefineTmpl)
}

func typeTemplateName(value types.TypeRef) string {
	if ref, ok := value.(*types.TypeNameRef); ok {
		switch ref.Underlying.(type) {
		case *types.Callback:
			return "callback"
		case *types.Enum:
			return "enum"
		default:
			panic(fmt.Sprintf("unable to handle %T", ref.Underlying))
		}
	}
	return convertType(value, value, typeTemplateNameTmpl)
}

func convertType(value types.TypeRef, data interface{}, tmpl *template.Template) string {
	info := reflect.TypeOf(value)
	if info.Kind() == reflect.Ptr {
		info = info.Elem()
	}
	name := info.Name()
	t := tmpl.Lookup(name)
	var tmplName string
	if t == nil {
		// unable to find for tag, trying extract other way
		tmplName = "type-" + typeTemplateName(value)
		t = tmpl.Lookup(tmplName)
	}
	if t == nil {
		panic(fmt.Sprintf("unable to find type template '%s' : %T : %s : %s", name, value, tmpl.Name(), tmplName))
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		panic(err)
	}
	out := buf.String()
	// out = strings.Replace(out, "\n", " ", -1)
	out = strings.TrimSpace(out)
	if strings.Index(out, "\n") != -1 {
		out += "\n"
	}
	return out
}
