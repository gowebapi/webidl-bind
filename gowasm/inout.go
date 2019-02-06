package gowasm

import (
	"bytes"
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

{{define "param-start"}}
	{{if .Optional}}
		{{if .AnyType}}
			if {{.In}}.Type() != js.TypeUndefined {
		{{else}}
			if {{.In}} != nil {
		{{end}}
	{{end}}
{{end}}
{{define "param-end"}}
	{{.Assign}}
	{{if .Optional}}
		}
	{{end}}
{{end}}

{{define "type-primitive"}}		{{.Out}} := {{.In}} {{end}}
{{define "type-dictionary"}}	{{.Out}} := {{.In}}.JSValue() {{end}}
{{define "type-interface"}}		{{.Out}} := {{.In}}.JSValue() {{end}}

{{define "type-callback"}}
	{{.Out}} := js.NewCallback(func (_cb_args []js.Value) {
		invoke{{.Info.Def}} ( {{.In}}, _cb_args )
	})
	_releaseList = append(_releaseList, {{.Out}})
{{end}}
{{define "type-enum"}}      {{.Out}} := {{.In}}.JSValue() {{end}}
{{define "type-union"}}	{{.Out}} := {{.In}}.JSValue() {{end}}
{{define "type-any"}}    {{.Out}} := {{.In}} {{end}}
{{define "type-typedarray"}} {{.Out}} := {{.In}} {{end}}
{{define "type-parametrized"}}	{{.Out}} := {{.In}}.JSValue() {{end}}

{{define "type-sequence"}} 
	{{.Out}} := js.Global().Get("Array").New(len( {{if .Info.Pointer}}*{{end}} {{.In}} ))
	for __idx, __in := range {{if .Info.Pointer}}*{{end}} {{.In}} {
		{{.Inner}}
		{{.Out}} .SetIndex(__idx, __out )
	}
{{end}}

{{define "type-variadic"}}
	for _, __in := range {{.In}} {
		{{.Inner}}
		_args[_end] = __out
		_end++
	} 
{{end}}

`

const inoutFromTmplInput = `
{{define "start"}}
	var (
	{{range .ParamList}}
		{{.Out}} {{.Info.InOut}} // javascript: {{.Info.Idl}} {{.Name}}
	{{end}}
	)
{{end}}
{{define "end"}}{{end}}

{{define "param-start"}}
	{{if .Optional}}
		if len(args) > {{.Idx}} {
	{{end}}
	{{if .Nullable}}
		if {{.In}}.Type() != js.TypeNull {
	{{end}}
{{end}}
{{define "param-end"}}
	{{if .Optional}}
		}
	{{end}}
	{{if .Nullable}}
		}
	{{end}}
{{end}}

{{define "type-primitive"}}	
	{{if .Info.Pointer}}__tmp := {{else}} {{.Out}} = {{end}} {{if .Type.Cast}}( {{.Type.Lang}} ) ( {{end}} ( {{.In}} ) . {{.Type.JsMethod}} () {{if .Type.Cast}} ) {{end}}
	{{if .Info.Pointer}} {{.Out}} = &__tmp {{end}}
{{end}}
{{define "type-callback"}}	{{.Out}} = {{.Info.Def}}FromJS( {{.In}} ) {{end}}
{{define "type-enum"}}		{{.Out}} = {{.Info.Def}}FromJS( {{.In}} ) {{end}}
{{define "type-interface"}}	{{.Out}} = {{.Info.Def}}FromJS( {{.In}} ) {{end}}
{{define "type-union"}}  {{.Out}} = {{.Info.Def}}FromJS( {{.In}} ) {{end}}
{{define "type-any"}}    {{.Out}} = {{.In}} {{end}}
{{define "type-typedarray"}} {{.Out}} = {{.In}} {{end}}
{{define "type-parametrized"}}	{{.Out}} = {{.Info.Def}}FromJS( {{.In}} ) {{end}}
{{define "type-dictionary"}}	{{.Out}} = {{.Info.Def}}FromJS( {{.In}} ) {{end}}

{{define "type-sequence"}}
	__length{{.Idx}} := {{.In}}.Length() 
	{{.Out}} = make( {{.Info.InOut}} , __length{{.Idx}}, __length{{.Idx}} )
	for __idx := 0; __idx < __length{{.Idx}} ; __idx++ {
		var __out {{.InnerInfo.InOut}}
		__in := {{.In}}.Index(__idx)
		{{.Inner}}
		{{.Out}}[__idx] = __out
	}
{{end}}
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

	// Param references input parameter
	Param *types.Parameter

	// Inner type definintion
	Type types.TypeRef
}

func parameterArgumentLine(input []*types.Parameter) (all string, list []string) {
	for _, value := range input {
		info, _ := value.Type.Param(false, value.Optional, value.Variadic)
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
			Param: pi,
			In:    setupVarName(in, idx, pi.Name),
			Out:   setupVarName(out, idx, pi.Name),
		}
		po.Info, po.Type = pi.Type.Param(false, pi.Optional, pi.Variadic)
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
		Param: pi,
		In:    setupVarName(in, idx, pi.Name),
		Out:   setupVarName(out, idx, pi.Name),
	}
	po.Info, po.Type = pi.Type.Param(false, pi.Optional, pi.Variadic)
	po.Tmpl = po.Info.Template
	return &inoutData{
		ParamList:  []inoutParam{po},
		Params:     fmt.Sprint(pi.Name, " ", po.Info.InOut),
		ReleaseHdl: pi.Type.NeedRelease(),
		AllOut:     po.Out,
	}
}
func setupInOutWasmForType(t types.TypeRef, name, in, out string) *inoutData {
	pi := types.Parameter{
		Name:     name,
		Optional: false,
		Variadic: false,
		Type:     t,
	}
	return setupInOutWasmForOne(&pi, in, out)
}

