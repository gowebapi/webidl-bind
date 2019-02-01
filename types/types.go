package types

import (
	"fmt"

	"github.com/dennwc/webidl/ast"
)

type TypeRef interface {
	link(conv *Convert)

	// Basic type infomation
	Basic() BasicInfo
	// Param type information
	Param(nullable, option, vardic bool) *TypeInfo
	// DefaultParam return how this parameter should be processed by default
	DefaultParam() *TypeInfo

	// the type is doing some allocation that needs manual release.
	NeedRelease() bool
}

func convertType(in ast.Type) TypeRef {
	var ret TypeRef
	switch in := in.(type) {
	case *ast.TypeName:
		switch in.Name {
		case "boolean":
			ret = newPrimitiveType(in.Name, "bool", "Bool")
		case "short":
			ret = newPrimitiveType(in.Name, "int", "Int")
		case "unsigned short":
			ret = newPrimitiveType(in.Name, "int", "Int")
		case "long":
			ret = newPrimitiveType(in.Name, "int", "Int")
		case "unsigned long":
			ret = newPrimitiveType(in.Name, "uint", "Int")
		case "long long":
			ret = newPrimitiveType(in.Name, "int", "Int")
		case "unsigned long long":
			ret = newPrimitiveType(in.Name, "int", "Int")
		case "double":
			ret = newPrimitiveType(in.Name, "float64", "Float")
		case "unrestricted double":
			ret = newPrimitiveType(in.Name, "float64", "Float")
		case "void":
			ret = newVoidType(in)
		case "DOMString":
			ret = newPrimitiveType(in.Name, "string", "String")
		case "USVString":
			ret = newPrimitiveType(in.Name, "string", "String")
		default:
			ret = newTypeNameRef(in)
		}
	case *ast.AnyType:
		ret = newAnyType()
	case *ast.SequenceType:
		elem := convertType(in.Elem)
		ret = newSequenceType(elem)
	case *ast.RecordType:
		panic(fmt.Sprintf("support not implemented: input source line %d", in.Line))
	case *ast.ParametrizedType:
		var elems []TypeRef
		for _, e := range in.Elems {
			elems = append(elems, convertType(e))
		}
		ret = newParametrizedType(in.Name, elems)
	case *ast.UnionType:
		ret = newUnionType(in.Types)
	case *ast.NullableType:
		inner := convertType(in.Type)
		ret = newNullableType(inner)
	}
	if ret == nil {
		msg := fmt.Sprintf("unknown type %T: %#v", in, in)
		panic(msg)
	}
	return ret
}

type basicType struct {
	needRelease bool
}

func (t *basicType) link(conv *Convert) {

}

func (t *basicType) NeedRelease() bool {
	return t.needRelease
}

type AnyType struct {
	basicType
}

var _ TypeRef = &AnyType{}

func newAnyType() *AnyType {
	return &AnyType{
		basicType: basicType{
			// if the any type is a js.Func or js.TypeArray a
			// release handle is needed
			needRelease: true,
		},
	}
}

func (t *AnyType) Basic() BasicInfo {
	ret := BasicInfo{
		Idl:      "any",
		Package:  "<build-in-any>",
		Def:      "interface{}",
		Internal: "<any>",
		Template: "any",
	}
	return ret
}

func (t *AnyType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *AnyType) Param(nullable, option, vardict bool) *TypeInfo {
	// TODO shoud returned any type be js.Value ?
	ret := &TypeInfo{
		BasicInfo:   t.Basic(),
		InOut:       "interface{}",
		Pointer:     false,
		NeedRelease: false,
		Nullable:    false,
		Option:      option,
		Vardict:     vardict,
	}
	if vardict {
		ret.Def = "..." + ret.Def
		ret.InOut = "..." + ret.InOut
	}
	return ret
}

type InterfaceType struct {
	basicType
	If *Interface
}

// InterfaceType must implement TypeRef
var _ TypeRef = &InterfaceType{}

func newInterfaceType(link *Interface) *InterfaceType {
	return &InterfaceType{
		basicType: basicType{
			needRelease: false,
		},
		If: link,
	}
}

func (t *InterfaceType) Basic() BasicInfo {
	return t.If.Basic()
}

func (t *InterfaceType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *InterfaceType) Param(nullable, option, vardict bool) *TypeInfo {
	return t.If.Param(nullable, option, vardict)
}

type NullableType struct {
	Type TypeRef
}

var _ TypeRef = &NullableType{}

func newNullableType(inner TypeRef) *NullableType {
	return &NullableType{Type: inner}
}

func (t *NullableType) Basic() BasicInfo {
	return t.Type.Basic()
}

func (t *NullableType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *NullableType) link(conv *Convert) {
	t.Type.link(conv)
}

func (t *NullableType) Param(nullable, option, vardict bool) *TypeInfo {
	return t.Type.Param(true, option, vardict)
}

