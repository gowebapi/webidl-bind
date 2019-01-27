package gowasm

// TODO what to return if the value doesn't exist?

import (
	"io"
	"text/template"
	"wasm/generator/types"
)

const enumTmplInput = `
{{define "header"}}
type {{.Name.Def}} int

const (
{{range $idx, $v := .Values}}	{{$v.Go}}{{if eq $idx 0}} {{$.Name.Def}} = iota{{end}}
{{end}}
)

var {{.Name.Internal}}ToWasmTable = []string{
	{{range .Values}}"{{.Idl}}", {{end}}
}

var {{.Name.Internal}}FromWasmTable = map[string]{{.Name.Def}} {
	{{range .Values}}"{{.Idl}}": {{.Go}},{{end}}
}

func {{.Name.Internal}}ToWasm(in {{.Name.InOut}}) string {
	idx := int(in)
	if idx >= 0 && idx < len({{.Name.Internal}}ToWasmTable) {
		return {{.Name.Internal}}ToWasmTable[idx]
	}
	panic("unknown input value")
}

func {{.Name.Internal}}FromWasm(value js.Value) {{.Name.InOut}} {
	key := value.String()
	conv, ok := {{.Name.Internal}}FromWasmTable[key]
	if !ok {
		panic("unable to convert '" + key + "'")
	}
	return conv
}
{{end}}
`

var enumTempl = template.Must(template.New("enum").Parse(enumTmplInput))

func writeEnum(dst io.Writer, e types.Type) error {
	return enumTempl.ExecuteTemplate(dst, "header", e)
}
