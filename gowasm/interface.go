package gowasm

import (
	"io"
	"strings"
	"text/template"
	"wasm/generator/types"
)

const interfaceTmplInput = `
{{define "header"}}
type {{.If.Name.Public}} struct {
}

{{end}}

{{define "get-static-attribute"}}
func {{.Name.Public}}() {{.Type}} {
	klass := js.Global().Get("{{.If.Name.Public}}")
	value := klass.Get("{{.Name.Idl}}")
	{{.From}}
	return ret
}
{{end}}

{{define "set-static-attribute"}}
fix set-static-attribute
{{end}}

{{define "static-method-start"}}
func {{.Name.Public}}({{.To.Params}}) ({{.ReturnList}}) {
	_klass := js.Global().Get("{{.If.Name.Idl}}")
	_method := _klass.Get("{{.Name.Idl}}")
	{{end}}
{{define "static-method-invoke"}}
	{{if not .IsVoidReturn}}ret :={{end}} _method.Invoke({{.To.AllOut}})
{{end}}
{{define "static-method-end"}}
	{{if not .IsVoidReturn}}result = value{{end}}
	{{if .To.ReleaseHdl}}release = _releaseList{{end}}
	return
}
{{end}}

{{define "constructor-start"}}
func {{.Name.Public}}({{.To.Params}}) ({{.ReturnList}}) {
	_klass := js.Global().Get("{{.If.Name.Idl}}")
{{end}}
{{define "constructor-invoke"}}
	_returned := _klass.New()
{{end}}
{{define "constructor-end"}}
	result = _result
	return
}
{{end}}

`

var interfaceTmpl = template.Must(template.New("interface").Parse(interfaceTmplInput))

type interfaceData struct {
	If *types.Interface
}

type interfaceAttribute struct {
	Name types.Name
	Type string
	From string
	If   *types.Interface
}

type interfaceMethod struct {
	Name         types.Name
	If           *types.Interface
	Return       string
	ReturnList   string
	IsVoidReturn bool
	To           *inoutData
}

func writeInterface(dst io.Writer, input types.Type) error {
	value := input.(*types.Interface)
	data := &interfaceData{
		If: value,
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
	return nil
}

func writeInterfaceVars(vars []*types.IfVar, main *types.Interface, get, set string, dst io.Writer) error {
	for _, a := range vars {
		in := &interfaceAttribute{
			Name: a.Name(),
			Type: typeDefine(a.Type),
			From: inoutGetToFromWasm(a.Type, "ret", "value", inoutFromTmpl),
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
	to := setupInOutWasmData(m.Params, "@name", "_p%d")
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
	lang = typeDefine(t)
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
