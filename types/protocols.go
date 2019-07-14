package types

import (
	"text/template"

	"github.com/gowebapi/webidlparser/ast"
)

// this file contains different "protocol" that types can
// implement, like iterable, maplike etc.
// this is implemented by parsing an internal file before
// reading the real one.

const protocolTemplateInput = `
{{define "stringifier"}}
interface mixin {{.Name}}Stringifier {
	USVString toString();
};
{{.Name}} includes {{.Name}}Stringifier;
{{end}}


{{define "iterable"}}
interface mixin {{.Name}}Iterable {
	{{.Name}}EntryIterator entries();
	void forEach( {{.Name}}ForEach callback, optional any optionalThisForCallbackArgument );
	{{.Name}}KeyIterator keys();
	{{.Name}}ValueIterator values();
};
{{.Name}} includes {{.Name}}Iterable;
callback {{.Name}}ForEach = void ( {{.Name}}_TypeDef_Value currentValue, long currentIndex, {{.Name}} listObj);
{{template "entry-iterator" .}}
{{template "key-iterator" .}}
{{template "value-iterator" .}}
{{end}}


{{define "maplike"}}
partial interface {{.Name}} {
	readonly attribute long size;
	{{.Name}}EntryIterator entries();
	void forEach( {{.Name}}ForEach callback, optional any optionalThisForCallbackArgument );
	{{.Name}}KeyIterator keys();
	{{.Name}}ValueIterator values();
	{{.Name}}_TypeDef_Value? get({{.Name}}_TypeDef_Key key);
	boolean has({{.Name}}_TypeDef_Key key);
{{if not .ReadOnly}}
	[ReplaceOnOverride] void clear();
	[ReplaceOnOverride] boolean delete({{.Name}}_TypeDef_Key key);
	[ReplaceOnOverride] {{.Name}} set({{.Name}}_TypeDef_Key key, {{.Name}}_TypeDef_Value value);
{{end}}
};
callback {{.Name}}ForEach = void ( {{.Name}}_TypeDef_Value currentValue, {{.Name}}_TypeDef_Key currentKey, {{.Name}} listObj);
{{template "entry-iterator" .}}
{{template "key-iterator" .}}
{{template "value-iterator" .}}
{{end}}


{{define "setlike"}}
// Note: In the ECMAScript language binding, the API for interacting with
// the set entries is similar to that available on ECMAScript Set objects.
// If the readonly keyword is used, this includes entries, forEach, has,
// keys, values, @@iterator methods, and a size getter.
// For read–write setlikes, it also includes add, clear, and delete methods.
partial interface {{.Name}} {
	readonly attribute long size;
	{{.Name}}EntryIterator entries();
	void forEach( {{.Name}}ForEach callback, optional any optionalThisForCallbackArgument );
	{{.Name}}KeyIterator keys();
	{{.Name}}ValueIterator values();
	{{.Name}}_TypeDef_Value? get({{.Name}}_TypeDef_Key key);
	boolean has({{.Name}}_TypeDef_Key key);

{{if not .ReadOnly}}
	[ReplaceOnOverride] {{.Name}} add({{.Name}}_TypeDef_Value value);
	[ReplaceOnOverride] void clear();
	[ReplaceOnOverride] boolean delete({{.Name}}_TypeDef_Value value);
{{end}}
};
callback {{.Name}}ForEach = void ( {{.Name}}_TypeDef_Value currentValue, {{.Name}}_TypeDef_Value currentValueAgain, {{.Name}} listObj);
interface {{.Name}}EntryIterator {
	{{.Name}}EntryIteratorValue next();
};
dictionary {{.Name}}EntryIteratorValue {
	sequence< {{.Name}}_TypeDef_Value > value;
	boolean done;
};
{{template "key-iterator" .}}
{{template "value-iterator" .}}
{{end}}

{{define "entry-iterator"}}
interface {{.Name}}EntryIterator {
	{{.Name}}EntryIteratorValue next();
};
dictionary {{.Name}}EntryIteratorValue {
	sequence<any> value;
	boolean done;
};
{{end}}

{{define "key-iterator"}}
interface {{.Name}}KeyIterator {
	{{.Name}}KeyIteratorValue next();
};
dictionary {{.Name}}KeyIteratorValue {
	{{.Name}}_TypeDef_Key value;
	boolean done;
};
{{end}}

{{define "value-iterator"}}
interface {{.Name}}ValueIterator {
	{{.Name}}ValueIteratorValue next();
};
dictionary {{.Name}}ValueIteratorValue {
	{{.Name}}_TypeDef_Value value;
	boolean done;
};
{{end}}
`

var protocolTemplate = template.Must(template.New("protocol").Parse(protocolTemplateInput))

func (et *extractTypes) queueProtocolInterfaceStringifier(name string, ref *Ref) {
	et.protocolAddTemplate("stringifier", name, false, ref)
}

func (et *extractTypes) queueProtocolIterableOne(name string, value ast.Type, ref *Ref) {
	et.protocolAddTemplate("iterable", name, false, ref)
	key := &ast.TypeName{Name: "unsigned long"}
	et.protocolAddTypeDef(name+"_TypeDef_Key", key, ref)
	et.protocolAddTypeDef(name+"_TypeDef_Value", value, ref)
}

func (et *extractTypes) queueProtocolIterableTwo(name string, key, value ast.Type, ref *Ref) {
	et.protocolAddTemplate("iterable", name, false, ref)
	et.protocolAddTypeDef(name+"_TypeDef_Key", key, ref)
	et.protocolAddTypeDef(name+"_TypeDef_Value", value, ref)
}

func (et *extractTypes) queueProtocolMaplike(name string, readonly bool, key, elem ast.Type, ref *Ref) {
	et.protocolAddTemplate("maplike", name, readonly, ref)
	et.protocolAddTypeDef(name+"_TypeDef_Key", key, ref)
	et.protocolAddTypeDef(name+"_TypeDef_Value", elem, ref)
}

func (et *extractTypes) queueProtocolSetlike(name string, readonly bool, elem ast.Type, ref *Ref) {
	et.protocolAddTemplate("setlike", name, readonly, ref)
	et.protocolAddTypeDef(name+"_TypeDef_Key", elem, ref)
	et.protocolAddTypeDef(name+"_TypeDef_Value", elem, ref)
}

func (et *extractTypes) protocolAddTemplate(tmpl, name string, readonly bool, ref *Ref) {
	data := struct {
		Name     string
		ReadOnly bool
	}{
		Name:     name,
		ReadOnly: readonly,
	}
	if err := protocolTemplate.ExecuteTemplate(&et.protocol, tmpl, &data); err != nil {
		et.failing(ref, "internal error: template execute: %s", err)
	}
}

func (et *extractTypes) protocolAddTypeDef(name string, value ast.Type, ref *Ref) {
	typedef := &ast.Typedef{
		Base: ast.Base{
			Line: ref.Line,
		},
		Name: name,
		Type: value,
	}
	et.Typedef(typedef)
}
