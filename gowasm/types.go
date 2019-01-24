package gowasm

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"wasm/generator/types"
)

const typeTmplInput = `
{{define "PrimitiveType"}}
	{{.Lang}}
{{end}}

{{define "TypeNameRef"}}
	{{.Name.Public}}
{{end}}

{{define "VoidType"}}
{{end}}

`

var typeTmpl = template.Must(template.New("type").Parse(typeTmplInput))

func convertType(value types.TypeRef) string {
	info := reflect.TypeOf(value)
	if info.Kind() == reflect.Ptr {
		info = info.Elem()
	}
	name := info.Name()
	t := typeTmpl.Lookup(name)
	if t == nil {
		panic(fmt.Sprintf("unable to find type template '%s' : %T", name, value))
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, value); err != nil {
		panic(err)
	}
	out := buf.String()
	out = strings.Replace(out, "\n", " ", -1)
	return out
}

func convertParameter(value *types.Parameter) string {
	if value.Optional || value.Variadic {
		panic("todo")
	}
	t := convertType(value.Type)
	return value.Name + " " + t
}
