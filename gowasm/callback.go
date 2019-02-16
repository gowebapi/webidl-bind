package gowasm

import (
	"io"
	"text/template"

	"github.com/gowebapi/webidlgenerator/types"
)

const callbackTmplInput = `
{{define "start"}}
// callback: {{.Type.Idl}}
type {{.Type.Def}}Func func ({{.ParamLine}}) {{.Return.InOut}}

// {{.Type.Def}} is a javascript function type. 
// 
// Call Release() when done to release resouces
// allocated to this type.
type {{.Type.Def}} js.Func

func {{.Type.Def}}ToJS(callback {{.Type.Def}}Func ) * {{.Type.Def}} {
	if callback == nil {
		return nil
	}
	ret := {{.Type.Def}} (js.FuncOf(func (this js.Value, args []js.Value) interface{} {
{{end}}
	
{{define "middle-1"}}
		{{if not .VoidRet}} _returned := {{end}} callback({{.InOut.AllOut}})
{{end}}
{{define "middle-2"}}
		{{if not .VoidRet}}
			return _converted
		{{else}}
			// returning no return value
			return nil
		{{end}}
	}))
	return &ret
}

func {{.Type.Def}}FromJS(_value js.Value) {{.Type.Output}} {
	return func( {{.ParamLine}} ) ( {{if len .Return.InOut}}_result{{end}} {{.Return.InOut}} ) {
		var (
			_args {{.ArgVar}} 
			_end int 
		)
{{end}}
		
{{define "invoke"}}
		{{if not .VoidRet}}_returned := {{end}} _value.Invoke(_args[0:_end]...)
{{end}}
{{define "end"}}
		{{if not .VoidRet}}_result = _converted{{end}}
		return
	}
}
{{end}}
`

var callbackTempl = template.Must(template.New("callback").Parse(callbackTmplInput))

type callbackData struct {
	Type      *types.TypeInfo
	Return    *types.TypeInfo
	VoidRet   bool
	Params    []string
	ParamLine string
	InOut     *inoutData
	ArgVar    string
}

func writeCallback(dst io.Writer, value types.Type) error {
	cb := value.(*types.Callback)
	data := &callbackData{
		InOut:   setupInOutWasmData(cb.Parameters, "args[%d@variadicSlice@]", "_p%d"),
		VoidRet: types.IsVoid(cb.Return),
	}
	data.ArgVar = calculateMethodArgsSize(data.InOut)
	data.Return, _ = cb.Return.DefaultParam()
	data.Type, _ = cb.DefaultParam()
	data.ParamLine, data.Params = parameterArgumentLine(cb.Parameters)
	if err := callbackTempl.ExecuteTemplate(dst, "start", data); err != nil {
		return err
	}
	if err := writeInOutFromWasm(data.InOut, "", dst); err != nil {
		return err
	}
	if err := callbackTempl.ExecuteTemplate(dst, "middle-1", data); err != nil {
		return err
	}
	if !data.VoidRet {
		result := setupInOutWasmForType(cb.Return, "", "_returned", "_converted")
		if err := writeInOutToWasm(result, "", dst); err != nil {
			return err
		}
	}
	if err := callbackTempl.ExecuteTemplate(dst, "middle-2", data); err != nil {
		return err
	}
	fromjs := setupInOutWasmData(cb.Parameters, "@name@", "_p%d")
	assign := "_args[%d] = _p%d; _end++"
	if err := writeInOutToWasm(fromjs, assign, dst); err != nil {
		return err
	}
	if err := callbackTempl.ExecuteTemplate(dst, "invoke", data); err != nil {
		return err
	}
	if !data.VoidRet {
		result := setupInOutWasmForType(cb.Return, "", "_returned", "_converted")
		if err := writeInOutFromWasm(result, "", dst); err != nil {
			return err
		}
	}
	if err := callbackTempl.ExecuteTemplate(dst, "end", data); err != nil {
		return err
	}
	return nil
}
