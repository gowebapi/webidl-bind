package gowasm

import (
	"io"
	"strings"
	"text/template"
	"wasm/generator/types"
)

const interfaceTmplInput = `
{{define "header"}}
type {{.Type.Def}} struct {
	value js.Value
}

func (t *{{.Type.Def}}) JSValue() js.Value {
	return t.value
}

func {{.Type.Internal}}FromWasm(input js.Value) {{.Type.InOut}} {
	return {{if .Type.Pointer}}&{{end}} {{.Type.Def}} {value: input}
}

func {{.Type.Internal}}ToWasm(input {{.Type.InOut}}) js.Value {
	return input.value
}

{{end}}

{{define "get-static-attribute"}}
func {{.Name.Def}} () {{.Type.InOut}} {
	klass := js.Global().Get("{{.If.Basic.Idl}}")
	value := klass.Get("{{.Name.Idl}}")
	{{.From}}
	return ret
}
{{end}}

{{define "set-static-attribute"}}
func Set{{.Name.Def}} ( value {{.Type.InOut}} ) {
	klass := js.Global().Get("{{.If.Basic.Idl}}")
	{{.To}}
	klass.Set("{{.Name.Idl}}", input)
}
{{end}}

{{define "get-object-attribute"}}
func (_this * {{.If.Basic.Def}} ) {{.Name.Def}} () {{.Type.InOut}} {
	value := _this.value.Get("{{.Name.Idl}}")
	{{.From}}
	return ret
}
{{end}}

{{define "set-object-attribute"}}
func (_this * {{.If.Basic.Def}} ) Set{{.Name.Def}} ( value {{.Type.InOut}} )  {
	{{.To}}
	_this.value.Set("{{.Name.Idl}}", input)
}
{{end}}


{{define "static-method-start"}}
func {{.Name.Def}}({{.To.Params}}) ({{.ReturnList}}) {
	_klass := js.Global().Get("{{.If.Basic.Idl}}")
	_method := _klass.Get("{{.Name.Idl}}")
{{end}}
{{define "static-method-invoke"}}
	{{if not .IsVoidReturn}}ret :={{end}} _method.Invoke( {{.To.AllOut}} )
{{end}}
{{define "static-method-end"}}
	{{if not .IsVoidReturn}}result = value{{end}}
	{{if .To.ReleaseHdl}}release = _releaseList{{end}}
	return
}
{{end}}

{{define "constructor-start"}}
func {{.Name.Def}}({{.To.Params}}) ({{.ReturnList}}) {
	_klass := js.Global().Get("{{.If.Basic.Idl}}")
{{end}}
{{define "constructor-invoke"}}
	_returned := _klass.New({{.To.AllOut}})
{{end}}
{{define "constructor-end"}}
	result = _result
	return
}
{{end}}


{{define "object-method-start"}}
func ( _this * {{.If.Basic.Def}} ) {{.Name.Def}} ( {{.To.Params}} ) ( {{.ReturnList}} ) {
	_method := _this.value.Get("{{.Name.Idl}}")
{{end}}
{{define "object-method-invoke"}}
	{{if not .IsVoidReturn}}ret :={{end}} _method.Invoke({{.To.AllOut}})
{{end}}
{{define "object-method-end"}}
	{{if not .IsVoidReturn}}result = value{{end}}
	{{if .To.ReleaseHdl}}release = _releaseList{{end}}
	return
}
{{end}}

`

var interfaceTmpl = template.Must(template.New("interface").Parse(interfaceTmplInput))

type interfaceData struct {
	If   *types.Interface
	Type *types.TypeInfo
}

type interfaceAttribute struct {
	Name types.MethodName
	Type *types.TypeInfo
	From string
	To   string
	If   *types.Interface
}

type interfaceMethod struct {
	Name         types.MethodName
	If           *types.Interface
	Return       string
	ReturnList   string
	IsVoidReturn bool
	To           *inoutData
}

func writeInterface(dst io.Writer, input types.Type) error {
	value := input.(*types.Interface)
	data := &interfaceData{
		If:   value,
		Type: value.DefaultParam(),
	}
	if err := interfaceTmpl.ExecuteTemplate(dst, "header", data); err != nil {
		return err
	}
	if err := writeInterfaceVars(value.StaticVars, value, "get-static-attribute", "set-static-attribute", dst); err != nil {
		return err
	}
	if err := writeInterfaceMethods(value.StaticMethod, value, "static-method", dst); err != nil {
		return err
	}
	if value.Constructor != nil {
		if err := writeInterfaceMethod(value.Constructor, value, "constructor", dst); err != nil {
			return err
		}
	}
	if err := writeInterfaceVars(value.Vars, value, "get-object-attribute", "set-object-attribute", dst); err != nil {
		return err
	}
	if err := writeInterfaceMethods(value.Method, value, "object-method", dst); err != nil {
		return err
	}
	return nil
}

func writeInterfaceVars(vars []*types.IfVar, main *types.Interface, get, set string, dst io.Writer) error {
	for _, a := range vars {
		in := &interfaceAttribute{
			Name: a.Name(),
			Type: a.Type.DefaultParam(),
			From: inoutGetToFromWasm(a.Type, nil, "ret", "value", inoutFromTmpl),
			To:   inoutGetToFromWasm(a.Type, nil, "input", "value", inoutToTmpl),
			If:   main,
		}
		if err := interfaceTmpl.ExecuteTemplate(dst, get, in); err != nil {
			return err
		}
		if !a.Readonly {
			if err := interfaceTmpl.ExecuteTemplate(dst, set, in); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeInterfaceMethods(methods []*types.IfMethod, main *types.Interface, tmpl string, dst io.Writer) error {
	for _, m := range methods {
		if err := writeInterfaceMethod(m, main, tmpl, dst); err != nil {
			return err
		}
	}
	return nil
}

func writeInterfaceMethod(m *types.IfMethod, main *types.Interface, tmpl string, dst io.Writer) error {
	to := setupInOutWasmData(m.Params, "@name@", "_p%d")
	retLang, retList, isVoid := calculateMethodReturn(m.Return, to.ReleaseHdl)
	in := &interfaceMethod{
		Name:         m.Name(),
		Return:       retLang,
		ReturnList:   retList,
		IsVoidReturn: isVoid,
		If:           main,
		To:           to,
	}
	if err := interfaceTmpl.ExecuteTemplate(dst, tmpl+"-start", in); err != nil {
		return err
	}
	if err := writeInOutToWasm(in.To, dst); err != nil {
		return err
	}
	if err := interfaceTmpl.ExecuteTemplate(dst, tmpl+"-invoke", in); err != nil {
		return err
	}
	if !isVoid {
		result := setupInOutWasmForType(m.Return, "_returned", "_result")
		if err := writeInOutFromWasm(result, dst); err != nil {
			return err
		}
	}
	if err := interfaceTmpl.ExecuteTemplate(dst, tmpl+"-end", in); err != nil {
		return err
	}
	return nil
}

func calculateMethodReturn(t types.TypeRef, releaseHdl bool) (lang, list string, isVoid bool) {
	info := t.Basic()
	lang = info.Def
	isVoid = types.IsVoid(t)

	candidate := []string{}
	if !isVoid {
		candidate = append(candidate, "result "+lang)
	}
	if releaseHdl {
		candidate = append(candidate, "release ReleasableApiResource")
	}
	list = strings.Join(candidate, ", ")
	return
}
