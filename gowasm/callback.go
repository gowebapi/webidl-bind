package gowasm

import (
	"io"
	"text/template"
	"wasm/generator/types"
)

const callbackTmplInput = `
{{define "start"}}
type {{.CB.Name.Public}} func ({{.ParamLine}})

func {{.CB.Name.Local}}FromWasm(callback {{.CB.Name.Public}}, args []js.Value) {
	if len(args) != 1 {
		panic("unexpected parameter count")
	}
{{end}}
	
{{define "end"}}
	callback({{.InOut.AllOut}})
}
{{end}}
`

var callbackTempl = template.Must(template.New("callback").Parse(callbackTmplInput))

type callbackData struct {
	CB        *types.Callback
	Return    string
	Params    []string
	ParamLine string
	InOut     *inoutData
}

func writeCallback(dst io.Writer, value types.Type) error {
	cb := value.(*types.Callback)
	data := &callbackData{
		CB:     cb,
		Return: typeDefine(cb.Return),
		InOut:  setupInOutWasmData(cb.Parameters, "args[%d]", "_p%d"),
	}
	data.ParamLine, data.Params = parameterArgumentLine(cb.Parameters)
	if err := callbackTempl.ExecuteTemplate(dst, "start", data); err != nil {
		return err
	}
	if err := writeInOutFromWasm(data.InOut, dst); err != nil {
		return err
	}
	if err := callbackTempl.ExecuteTemplate(dst, "end", data); err != nil {
		return err
	}
	return nil
}
