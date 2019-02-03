package gowasm

// TODO what to return if the value doesn't exist?

import (
	"io"
	"text/template"
	"wasm/generator/types"
)

const enumTmplInput = `
{{define "header"}}
type {{.Basic.Def}} int

const (
{{range $idx, $v := .Enum.Values}}	{{$v.Go}}{{if eq $idx 0}} {{$.Basic.Def}} = iota{{end}}
{{end}}
)

var {{.Basic.Internal}}ToWasmTable = []string{
	{{range .Enum.Values}}"{{.Idl}}", {{end}}
}

var {{.Basic.Internal}}FromWasmTable = map[string]{{.Basic.Def}} {
	{{range .Enum.Values}}"{{.Idl}}": {{.Go}},{{end}}
}

func {{.Basic.Internal}}ToWasm(in {{.DefaultParam.InOut}}) string {
	idx := int(in)
	if idx >= 0 && idx < len({{.Basic.Internal}}ToWasmTable) {
		return {{.Basic.Internal}}ToWasmTable[idx]
	}
	panic("unknown input value")
}

func {{.Basic.Internal}}FromWasm(value js.Value) {{.DefaultParam.InOut}} {
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
