package types

import (
	"html/template"

	"github.com/gowebapi/webidlparser/ast"
)

// this file contains different "protocol" that types can
// implement, like iterable, maplike etc.
// this is implemented by parsing an internal file before
// reading the real one.

const protocolTemplateInput = `
{{define "iterable"}}
interface mixin {{.Name}}Iterable {
	{{.Name}}EntryIterator entries();
	void forEach( {{.Name}}ForEach callback, optional any optionalThisForCallbackArgument );
	{{.Name}}KeyIterator keys();
	{{.Name}}ValueIterator values();
};
{{.Name}} includes {{.Name}}Iterable;

callback {{.Name}}ForEach = void ( {{.Name}}_TypeDef_Value currentValue, long currentIndex, {{.Name}} listObj);

interface {{.Name}}EntryIterator {
	{{.Name}}EntryIteratorValue next();
};
dictionary {{.Name}}EntryIteratorValue {
	sequence<any> value;
	boolean done;
};

interface {{.Name}}KeyIterator {
	{{.Name}}KeyIteratorValue next();
};
dictionary {{.Name}}KeyIteratorValue {
	{{.Name}}_TypeDef_Key value;
	boolean done;
};

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

func queueProtocolIterableOne(name string, value ast.Type, ref *Ref, et *extractTypes) {
	protocolAddIterableBase(name, ref, et)
	key := &ast.TypeName{Name: "unsigned long"}
	protocolAddTypeDef(name+"_TypeDef_Key", key, ref, et)
	protocolAddTypeDef(name+"_TypeDef_Value", value, ref, et)
}

func queueProtocolIterableTwo(name string, key, value ast.Type, ref *Ref, et *extractTypes) {
	protocolAddIterableBase(name, ref, et)
	protocolAddTypeDef(name+"_TypeDef_Key", value, ref, et)
	protocolAddTypeDef(name+"_TypeDef_Value", value, ref, et)
}

func protocolAddIterableBase(name string, ref *Ref, et *extractTypes) {
	data := struct {
		Name string
	}{
		Name: name,
	}
	if err := protocolTemplate.ExecuteTemplate(&et.protocol, "iterable", &data); err != nil {
		et.failing(ref, "internal error: template execute: %s", err)
	}
}

func protocolAddTypeDef(name string, value ast.Type, ref *Ref, et *extractTypes) {
	typedef := &ast.Typedef{
		Base: ast.Base{
			Line: ref.Line,
		},
		Name: name,
		Type: value,
	}
	et.Typedef(typedef)
}