func setupVarName(value string, idx int, name string) string {
	value = strings.Replace(value, "@name@", name, -1)
	count := strings.Count(value, "%")
	switch count {
	case 0:
	case 1:
		value = fmt.Sprintf(value, idx)
	case 2:
		value = fmt.Sprintf(value, idx, idx)
	default:
		panic("invalid count")
	}
	return value
}

func writeInOutToWasm(data *inoutData, assign string, dst io.Writer) error {
	return writeInOutLoop(data, assign, inoutToTmpl, dst)
}

func writeInOutFromWasm(data *inoutData, assign string, dst io.Writer) error {
	return writeInOutLoop(data, assign, inoutFromTmpl, dst)
}

func writeInOutLoop(data *inoutData, assign string, tmpl *template.Template, dst io.Writer) error {
	if err := tmpl.ExecuteTemplate(dst, "start", data); err != nil {
		return err
	}
	for idx, p := range data.ParamList {
		start := inoutParamStart(p.Type, p.Info, p.Out, p.In, idx, tmpl)
		if _, err := io.WriteString(dst, start); err != nil {
			return err
		}
		code := inoutGetToFromWasm(p.Type, p.Info, p.Out, p.In, idx, tmpl)
		if _, err := io.WriteString(dst, code); err != nil {
			return err
		}
		av := setupVarName(assign, idx, p.Name)
		end := inoutParamEnd(p.Info, av, tmpl)
		if _, err := io.WriteString(dst, end); err != nil {
			return err
		}
	}
	if err := tmpl.ExecuteTemplate(dst, "end", data); err != nil {
		return err
	}
	return nil
}

func inoutGetToFromWasm(t types.TypeRef, info *types.TypeInfo, out, in string, idx int, tmpl *template.Template) string {
	if info == nil {
		panic("null")
		// info = t.DefaultParam()
	}

	// convert current
	data := struct {
		In, Out string
		Type    types.TypeRef
		Info    *types.TypeInfo
		Idx     int
		Inner   string

		InnerInfo *types.TypeInfo
		InnerType types.TypeRef
	}{
		In:   in,
		Type: t,
		Out:  out,
		Info: info,
		Idx:  idx,
	}

	// sequence types need conversion of inner type
	if seq, ok := t.(*types.SequenceType); ok {
		data.InnerInfo, data.InnerType = seq.Elem.DefaultParam()
		data.Inner = inoutGetToFromWasm(data.InnerType, data.InnerInfo, "__out", "__in", idx*100, tmpl)
	}
	if data.Info.Variadic {
		copy := *data.Info
		copy.Variadic = false
		data.Inner = inoutGetToFromWasm(data.Type, &copy, "__out", "__in", idx*100, tmpl)
		t = types.ChangeTemplateName(t, "variadic")
	}
	return convertType(t, data, tmpl) + "\n"
}

func inoutParamStart(t types.TypeRef, info *types.TypeInfo, out, in string, idx int, tmpl *template.Template) string {
	data := struct {
		Nullable bool
		Optional bool
		Info     *types.TypeInfo
		Type     types.TypeRef
		In, Out  string
		Idx      int
		AnyType  bool
	}{
		Nullable: info.Nullable,
		Optional: info.Option,
		Info:     info,
		Type:     t,
		In:       in,
		Out:      out,
		Idx:      idx,
	}
	_, data.AnyType = t.(*types.AnyType)
	return executeTemplateToString("param-start", data, true, tmpl)
}

func inoutParamEnd(info *types.TypeInfo, assign string, tmpl *template.Template) string {
	if info.Variadic {
		assign = ""
	}
	data := struct {
		Nullable bool
		Optional bool
		Info     *types.TypeInfo
		Assign   string
	}{
		Nullable: info.Nullable,
		Optional: info.Option,
		Info:     info,
		Assign:   assign,
	}
	return executeTemplateToString("param-end", data, true, tmpl)
}

func executeTemplateToString(name string, data interface{}, newLine bool, tmpl *template.Template) string {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		panic(err)
	}
	out := buf.String()
	// out = strings.Replace(out, "\n", " ", -1)
	out = strings.TrimSpace(out)
	if newLine || strings.Index(out, "\n") != -1 {
		out += "\n"
	}
	return out
}

func inoutDictionaryVariableStart(dict *dictionaryData, from bool, tmpl *template.Template) string {
	type elem struct {
		Name types.MethodName
		In   string
		Out  string
		Info *types.TypeInfo
	}
	data := struct {
		ParamList  []*elem
		ReleaseHdl bool
	}{}
	for _, m := range dict.Members {
		v := &elem{
			Name: m.Name,
			In:   m.toIn,
			Out:  m.toOut,
			Info: m.Type,
		}
		if from {
			v.In, v.Out = m.fromIn, m.fromOut
		}
		data.ParamList = append(data.ParamList, v)
	}
	return executeTemplateToString("start", data, true, tmpl)
}
