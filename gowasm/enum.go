package gowasm

// TODO what to return if the value doesn't exist?

import (
	"io"
	"text/template"

	"github.com/gowebapi/webidl-bind/types"
)

const enumTmplInput = `
{{define "header"}}
// enum: {{.Basic.Idl}}
type {{.Basic.Def}} int

const (
{{range $idx, $v := .Enum.Values}}	{{$.Enum.Prefix}}{{$v.Def}}{{$.Enum.Suffix}}{{if eq $idx 0}} {{$.Basic.Def}} = iota{{end}}
{{end}}
)

var {{.Basic.Internal}}ToWasmTable = []string{
	{{range .Enum.Values}}"{{.Idl}}", {{end}}
}

var {{.Basic.Internal}}FromWasmTable = map[string]{{.Basic.Def}} {
	{{range .Enum.Values}}"{{.Idl}}": {{$.Enum.Prefix}}{{.Def}}{{$.Enum.Suffix}},{{end}}
}

// JSValue is converting this enum into a java object
func (this * {{.Basic.Def}} ) JSValue() js.Value {
	return js.ValueOf( this.Value() )
}

// Value is converting this into javascript defined
// string value
func (this {{.Basic.Def}} ) Value() string {
	idx := int(this)
	if idx >= 0 && idx < len({{.Basic.Internal}}ToWasmTable) {
		return {{.Basic.Internal}}ToWasmTable[idx]
	}
	panic("unknown input value")
}

// {{.Basic.Def}}FromJS is converting a javascript value into
// a {{.Basic.Def}} enum value.
func {{.Basic.Def}}FromJS(value js.Value) {{.DefaultParam.Output}} {
	key := value.String()
	conv, ok := {{.Basic.Internal}}FromWasmTable[key]
	if !ok {
		panic("unable to convert '" + key + "'")
	}
	return conv
}
{{end}}
`

var enumTempl = template.Must(template.New("enum").Parse(enumTmplInput))

func writeEnum(dst io.Writer, e types.Type) error {
	data := struct {
		Basic        types.BasicInfo
		Enum         types.Type
		DefaultParam *types.TypeInfo
	}{
		Enum: e,
	}
	data.DefaultParam, _ = e.DefaultParam()
	data.Basic = data.DefaultParam.BasicInfo
	return enumTempl.ExecuteTemplate(dst, "header", data)
}
