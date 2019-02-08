package gowasm

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/gowebapi/webidlgenerator/types"
)

func convertType(value types.TypeRef, data interface{}, tmpl *template.Template) string {
	t := findTypeTemplate(value, tmpl)
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		panic(err)
	}
	out := buf.String()
	// out = strings.Replace(out, "\n", " ", -1)
	out = strings.TrimSpace(out)
	if strings.Contains(out, "\n") {
		out += "\n"
	}
	return out
}

func findTypeTemplate(value types.TypeRef, tmpl *template.Template) *template.Template {
	// find based on type name
	debug := fmt.Sprintf("unable to find in '%s' template: %T", tmpl.Name(), value)
	tmplName := value.Basic().Template
	tmplName = "type-" + tmplName
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
