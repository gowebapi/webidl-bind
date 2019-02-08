package gowasm

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/gowebapi/webidlgenerator/types"
)

const interfaceTmplInput = `
{{define "header"}}
// interface: {{.Type.Idl}}
type {{.Type.Def}} struct {
	{{if .If.Inherits}} {{.If.Inherits.Basic.Def}}
	{{else}}
	value js.Value
	{{end}}
}

func (t *{{.Type.Def}}) JSValue() js.Value {
	return t.value
}

// {{.Type.Def}}FromJS is casting a js.Value into {{.Type.Def}}.
func {{.Type.Def}}FromJS(input js.Value) {{.Type.InOut}} {
	{{if .Type.Pointer}}
	if input.Type() == js.TypeNull {
		return nil
	}
	{{end}}
	ret := {{if .Type.Pointer}}&{{end}} {{.Type.Def}} { }
	ret.value = input
	return ret
}

{{end}}

{{define "const-var"}}
	const {{.If.ConstPrefix}}{{.Const.Name.Def}}{{.If.ConstSuffix}} {{.Info.Def}} = {{.Const.Value}}
{{end}}

{{define "get-static-attribute"}}
// {{.Name.Def}} returning attribute '{{.Name.Idl}}' with
// type {{.Type.Def}} (idl: {{.Type.Idl}}).
func {{.Name.Def}} () {{.Type.Output}} {
	var ret {{.Type.Output}}
	_klass := js.Global() {{if not .If.Global}} .Get("{{.If.Basic.Idl}}") {{end}}
	value := _klass.Get("{{.Name.Idl}}")
	{{.From}}
	return ret
}
{{end}}

{{define "set-static-attribute"}}
// {{.Name.Def}} returning attribute '{{.Name.Idl}}' with
// type {{.Type.Def}} (idl: {{.Type.Idl}}).
func Set{{.Name.Def}} ( value {{.Type.Input}} ) {{.Ret}} {
	{{if len .Ret}}var _releaseList releasableApiResourceList{{end}}
	_klass := js.Global() {{if not .If.Global}} .Get("{{.If.Basic.Idl}}") {{end}}
	{{.To}}
	_klass.Set("{{.Name.Idl}}", input)
	{{if len .Ret}}return{{end}}
}
{{end}}

{{define "get-object-attribute"}}
// {{.Name.Def}} returning attribute '{{.Name.Idl}}' with
// type {{.Type.Def}} (idl: {{.Type.Idl}}).
func (_this * {{.If.Basic.Def}} ) {{.Name.Def}} () {{.Type.Output}} {
	var ret {{.Type.InOut}}
	value := _this.value.Get("{{.Name.Idl}}")
	{{.From}}
	return ret
}
{{end}}

{{define "set-object-attribute"}}
// Set{{.Name.Def}} setting attribute '{{.Name.Idl}}' with
// type {{.Type.Def}} (idl: {{.Type.Idl}}).
func (_this * {{.If.Basic.Def}} ) Set{{.Name.Def}} ( value {{.Type.Input}} ) {{.Ret}} {
	{{if len .Ret}}var _releaseList releasableApiResourceList{{end}}
	{{.To}}
	_this.value.Set("{{.Name.Idl}}", input)
	{{if len .Ret}}return{{end}}
}
{{end}}


{{define "static-method-start"}}
func {{.Name.Def}}({{.To.Params}}) ({{.ReturnList}}) {
	_klass := js.Global() {{if not .If.Global}} .Get("{{.If.Basic.Idl}}") {{end}}
	_method := _klass.Get("{{.Name.Idl}}")
	var (
		_args {{.ArgVar}} 
		_end int 
	)
{{end}}
{{define "static-method-invoke"}}
	{{if not .IsVoidReturn}}_returned :={{end}} _method.Invoke( _args[0:_end]... )
{{end}}
{{define "static-method-end"}}
	{{if not .IsVoidReturn}}_result = value{{end}}
	{{if .To.ReleaseHdl}}_release = _releaseList{{end}}
	return
}
{{end}}

{{define "constructor-start"}}
func {{.Name.Def}}({{.To.Params}}) ({{.ReturnList}}) {
	_klass := js.Global().Get("{{.If.Basic.Idl}}")
	var (
		_args {{.ArgVar}} 
		_end int 
	)
{{end}}
{{define "constructor-invoke"}}
	_returned := _klass.New( _args[0:_end]... )
{{end}}
{{define "constructor-end"}}
	_result = _converted
	return
}
{{end}}

{{define "object-method-start"}}
func ( _this * {{.If.Basic.Def}} ) {{.Name.Def}} ( {{.To.Params}} ) ( {{.ReturnList}} ) {
	_method := _this.value.Get("{{.Name.Idl}}")
	var (
		_args {{.ArgVar}} 
		_end int 
	)
{{end}}
{{define "object-method-invoke"}}
	{{if not .IsVoidReturn}}_returned :={{end}} _method.Invoke( _args[0:_end]... )
{{end}}
{{define "object-method-end"}}
	{{if not .IsVoidReturn}}_result = _converted{{end}}
	{{if .To.ReleaseHdl}}_release = _releaseList{{end}}
	return
}
{{end}}
`

