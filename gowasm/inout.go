package gowasm

import (
	"fmt"
	"io"
	"strings"
	"text/template"
	"wasm/generator/types"
)

const inoutToTmplInput = `
{{define "start"}}
{{if .ReleaseHdl}}	var _releaseList releasableApiResourceList {{end}}
{{end}}
{{define "end"}}
{{end}}

{{define "type-primitive"}}		{{.Out}} := {{.In}} {{end}}
{{define "type-dictionary"}}	{{.Out}} := {{.Info.Internal}}ToWasm( {{.In}} ) {{end}}
{{define "type-interface"}}		{{.Out}} := {{.Info.Internal}}ToWasm( {{.In}} ) {{end}}

{{define "type-callback"}}
	{{.Out}} := js.NewCallback(func (_cb_args []js.Value) {
		{{.Info.Internal}}FromWasm({{.In}}, _cb_args)
	})
	_releaseList = append(_releaseList, {{.Out}})
{{end}}
{{define "type-enum"}}      {{.Out}} := {{.Info.Internal}}ToWasm({{.In}}) {{end}}
`

const inoutFromTmplInput = `
{{define "start"}}
{{end}}
{{define "end"}}
{{end}}

{{define "type-primitive"}}	{{.Out}} := ({{.In}}).{{.Type.JsMethod}}() {{end}}
{{define "type-callback"}}	callbackInFrom() {{end}}
{{define "type-enum"}}		{{.Out}} := {{.Info.Internal}}FromWasm( {{.In}} ) {{end}}
{{define "type-interface-type"}} {{.Out}} := {{.Info.Internal}}FromWasm( {{.In}} ) {{end}}
{{define "type-interface"}}	{{.Out}} := {{.Info.Internal}}FromWasm( {{.In}} ) {{end}}
{{define "type-union"}}  {{.Out}} := {{.Info.Internal}}FromWasm( {{.In}} ) {{end}}
{{define "type-any"}}    {{.Out}} := {{.In}} {{end}}

`

var inoutToTmpl = template.Must(template.New("inout-to").Parse(inoutToTmplInput))
var inoutFromTmpl = template.Must(template.New("inout-from").Parse(inoutFromTmplInput))

type inoutData struct {
	Params    string
	ParamList []inoutParam
	AllOut    string

	// ReleaseHdl indicate that some input parameter require a returning
	// release handle
	ReleaseHdl bool
}

type inoutParam struct {
	// IDl variable name
	Name string
	// Info about the type
	Info *types.TypeInfo
	// template name
	Tmpl string
	// input variable during convert to/from wasm
	In string
	// output variable during convert to/from wasm
	Out string

	// RealP references input parameter
	RealP *types.Parameter
	RealT types.TypeRef
}

func parameterArgumentLine(input []*types.Parameter) (all string, list []string) {
	for _, value := range input {
		info := value.Type.Param(false, value.Optional, value.Variadic)
		name := value.Name + " " + info.InOut
		list = append(list, name)
	}
	all = strings.Join(list, ", ")
	return
}

func setupInOutWasmData(params []*types.Parameter, in, out string) *inoutData {
	paramTextList := []string{}
	paramList := []inoutParam{}
	allout := []string{}
	releaseHdl := false
	for idx, pi := range params {
		po := inoutParam{
			Name:  pi.Name,
			Info:  pi.Type.Param(false, pi.Optional, pi.Variadic),
			RealP: pi,
			RealT: pi.Type,
			In:    setupVarName(in, idx, pi.Name),
			Out:   setupVarName(out, idx, pi.Name),
		}
		po.Tmpl = po.Info.Template
		releaseHdl = releaseHdl || pi.Type.NeedRelease()
		paramList = append(paramList, po)
		paramTextList = append(paramTextList, fmt.Sprint(pi.Name, " ", po.Info.InOut))
		allout = append(allout, po.Out)
	}
	return &inoutData{
		ParamList:  paramList,
		Params:     strings.Join(paramTextList, ", "),
		ReleaseHdl: releaseHdl,
		AllOut:     strings.Join(allout, ", "),
	}
}

func setupInOutWasmForOne(param *types.Parameter, in, out string) *inoutData {
	idx := 0
	pi := param
	po := inoutParam{
		Name:  pi.Name,
		Info:  pi.Type.Param(false, pi.Optional, pi.Variadic),
		RealP: pi,
		RealT: pi.Type,
		In:    setupVarName(in, idx, pi.Name),
		Out:   setupVarName(out, idx, pi.Name),
	}
	po.Tmpl = po.Info.Template
	return &inoutData{
		ParamList:  []inoutParam{po},
		Params:     fmt.Sprint(pi.Name, " ", po.Info.InOut),
		ReleaseHdl: pi.Type.NeedRelease(),
		AllOut:     po.Out,
	}
}
func setupInOutWasmForType(t types.TypeRef, in, out string) *inoutData {
	pi := types.Parameter{
		Name:     "<only-type>",
		Optional: false,
		Variadic: false,
		Type:     t,
	}
	return setupInOutWasmForOne(&pi, in, out)
}

func setupVarName(value string, idx int, name string) string {
	value = strings.Replace(value, "@name@", name, -1)
	if strings.Index(value, "%") != -1 {
		return fmt.Sprintf(value, idx)
	}
	return value
}

func writeInOutToWasm(data *inoutData, dst io.Writer) error {
	return writeInOutLoop(data, inoutToTmpl, dst)
}

func writeInOutFromWasm(data *inoutData, dst io.Writer) error {
	return writeInOutLoop(data, inoutFromTmpl, dst)
}

func writeInOutLoop(data *inoutData, tmpl *template.Template, dst io.Writer) error {
	if err := tmpl.ExecuteTemplate(dst, "start", data); err != nil {
		return err
	}
	for _, p := range data.ParamList {
		code := inoutGetToFromWasm(p.RealT, p.Info, p.Out, p.In, tmpl)
		if _, err := io.WriteString(dst, code); err != nil {
			return err
		}
	}
	if err := tmpl.ExecuteTemplate(dst, "end", data); err != nil {
		return err
	}
	return nil
}

func inoutGetToFromWasm(t types.TypeRef, info *types.TypeInfo, out, in string, tmpl *template.Template) string {
	if info == nil {
		info = t.DefaultParam()
	}
	data := struct {
		In, Out string
		Type    types.TypeRef
		Info    *types.TypeInfo
	}{
		In:   in,
		Out:  out,
		Type: t,
		Info: info,
	}
	return convertType(t, data, tmpl) + "\n"
}
