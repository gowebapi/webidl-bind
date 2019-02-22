package gowasm

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/gowebapi/webidl-bind/types"
)

const interfaceTmplInput = `
{{define "header"}}
// interface: {{.Type.Idl}}
type {{.Type.Def}} struct {
	{{if .If.Inherits}}
		{{.If.Inherits.Basic.Def}}
	{{else}}
		// Value_JS holds a reference to a javascript value
		Value_JS js.Value
	{{end}}
}

{{if not .If.Inherits}}
func (_this *{{.Type.Def}}) JSValue() js.Value {
	return _this.Value_JS
}
{{end}}

// {{.Type.Def}}FromJS is casting a js.Wrapper into {{.Type.Def}}.
func {{.Type.Def}}FromJS(value js.Wrapper) {{.Type.Output}} {
	input := value.JSValue()
	{{if .Type.Pointer}}
	if input.Type() == js.TypeNull {
		return nil
	}
	{{end}}
	ret := {{if .Type.Pointer}}&{{end}} {{.Type.Def}} { }
	ret.Value_JS = input
	return ret
}

{{end}}

{{define "const-var"}}
	const {{.If.ConstPrefix}}{{.Const.Name.Def}}{{.If.ConstSuffix}} {{.Info.Def}} = {{.Value}}
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
	var ret {{.Type.Output}}
	value := _this.Value_JS.Get("{{.Name.Idl}}")
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
	_this.Value_JS.Set("{{.Name.Idl}}", input)
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
	{{if not .IsVoidReturn}}_result = _converted{{end}}
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
	var (
		_args {{.ArgVar}} 
		_end int 
	)
{{end}}
{{define "object-method-invoke"}}
	{{if not .IsVoidReturn}}_returned :={{end}} _this.Value_JS.Call("{{.Name.Idl}}", _args[0:_end]... )
{{end}}
{{define "object-method-end"}}
	{{if not .IsVoidReturn}}_result = _converted{{end}}
	{{if .To.ReleaseHdl}}_release = _releaseList{{end}}
	return
}
{{end}}

{{define "callback-header"}}
// {{.Type.Def}} is a callback interface.
type {{.Type.Def}} interface {
	{{range .Methods}}
		{{.Name.Def}} ( {{.To.Params}} ) ( {{ .ReturnList }} )
	{{end}}
}

// {{.Type.Def}}Value is javascript reference value for callback interface {{.Type.Def}}.
// This is holding the underlaying javascript object.
type {{.Type.Def}}Value struct {
	// Value is the underlying javascript object{{if .If.FunctionCB}} or function{{end}}.
	Value     js.Value
	// Functions is the underlying function objects that is allocated for the interface callback
	Functions [ {{len .Methods}} ] js.Func

	// Go interface to invoke
	impl      {{.Type.Def}}
	{{if eq (len .Methods) 1 }}
		function  func ( {{(index .Methods 0).To.Params}} ) ( {{(index .Methods 0).ReturnList}} )
		useInvoke bool
	{{end}}
}

// JSValue is returning the javascript object that implements this callback interface
func (t * {{.Type.Def}}Value ) JSValue() js.Value {
	return t.Value
}

// Release is releasing all resources that is allocated.
func (t * {{.Type.Def}}Value ) Release() {
	for i := range t.Functions {
		if t.Functions[i].Type() != js.TypeUndefined {
			t.Functions[i].Release()
		}
	}
}

// New{{.Type.Def}} is allocating a new javascript object that
// implements {{.Type.Def}}.
func New{{.Type.Def}} ( callback {{.Type.Def}} ) * {{.Type.Def}}Value {
	ret := & {{.Type.Def}}Value { impl: callback }
	ret.Value = js.Global().Get("Object").New()
	{{range $idx, $value := .Methods}}
		ret.Functions[ {{$idx}} ] = ret.allocate{{$value.Name.Def}} ()
		ret.Value.Set( "{{$value.Name.Idl}}" , ret.Functions[ {{$idx}} ] )
	{{end}}
	return ret
}

{{if eq (len .Methods) 1 }}
// New{{.Type.Def}}Func is allocating a new javascript 
// {{if .If.FunctionCB}}function{{else}}object{{end}} is implements
// {{.Type.Def}} interface.
func New{{.Type.Def}}Func( f func( {{(index .Methods 0).To.Params}} ) ( {{(index .Methods 0).ReturnList}} ) ) * {{.Type.Def}}Value {
	{{if .If.FunctionCB}}
		// single function will result in javascript function type, not an object
		ret := & {{.Type.Def}}Value { function: f }
		ret.Functions[0] = ret.allocate{{ (index .Methods 0).Name.Def }}()
		ret.Value = ret.Functions[0].Value
	{{else}}
		ret := & {{.Type.Def}}Value { Impl: implementation }
		ret.Value = js.Global().Get("Object").New()
		{{range $idx, $value := .Methods}}
			ret.Functions[ {{$idx}} ] = ret.allocate{{$value.Name.Def}} ()
			ret.Value.Set( "{{$value.Name.Idl}}" , ret.Function[ {{$idx}} ] )
		{{end}}
	{{end}}
	return ret
}
{{end}}

// {{.Type.Def}}FromJS is taking an javascript object that reference to a 
// callback interface and return a corresponding interface that can be used
// to invoke on that element.
func {{.Type.Def}}FromJS(value js.Wrapper) * {{.Type.Def}}Value {
	input := value.JSValue()
	if input.Type() == js.TypeObject {
		return &{{.Type.Def}}Value { Value: input }
	}
	{{if eq (len .Methods) 1}}
	if input.Type() == js.TypeFunction {
		return &{{.Type.Def}}Value { Value: input, useInvoke: true }
	}
	{{else}}
		// note: have no support for functions, method count: {{len .Methods}}
	{{end}}
	panic("unsupported type")
}

{{end}}

{{define "callback-allocate-start"}}
func ( t * {{.If.Basic.Def}}Value ) allocate{{.Name.Def}} () js.Func {
	return js.FuncOf(func (this js.Value, args []js.Value) interface{} {

{{end}}

{{define "callback-allocate-invoke"}}
	{{if not .IsVoidReturn}}
		var _returned {{.Return}}
	{{end}}
	{{if eq (len .If.Method) 1}}
		if t.function != nil {
			{{if not .IsVoidReturn}}_returned={{end}} t.function ( {{range $idx, $value := .To.ParamList}} {{if ge $idx 1}},{{end}} _p{{$idx}} {{end}} )
		} else {
	{{end}}
		{{if not .IsVoidReturn}}_returned={{end}} t.impl.{{.Name.Def}} ( {{range $idx, $value := .To.ParamList}} {{if ge $idx 1}},{{end}} _p{{$idx}} {{end}} )
	{{if eq (len .If.Method) 1}}
		}
	{{end}}
{{end}}

{{define "callback-allocate-end"}}
		{{if not .IsVoidReturn}}
			return _converted
		{{else}}
			// returning no return value
			return nil
		{{end}}
	})
}
{{end}}

{{define "callback-invoke-start"}}
func (_this * {{.If.Basic.Def}}Value ) {{.Name.Def}} ( {{.To.Params}} ) ( {{.ReturnList}} ) {
	{{if eq (len .If.Method) 1 }}
		if _this.function != nil {
			{{if not .IsVoidReturn}}return {{end}} _this.function ( {{range $idx, $value := .To.ParamList}} {{if ge $idx 1}},{{end}} {{$value.Name}} {{end}} )
		}
	{{end}}
	if _this.impl != nil {
		{{if not .IsVoidReturn}}return {{end}} _this.impl. {{.Name.Def}} ( {{range $idx, $value := .To.ParamList}} {{if ge $idx 1}},{{end}} {{$value.Name}} {{end}} )
	}
	var (
		_args {{.ArgVar}} 
		_end int 
	)
{{end}}
{{define "callback-invoke-invoke"}}
	{{if not .IsVoidReturn}}
		var _returned js.Value
	{{end}}
	{{if eq (len .If.Method) 1 }}
	if _this.useInvoke {
		// invoke a javascript function
		{{if not .IsVoidReturn}}_returned ={{end}} _this.Value.Invoke(_args[0:_end]... )
	} else {
	{{end}}
		{{if not .IsVoidReturn}}_returned ={{end}} _this.Value.Call("{{.Name.Idl}}", _args[0:_end]... )
	{{if eq (len .If.Method) 1 }}
	}
	{{end}}

{{end}}
{{define "callback-invoke-end"}}
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
	Method       *types.IfMethod
	Return       string
	ReturnList   string
	IsVoidReturn bool
	To           *inoutData
	ArgVar       string
}

func writeInterface(dst io.Writer, input types.Type) error {
	value := input.(*types.Interface)
	if value.Callback {
		return writeCallbackInterface(value, dst)
	}
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
	if err := writeInterfaceMethods(value.StaticMethod, value, "static-method", useIn, dst); err != nil {
		return err
	}
	if value.Constructor != nil {
		if err := writeInterfaceMethod(value.Constructor, value, "constructor", useIn, dst); err != nil {
			return err
		}
	}
	if err := writeInterfaceVars(value.Vars, value, "get-object-attribute", "set-object-attribute", dst); err != nil {
		return err
	}
	if err := writeInterfaceMethods(value.Method, value, "object-method", useIn, dst); err != nil {
		return err
	}
	return nil
}

// callback interface code
func writeCallbackInterface(value *types.Interface, dst io.Writer) error {
	if err := writeInterfaceConst(value.Consts, value, dst); err != nil {
		return err
	}
	// first we setup method information
	methods := []*interfaceMethod{}
	for _, m := range value.Method {
		to := setupInOutWasmData(m.Params, "args[%d]", "_p%d", useOut)
		retLang, retList, isVoid := calculateMethodReturn(m.Return, to.ReleaseHdl)
		in := &interfaceMethod{
			Name:         *m.Name(),
			Return:       retLang,
			ReturnList:   retList,
			IsVoidReturn: isVoid,
			If:           value,
			Method:       m,
			To:           to,
			ArgVar:       calculateMethodArgsSize(to),
		}
		methods = append(methods, in)
	}

	data := struct {
		Type    *types.TypeInfo
		Ref     types.TypeRef
		Methods []*interfaceMethod
		If      *types.Interface
	}{
		Methods: methods,
		If:      value,
	}
	data.Type, data.Ref = value.DefaultParam()
	if err := interfaceTmpl.ExecuteTemplate(dst, "callback-header", data); err != nil {
		return err
	}
	for _, m := range methods {
		assign := ""
		if err := writeInterfaceCallbackMethod(m, assign, "callback-allocate", dst); err != nil {
			return err
		}
	}
	if err := writeInterfaceMethods(value.Method, value, "callback-invoke", useOut, dst); err != nil {
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
			Value string
		}{
			Const: a,
			Idx:   idx,
			If:    main,
			Value: a.Value,
		}
		data.Info, data.Type = a.Type.DefaultParam()
		if types.IsString(data.Type) {
			data.Value = "\"" + data.Value + "\""
		}
		if err := interfaceTmpl.ExecuteTemplate(dst, "const-var", data); err != nil {
			return err
		}
	}
	return nil
}

func writeInterfaceVars(vars []*types.IfVar, main *types.Interface, get, set string, dst io.Writer) error {
	for _, a := range vars {
		typ, ref := a.Type.DefaultParam()
		ret := ""
		if a.Type.NeedRelease() {
			ret = "(_release ReleasableApiResource)"
		}
		idx := 0
		from := inoutParamStart(ref, typ, "ret", "value", idx, useOut, inoutFromTmpl)
		from += inoutGetToFromWasm(ref, typ, "ret", "value", idx, useOut, inoutFromTmpl)
		from += inoutParamEnd(typ, "", inoutFromTmpl)
		to := inoutParamStart(ref, typ, "input", "value", idx, useIn, inoutToTmpl)
		to += inoutGetToFromWasm(ref, typ, "input", "value", idx, useIn, inoutToTmpl)
		to += inoutParamEnd(typ, "", inoutToTmpl)
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

func writeInterfaceMethods(methods []*types.IfMethod, main *types.Interface, tmpl string, use useInOut, dst io.Writer) error {
	for _, m := range methods {
		if err := writeInterfaceMethod(m, main, tmpl, use, dst); err != nil {
			return err
		}
	}
	return nil
}

func writeInterfaceMethod(m *types.IfMethod, main *types.Interface, tmpl string, use useInOut, dst io.Writer) error {
	to := setupInOutWasmData(m.Params, "@name@", "_p%d", use)
	retLang, retList, isVoid := calculateMethodReturn(m.Return, to.ReleaseHdl)
	in := &interfaceMethod{
		Name:         *m.Name(),
		Return:       retLang,
		ReturnList:   retList,
		IsVoidReturn: isVoid,
		If:           main,
		Method:       m,
		To:           to,
		ArgVar:       calculateMethodArgsSize(to),
	}
	if err := interfaceTmpl.ExecuteTemplate(dst, tmpl+"-start", in); err != nil {
		return err
	}
	assign := "_args[%d] = _p%d; _end++"
	if err := writeInOutToWasm(in.To, assign, useIn, dst); err != nil {
		return err
	}
	if err := interfaceTmpl.ExecuteTemplate(dst, tmpl+"-invoke", in); err != nil {
		return err
	}
	if !in.IsVoidReturn {
		result := setupInOutWasmForType(in.Method.Return, "_what_return_name", "_returned", "_converted", useOut)
		if err := writeInOutFromWasm(result, "", useOut, dst); err != nil {
			return err
		}
	}
	if err := interfaceTmpl.ExecuteTemplate(dst, tmpl+"-end", in); err != nil {
		return err
	}
	return nil
}

func writeInterfaceCallbackMethod(in *interfaceMethod, assign, tmpl string, dst io.Writer) error {
	if err := interfaceTmpl.ExecuteTemplate(dst, tmpl+"-start", in); err != nil {
		return err
	}
	if err := writeInOutFromWasm(in.To, assign, useOut, dst); err != nil {
		return err
	}
	if err := interfaceTmpl.ExecuteTemplate(dst, tmpl+"-invoke", in); err != nil {
		return err
	}
	if !in.IsVoidReturn {
		result := setupInOutWasmForType(in.Method.Return, "_what_return_name", "_returned", "_converted", useOut)
		if err := writeInOutToWasm(result, "", useOut, dst); err != nil {
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
