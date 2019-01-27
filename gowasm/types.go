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
	{{.Value.Lang}}
{{end}}

{{define "TypeNameRef"}}
	{{if .InOut}}	
		{{.Value.Name.InOut}}
	{{else}}
		{{.Value.Name.Def}}
	{{end}}
{{end}}

{{define "VoidType"}}
{{end}}

{{define "InterfaceType"}}
	{{if .InOut}}	
		{{.Value.If.Name.InOut}}
	{{else}}
		{{.Value.If.Name.Def}}
	{{end}}
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

func typeDefine(value types.TypeRef, inout bool) string {
	data := struct {
		Value types.TypeRef
		InOut bool
	}{
		Value: value,
		InOut: inout,
	}
	return convertType(value, data, typeDefineTmpl)
}

func typeTemplateName(value types.TypeRef) string {
	if ref, ok := value.(*types.TypeNameRef); ok {
		switch ref.Underlying.(type) {
		case *types.Callback:
			return "callback"
		case *types.Enum:
			return "enum"
		case *types.Dictionary:
			return "dictionary"
		default:
			panic(fmt.Sprintf("unable to handle %T: %#v", ref.Underlying, ref))
		}
	}
	switch value.(type) {
	case *types.PrimitiveType:
		// TODO expand primitive type?
		return "PrimitiveType"
	case *types.VoidType:
		return "VoidType"
	case *types.InterfaceType:
		return "InterfaceType"
	}
	panic(fmt.Sprintf("unknown type %T", value))
}

func convertType(value types.TypeRef, data interface{}, tmpl *template.Template) string {
	t := findTypeTemplate(value, tmpl)
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

func findTypeTemplate(value types.TypeRef, tmpl *template.Template) *template.Template {
	// find based on type name
	debug := fmt.Sprintf("unable to find in '%s' template: %T", tmpl.Name(), value)
	tmplName := "type-" + typeTemplateName(value)
	t := tmpl.Lookup(tmplName)
	if t != nil {
		return t
	}
	debug += " : " + tmplName

	// try some more "global" name
	info := reflect.TypeOf(value)
	if info.Kind() == reflect.Ptr {
		info = info.Elem()
	}
	name := info.Name()
	t = tmpl.Lookup(name)
	if t == nil {
		debug += " : " + name
		panic(debug)
	}
	return t
}
