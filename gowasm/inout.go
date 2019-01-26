package gowasm

import (
	"fmt"
	"io"
	"strings"
	"text/template"
	"wasm/generator/types"
)

const inoutTmplInput = `
{{define "to-wasm-start"}}
{{if .ReleaseHdl}}	var _releaseList releasableApiResourceList {{end}}
{{end}}
{{define "to-wasm-end"}}
{{end}}

{{define "to-wasm-type-callback"}}
	{{.Out}} := js.NewCallback(func (_cb_args []js.Value) {
		_cb_value := {{.Name}}
		{{.From}}
	})
	_releaseList = append(_releaseList, {{.Out}})
{{end}}

{{define "from-wasm-start"}}
{{end}}
{{define "from-wasm-end"}}
{{end}}

{{define "from-wasm-type-enum"}}	{{.Name.Local}}FromWasm(_value) {{end}}

`

var inoutTmpl = template.Must(template.New("inout").Parse(inoutTmplInput))

type inoutToWasm struct {
	Params    string
	ParamList []inoutParam
	AllOut    string

	// ReleaseHdl indicate that some input parameter require a returning
	// release handle
	ReleaseHdl bool
}

type inoutFromWasm struct {
}

type inoutParam struct {
	Name  string
	Type  string
	From  string
	To    string
	Tmpl  string
	Out   string
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

func setupInOutToWasm(params []*types.Parameter) *inoutToWasm {
	paramTextList := []string{}
	paramList := []inoutParam{}
	allout := []string{}
	releaseHdl := false
	for idx, pi := range params {
		po := inoutParam{
			Name:  pi.Name,
			Type:  typeDefine(pi.Type),
			From:  typeFromWasm(pi.Type),
			To:    typeToWasm(pi.Type),
			Tmpl:  typeTemplateName(pi.Type),
			Out:   fmt.Sprint("_p", idx),
			RealP: pi,
			RealT: pi.Type,
		}
		releaseHdl = releaseHdl || pi.Type.NeedRelease()
		paramList = append(paramList, po)
		paramTextList = append(paramTextList, fmt.Sprint(pi.Name, " ", po.Type))
		allout = append(allout, po.Out)
	}
	return &inoutToWasm{
		ParamList:  paramList,
		Params:     strings.Join(paramTextList, ", "),
		ReleaseHdl: releaseHdl,
		AllOut:     strings.Join(allout, ", "),
	}
}

func writeInOutToWasm(data *inoutToWasm, dst io.Writer) error {
	return writeInOutLoop(data, "to-wasm", dst)
}

func writeInOutFromWasm(data *inoutToWasm, dst io.Writer) error {
	return writeInOutLoop(data, "from-wasm", dst)
}

func writeInOutLoop(data *inoutToWasm, name string, dst io.Writer) error {
	if err := inoutTmpl.ExecuteTemplate(dst, name+"-start", data); err != nil {
		return err
	}
	for _, p := range data.ParamList {
		if err := inoutTmpl.ExecuteTemplate(dst, name+"-type-"+p.Tmpl, p); err != nil {
			return err
		}
	}
	if err := inoutTmpl.ExecuteTemplate(dst, name+"-end", data); err != nil {
		return err
	}
	return nil
}
