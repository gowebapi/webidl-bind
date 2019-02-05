package gowasm

import (
	"io"
	"text/template"
	"wasm/generator/types"
)

const callbackTmplInput = `
{{define "start"}}
// callback: {{.Type.Idl}}
type {{.Type.Def}} func ({{.ParamLine}})

func {{.Type.Internal}}FromWasm(callback {{.Type.InOut}}, args []js.Value) {
{{end}}
	
{{define "end"}}
	callback({{.InOut.AllOut}})
}
{{end}}
`

var callbackTempl = template.Must(template.New("callback").Parse(callbackTmplInput))

type callbackData struct {
	CB        *types.Callback
	Type      *types.TypeInfo
	Return    types.BasicInfo
	Params    []string
	ParamLine string
	InOut     *inoutData
}

func writeCallback(dst io.Writer, value types.Type) error {
	cb := value.(*types.Callback)
	data := &callbackData{
		CB:     cb,
		Return: cb.Return.Basic(),
		InOut:  setupInOutWasmData(cb.Parameters, "args[%d]", "_p%d"),
	}
	data.Type, _ = cb.DefaultParam()
	data.ParamLine, data.Params = parameterArgumentLine(cb.Parameters)
	if err := callbackTempl.ExecuteTemplate(dst, "start", data); err != nil {
		return err
	}
	if err := writeInOutFromWasm(data.InOut, "", dst); err != nil {
		return err
	}
	if err := callbackTempl.ExecuteTemplate(dst, "end", data); err != nil {
		return err
	}
	return nil
}
