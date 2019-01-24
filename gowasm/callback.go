package gowasm

import (
	"io"
	"strings"
	"text/template"
	"wasm/generator/types"
)

const callbackTmplInput = `
{{define "header"}}
type {{.CB.Name.Public}} func ({{.ParamLine}})

{{end}}
`

var callbackTempl = template.Must(template.New("callback").Parse(callbackTmplInput))

type callbackData struct {
	CB        *types.Callback
	Return    string
	Params    []string
	ParamLine string
}

func writeCallback(dst io.Writer, value types.Type) error {
	cb := value.(*types.Callback)
	data := &callbackData{
		CB:     cb,
		Return: convertType(cb.Return),
	}
	for _, pi := range cb.Parameters {
		po := convertParameter(pi)
		data.Params = append(data.Params, po)
	}
	data.ParamLine = strings.Join(data.Params, ", ")

	return callbackTempl.ExecuteTemplate(dst, "header", data)
}
