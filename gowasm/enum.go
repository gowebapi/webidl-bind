package gowasm

import (
	"io"
	"text/template"
	"wasm/generator/types"
)

const enumTmplInput = `
{{define "header"}}
type {{.Name.Public}} int

const (
{{range $idx, $v := .Values}}	{{$v.Go}}{{if eq $idx 0}} {{$.Name.Public}} = iota{{end}}
{{end}}
)

var {{.Name.Local}}ToWasmTable = []string{
	{{range .Values}}"{{.Idl}}", {{end}}
}

func {{.Name.Local}}ToWasm(in {{.Name.Public}}) string {
	idx := int(in)
	if idx >= 0 && idx < len({{.Name.Local}}ToWasmTable) {
		return {{.Name.Local}}ToWasmTable[idx]
	}
	panic("unknown input value")
}
{{end}}
`

var enumTempl = template.Must(template.New("enum").Parse(enumTmplInput))

func writeEnum(dst io.Writer, e types.Type) error {
	return enumTempl.ExecuteTemplate(dst, "header", e)
}