var interfaceTmpl = template.Must(template.New("interface").Parse(interfaceTmplInput))

type interfaceData struct {
	If   *types.Interface
	Type *types.TypeInfo
	Ref  types.TypeRef
}

type interfaceAttribute struct {
	Name types.MethodName
	Type *types.TypeInfo
	Ref  types.TypeRef
	From string
	To   string
	If   *types.Interface
	Ret  string
}

type interfaceMethod struct {
	Name         types.MethodName
	If           *types.Interface
	Return       string
	ReturnList   string
	IsVoidReturn bool
	To           *inoutData
	ArgVar       string
}

func writeInterface(dst io.Writer, input types.Type) error {
	value := input.(*types.Interface)
	data := &interfaceData{
		If: value,
	}
	data.Type, data.Ref = value.DefaultParam()
	if !value.Global {
		if err := interfaceTmpl.ExecuteTemplate(dst, "header", data); err != nil {
			return err
		}
	}
	if err := writeInterfaceConst(value.Consts, value, dst); err != nil {
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

func writeInterfaceConst(vars []*types.IfConst, main *types.Interface, dst io.Writer) error {
	for idx, a := range vars {
		data := struct {
			Const *types.IfConst
			Idx   int
			Type  types.TypeRef
			Info  *types.TypeInfo
			If    *types.Interface
		}{
			Const: a,
			Idx:   idx,
			If:    main,
		}
		data.Info, data.Type = a.Type.DefaultParam()
		if err := interfaceTmpl.ExecuteTemplate(dst, "const-var", data); err != nil {
			return err
		}
	}
	return nil
}

func writeInterfaceVars(vars []*types.IfVar, main *types.Interface, get, set string, dst io.Writer) error {
	for idx, a := range vars {
		typ, ref := a.Type.DefaultParam()
		ret := ""
		if a.Type.NeedRelease() {
			ret = "(_release ReleasableApiResource)"
		}
		from := inoutParamStart(ref, typ, "ret", "value", idx, inoutFromTmpl)
		from += inoutGetToFromWasm(ref, typ, "ret", "value", idx, inoutFromTmpl)
		from += inoutParamEnd(typ, "", inoutFromTmpl)
		to := inoutParamStart(ref, typ, "input", "value", idx, inoutToTmpl)
		to += inoutGetToFromWasm(ref, typ, "input", "value", idx, inoutToTmpl)
		to += inoutParamStart(ref, typ, "input", "value", idx, inoutToTmpl)
		in := &interfaceAttribute{
			Name: *a.Name(),
			Type: typ,
			Ref:  ref,
			From: from,
			To:   to,
			If:   main,
			Ret:  ret,
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
		Name:         *m.Name(),
		Return:       retLang,
		ReturnList:   retList,
		IsVoidReturn: isVoid,
		If:           main,
		To:           to,
		ArgVar:       calculateMethodArgsSize(to),
	}
	if err := interfaceTmpl.ExecuteTemplate(dst, tmpl+"-start", in); err != nil {
		return err
	}
	assign := "_args[%d] = _p%d; _end++"
	if err := writeInOutToWasm(in.To, assign, dst); err != nil {
		return err
	}
	if err := interfaceTmpl.ExecuteTemplate(dst, tmpl+"-invoke", in); err != nil {
		return err
	}
	if !isVoid {
		result := setupInOutWasmForType(m.Return, "_what_return_name", "_returned", "_converted")
		if err := writeInOutFromWasm(result, "", dst); err != nil {
			return err
		}
	}
	if err := interfaceTmpl.ExecuteTemplate(dst, tmpl+"-end", in); err != nil {
		return err
	}
	return nil
}

func calculateMethodReturn(t types.TypeRef, releaseHdl bool) (lang, list string, isVoid bool) {
	info, _ := t.DefaultParam()
	lang = info.Output
	isVoid = types.IsVoid(t)

	candidate := []string{}
	if !isVoid {
		candidate = append(candidate, "_result "+lang)
	}
	if releaseHdl {
		candidate = append(candidate, "_release ReleasableApiResource")
	}
	list = strings.Join(candidate, ", ")
	return
}

func calculateMethodArgsSize(method *inoutData) string {
	for _, p := range method.ParamList {
		if p.Info.Variadic {
			return fmt.Sprintf("[]interface{} = make([]interface{}, %d + len(%s))",
				len(method.ParamList)-1, p.Name)
		}
	}
	return fmt.Sprintf("[%d]interface{}", len(method.ParamList))
}
