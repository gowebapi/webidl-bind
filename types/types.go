package types

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dennwc/webidl/ast"
)

type TypeRef interface {
	link(conv *Convert, inuse inuseLogic) TypeRef

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
		ret = newUnionType(in)
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
		Package:  "<build-in>",
		Def:      "js.Value",
		Internal: "<any>",
		Template: "any",
	}
	return ret
}

func (t *AnyType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *AnyType) link(conv *Convert, inuse inuseLogic) TypeRef {
	return t
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

type interfaceType struct {
	basicType
	If *Interface
}

// InterfaceType must implement TypeRef
var _ TypeRef = &interfaceType{}

func newInterfaceType(link *Interface) *interfaceType {
	return &interfaceType{
		basicType: basicType{
			needRelease: false,
		},
		If: link,
	}
}

func (t *interfaceType) Basic() BasicInfo {
	return t.If.Basic()
}

func (t *interfaceType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *interfaceType) link(conv *Convert, inuse inuseLogic) TypeRef {
	return t.If
}

func (t *interfaceType) Param(nullable, option, vardict bool) *TypeInfo {
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

func (t *NullableType) link(conv *Convert, inuse inuseLogic) TypeRef {
	t.Type = t.Type.link(conv, inuse)
	return t
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

func (t *ParametrizedType) link(conv *Convert, inuse inuseLogic) TypeRef {
	for i := range t.Elems {
		inner := make(inuseLogic)
		t.Elems[i] = t.Elems[i].link(conv, inner)
	}
	return t
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

func (t *PrimitiveType) link(conv *Convert, inuse inuseLogic) TypeRef {
	return t
}

func (t *PrimitiveType) Param(nullable, option, vardict bool) *TypeInfo {
	return newTypeInfo(t.Basic(), nullable, option, vardict, false, false, false)
}

type SequenceType struct {
	Elem  TypeRef
	basic BasicInfo
}

var _ TypeRef = &SequenceType{}

func newSequenceType(elem TypeRef) *SequenceType {
	ret := &SequenceType{
		Elem: elem,
		basic: BasicInfo{
			Idl:      "idl-sequence",
			Package:  "<built-in>",
			Def:      "def-sequence",
			Internal: "internal-sequence",
			Template: "sequence",
		},
	}
	return ret
}

func (t *SequenceType) Basic() BasicInfo {
	return t.basic
}

func (t *SequenceType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *SequenceType) link(conv *Convert, inuse inuseLogic) TypeRef {
	inner := make(inuseLogic)
	t.Elem = t.Elem.link(conv, inner)
	return t
}

func (t *SequenceType) Param(nullable, option, vardict bool) *TypeInfo {
	return newTypeInfo(t.basic, nullable, option, vardict, false, false, false)
}

func (t *SequenceType) NeedRelease() bool {
	return t.Elem.NeedRelease()
}

type typeNameRef struct {
	in         *ast.TypeName
	Underlying TypeRef
}

var _ TypeRef = &typeNameRef{}

func newTypeNameRef(in *ast.TypeName) *typeNameRef {
	return &typeNameRef{
		in: in,
	}
}

func (t *typeNameRef) Basic() BasicInfo {
	return t.Underlying.Basic()
}

func (t *typeNameRef) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *typeNameRef) link(conv *Convert, inuse inuseLogic) TypeRef {
	candidate := getIdlName(t.in.Name)
	if real, f := conv.Types[candidate]; f {
		t.Underlying = real.link(conv, inuse)
		return t.Underlying
	} else {
		conv.failing(t.in, "reference to unknown type '%s' (%s)", candidate, t.in.Name)
		return t
	}
}

func (t *typeNameRef) Param(nullable, option, vardict bool) *TypeInfo {
	return t.Underlying.Param(nullable, option, vardict)
}

func (t *typeNameRef) NeedRelease() bool {
	return t.Underlying.NeedRelease()
}

type UnionType struct {
	in    *ast.UnionType
	name  string
	Types []TypeRef
	basic BasicInfo
	use   bool
}

var _ TypeRef = &UnionType{}

func newUnionType(in *ast.UnionType) *UnionType {
	ret := &UnionType{in: in}
	for _, t := range in.Types {
		ret.Types = append(ret.Types, convertType(t))
	}
	return ret
}

func (t *UnionType) Basic() BasicInfo {
	return t.basic
}

func (t *UnionType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *UnionType) link(conv *Convert, inuse inuseLogic) TypeRef {
	if t.use {
		return t
	}
	t.use = true
	conv.Unions = append(conv.Unions, t)
	names := []string{}
	for idx := range t.Types {
		inner := make(inuseLogic)
		t.Types[idx] = t.Types[idx].link(conv, inner)
		n := toCamelCase(t.Basic().Idl, true)
		names = append(names, n)
	}
	sort.Strings(names)
	t.name = strings.Join(names, "")
	t.basic = BasicInfo{
		Idl:      t.name + "Union",
		Package:  "<built-in>",
		Def:      t.name + "Union",
		Internal: "union" + t.name,
		Template: "union",
	}
	return t
}

func (t *UnionType) Param(nullable, option, vardict bool) *TypeInfo {
	return newTypeInfo(t.basic, nullable, option, vardict, true, false, false)
}

func (t *UnionType) NeedRelease() bool {
	for _, t := range t.Types {
		if t.NeedRelease() {
			return true
		}
	}
	return false
}

type voidType struct {
	basicType
	in *ast.TypeName
}

var _ TypeRef = &voidType{}

func newVoidType(in *ast.TypeName) *voidType {
	return &voidType{
		basicType: basicType{
			needRelease: false,
		},
		in: in,
	}
}

func (t *voidType) Basic() BasicInfo {
	return BasicInfo{
		Idl:      "void",
		Package:  "<built-in>",
		Def:      "",
		Internal: "void",
		Template: "void",
	}
}

func (t *voidType) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *voidType) link(conv *Convert, inuse inuseLogic) TypeRef {
	return t
}

func (t *voidType) Param(nullable, option, vardict bool) *TypeInfo {
	return &TypeInfo{
		BasicInfo:   t.Basic(),
		InOut:       "",
		NeedRelease: false,
		Nullable:    false,
		Option:      false,
		Vardict:     false,
	}
}
