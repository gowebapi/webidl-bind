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

{{define "type-callback"}}
	{{.Out}} := js.NewCallback(func (_cb_args []js.Value) {
		{{.TypeRef.Name.Local}}FromWasm({{.In}}, _cb_args)
	})
	_releaseList = append(_releaseList, {{.Out}})
{{end}}
{{define "type-enum"}}      {{.Type.Name.Local}}ToWasm(_value) {{end}}
`

const inoutFromTmplInput = `
{{define "start"}}
{{end}}
{{define "end"}}
{{end}}

{{define "type-callback"}} callbackInFrom() {{end}}
{{define "type-enum"}}	{{.Out}} := {{.TypeRef.Name.Local}}FromWasm({{.In}}) {{end}}

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
	// Type in text
	Type string
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
		if value.Optional || value.Variadic {
			panic("todo")
		}
		t := typeDefine(value.Type)
		name := value.Name + " " + t
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
			Type:  typeDefine(pi.Type),
			Tmpl:  typeTemplateName(pi.Type),
			RealP: pi,
			RealT: pi.Type,
			In:    setupVarName(in, idx, pi.Name),
			Out:   setupVarName(out, idx, pi.Name),
		}
		releaseHdl = releaseHdl || pi.Type.NeedRelease()
		paramList = append(paramList, po)
		paramTextList = append(paramTextList, fmt.Sprint(pi.Name, " ", po.Type))
		allout = append(allout, po.Out)
	}
	return &inoutData{
		ParamList:  paramList,
		Params:     strings.Join(paramTextList, ", "),
		ReleaseHdl: releaseHdl,
		AllOut:     strings.Join(allout, ", "),
	}
}

func setupVarName(value string, idx int, name string) string {
	if value == "@name" {
		return name
	} else if strings.Index(value, "%") != -1 {
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
		code := inoutGetToFromWasm(p.RealT, p.Out, p.In, tmpl)
		if _, err := io.WriteString(dst, code); err != nil {
			return err
		}
	}
	if err := tmpl.ExecuteTemplate(dst, "end", data); err != nil {
		return err
	}
	return nil
}

func inoutGetToFromWasm(t types.TypeRef, out, in string, tmpl *template.Template) string {
	data := struct {
		In, Out string
		Type    types.TypeRef
		TypeRef *types.TypeNameRef
	}{
		In:   in,
		Out:  out,
		Type: t,
	}
	if ref, ok := t.(*types.TypeNameRef); ok {
		data.TypeRef = ref
	}
	return convertType(t, data, tmpl)
}