func (t *NullableType) NeedRelease() bool {
	return t.Type.NeedRelease()
}

type ParametrizedType struct {
	ParamName string
	Elems     []TypeRef
}

var _ TypeRef = &ParametrizedType{}

func newParametrizedType(name string, elems []TypeRef) *ParametrizedType {
	return &ParametrizedType{
		ParamName: name,
		Elems:     elems,
	}
}

func (t *ParametrizedType) Basic() BasicInfo {
	panic("todo")
}

func (t *ParametrizedType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *ParametrizedType) link(conv *Convert) {
	for _, t := range t.Elems {
		t.link(conv)
	}
}

func (t *ParametrizedType) Param(nullable, option, vardict bool) *TypeInfo {
	panic("todo")
}

func (t *ParametrizedType) NeedRelease() bool {
	for _, t := range t.Elems {
		if t.NeedRelease() {
			return true
		}
	}
	return false
}

type PrimitiveType struct {
	basicType
	Idl      string
	Lang     string
	JsMethod string
}

var _ TypeRef = &PrimitiveType{}

func newPrimitiveType(idl, lang, method string) *PrimitiveType {
	return &PrimitiveType{
		basicType: basicType{
			needRelease: false,
		},
		Idl:      idl,
		Lang:     lang,
		JsMethod: method,
	}
}

func (t *PrimitiveType) Basic() BasicInfo {
	return BasicInfo{
		Idl:      t.Idl,
		Package:  "<build-in>",
		Def:      t.Lang,
		Internal: "<primitive-internal-name>",
		Template: "primitive",
	}
}

func (t *PrimitiveType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *PrimitiveType) Param(nullable, option, vardict bool) *TypeInfo {
	return newTypeInfo(t.Basic(), nullable, option, vardict, false, false, false)
}

type SequenceType struct {
	Elem TypeRef
}

var _ TypeRef = &SequenceType{}

func newSequenceType(elem TypeRef) *SequenceType {
	return &SequenceType{
		Elem: elem,
	}
}

func (t *SequenceType) Basic() BasicInfo {
	panic("todo")
}

func (t *SequenceType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *SequenceType) link(conv *Convert) {
	t.Elem.link(conv)
}

func (t *SequenceType) Param(nullable, option, vardict bool) *TypeInfo {
	panic("todo")
}

func (t *SequenceType) NeedRelease() bool {
	return t.Elem.NeedRelease()
}

type TypeNameRef struct {
	in         *ast.TypeName
	Underlying Type
}

var _ TypeRef = &TypeNameRef{}

func newTypeNameRef(in *ast.TypeName) *TypeNameRef {
	return &TypeNameRef{
		in: in,
	}
}

func (t *TypeNameRef) Basic() BasicInfo {
	return t.Underlying.Basic()
}

func (t *TypeNameRef) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *TypeNameRef) link(conv *Convert) {
	candidate := getIdlName(t.in.Name)
	if real, f := conv.Types[candidate]; f {
		t.Underlying = real
	} else {
		conv.failing(t.in, "reference to unknown type '%s' (%s)", candidate, t.in.Name)
	}
}

func (t *TypeNameRef) Param(nullable, option, vardict bool) *TypeInfo {
	return t.Underlying.Param(nullable, option, vardict)
}

func (t *TypeNameRef) NeedRelease() bool {
	return t.Underlying.NeedRelease()
}

type UnionType struct {
	Types []TypeRef
}

var _ TypeRef = &UnionType{}

func newUnionType(input []ast.Type) *UnionType {
	ret := &UnionType{}
	for _, t := range input {
		ret.Types = append(ret.Types, convertType(t))
	}
	return ret
}

func (t *UnionType) Basic() BasicInfo {
	panic("todo")
}

func (t *UnionType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *UnionType) link(conv *Convert) {
	for _, t := range t.Types {
		t.link(conv)
	}
}

func (t *UnionType) Param(nullable, option, vardict bool) *TypeInfo {
	panic("todo")
}

func (t *UnionType) NeedRelease() bool {
	for _, t := range t.Types {
		if t.NeedRelease() {
			return true
		}
	}
	return false
}

type VoidType struct {
	basicType
	in *ast.TypeName
}

var _ TypeRef = &VoidType{}

func newVoidType(in *ast.TypeName) *VoidType {
	return &VoidType{
		basicType: basicType{
			needRelease: false,
		},
		in: in,
	}
}

func (t *VoidType) Basic() BasicInfo {
	return BasicInfo{
		Idl:      "void",
		Package:  "<built-in-void>",
		Def:      "",
		Internal: "void",
		Template: "void",
	}
}

func (t *VoidType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *VoidType) Param(nullable, option, vardict bool) *TypeInfo {
	return &TypeInfo{
		BasicInfo:   t.Basic(),
		InOut:       "",
		NeedRelease: false,
		Nullable:    false,
		Option:      false,
		Vardict:     false,
	}
}
