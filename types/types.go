package types

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gowebapi/webidlparser/ast"
)

type TypeRef interface {
	link(conv *Convert, inuse inuseLogic) TypeRef

	// Basic type infomation
	Basic() BasicInfo
	// Param type information
	Param(nullable, option, vardic bool) (info *TypeInfo, inner TypeRef)
	// DefaultParam return how this parameter should be processed by default
	DefaultParam() (info *TypeInfo, inner TypeRef)

	// the type is doing some allocation that needs manual release.
	NeedRelease() bool
}

const BuiltInPackage = "<built-in>"

func convertType(in ast.Type, exrType *extractTypes) TypeRef {
	ref := createRef(in, exrType)
	var ret TypeRef
	switch in := in.(type) {
	case *ast.TypeName:
		switch in.Name {
		case "boolean":
			ret = newPrimitiveType(in.Name, "bool", "Bool", false, false)
		case "short":
			ret = newPrimitiveType(in.Name, "int", "Int", false, true)
		case "unsigned short":
			ret = newPrimitiveType(in.Name, "int", "Int", false, true)
		case "long":
			ret = newPrimitiveType(in.Name, "int", "Int", false, false)
		case "unsigned long":
			ret = newPrimitiveType(in.Name, "uint", "Int", true, false)
		case "long long":
			ret = newPrimitiveType(in.Name, "int", "Int", false, false)
		case "unsigned long long":
			ret = newPrimitiveType(in.Name, "int", "Int", false, false)
		case "float":
			ret = newPrimitiveType(in.Name, "float32", "Float", false, true)
		case "unrestricted float":
			ret = newPrimitiveType(in.Name, "float32", "Float", false, true)
		case "double":
			ret = newPrimitiveType(in.Name, "float64", "Float", false, true)
		case "unrestricted double":
			ret = newPrimitiveType(in.Name, "float64", "Float", false, true)
		case "void":
			ret = newVoidType(in)
		case "DOMString":
			ret = newPrimitiveType(in.Name, "string", "String", false, false)
		case "USVString":
			ret = newPrimitiveType(in.Name, "string", "String", false, false)
		default:
			ret = newTypeNameRef(in, ref)
		}
	case *ast.AnyType:
		ret = newAnyType()
	case *ast.SequenceType:
		elem := convertType(in.Elem, exrType)
		if primitive, ok := elem.(*PrimitiveType); ok {
			if primitive.supportTypedArray {
				ret = newTypedArrayType(primitive)
			}
		}
		if ret == nil {
			ret = newSequenceType(elem)
		}
	case *ast.RecordType:
		panic(fmt.Sprintf("support not implemented: input source line %d", in.Line))
	case *ast.ParametrizedType:
		var elems []TypeRef
		for _, e := range in.Elems {
			elems = append(elems, convertType(e, exrType))
		}
		ret = newParametrizedType(in, in.Name, elems, ref)
	case *ast.UnionType:
		ret = newUnionType(in, exrType)
	case *ast.NullableType:
		inner := convertType(in.Type, exrType)
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
			needRelease: false,
		},
	}
}

func (t *AnyType) Basic() BasicInfo {
	ret := BasicInfo{
		Idl:      "any",
		Package:  BuiltInPackage,
		Def:      "js.Value",
		Internal: "<any>",
		Template: "any",
	}
	return TransformBasic(t, ret)
}

func (t *AnyType) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *AnyType) link(conv *Convert, inuse inuseLogic) TypeRef {
	return t
}

func (t *AnyType) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	ret := &TypeInfo{
		BasicInfo:   t.Basic(),
		InOut:       "js.Value",
		Input:       "js.Value",
		Output:      "js.Value",
		OutputInner: "js.Value",
		Pointer:     false,
		NeedRelease: false,
		Nullable:    false,
		Option:      option,
		Variadic:    variadic,
	}
	if variadic {
		ret.Def = "..." + ret.Def
		ret.InOut = "..." + ret.InOut
		ret.Input = "..." + ret.Input
		ret.Output = "[]" + ret.Output
	}
	return ret, t
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
	panic("not supported for this type")
	// return t.If.Basic()
}

func (t *interfaceType) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *interfaceType) link(conv *Convert, inuse inuseLogic) TypeRef {
	return t.If
}

func (t *interfaceType) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	panic("not supported for this type")
	// return t.If.Param(nullable, option, variadic)
}

type nullableType struct {
	Type TypeRef
}

var _ TypeRef = &nullableType{}

func newNullableType(inner TypeRef) *nullableType {
	return &nullableType{Type: inner}
}

func (t *nullableType) Basic() BasicInfo {
	return t.Type.Basic()
}

func (t *nullableType) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *nullableType) link(conv *Convert, inuse inuseLogic) TypeRef {
	t.Type = t.Type.link(conv, inuse)
	return t
}

func (t *nullableType) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	return t.Type.Param(true, option, variadic)
}

func (t *nullableType) NeedRelease() bool {
	return t.Type.NeedRelease()
}

// ParametrizedType is e.g. "Promise<any>"
type ParametrizedType struct {
	*Ref
	in        *ast.ParametrizedType
	ParamName string
	Elems     []TypeRef
	Type      TypeRef
}

var _ TypeRef = &ParametrizedType{}

func newParametrizedType(in *ast.ParametrizedType, name string, elems []TypeRef, ref *Ref) *ParametrizedType {
	// what types are parameterized? only Promise or is there more?
	if name != "Promise" && name != "FrozenArray" {
		panic("parameterized type name: " + name)
	}
	return &ParametrizedType{
		Ref:       ref,
		in:        in,
		ParamName: name,
		Elems:     elems,
	}
}

func (t *ParametrizedType) Basic() BasicInfo {
	basic := t.Type.Basic()
	basic.Template = "parametrized"
	return basic
}

func (t *ParametrizedType) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *ParametrizedType) link(conv *Convert, inuse inuseLogic) TypeRef {
	// names := []string{}
	for i := range t.Elems {
		inner := make(inuseLogic)
		t.Elems[i] = t.Elems[i].link(conv, inner)
		// names = append(names, t.Elems[i].Basic().Idl)
	}

	candidate := getIdlName(t.ParamName)
	if real, f := conv.Types[candidate]; f {
		t.Type = real.link(conv, inuse)
	} else {
		conv.failing(t, "reference to unknown type '%s' (%s)", candidate, t.in.Name)
		return t
	}
	return t
}

func (t *ParametrizedType) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	return newTypeInfo(t.Basic(), nullable, option, variadic, true, false, false), t
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
	Cast     bool

	// if this represent a primitive type that can be supported
	// by TypedArray, e.g. int8, int, float32 etc
	supportTypedArray bool
}

var _ TypeRef = &PrimitiveType{}

func newPrimitiveType(idl, lang, method string, cast, sta bool) *PrimitiveType {
	return &PrimitiveType{
		basicType: basicType{
			needRelease: false,
		},
		Idl:               idl,
		Lang:              lang,
		JsMethod:          method,
		Cast:              cast,
		supportTypedArray: sta,
	}
}

func (t *PrimitiveType) Basic() BasicInfo {
	basic := BasicInfo{
		Idl:      t.Idl,
		Package:  BuiltInPackage,
		Def:      t.Lang,
		Internal: "<primitive-internal-name>",
		Template: "primitive",
	}
	return TransformBasic(t, basic)
}

func (t *PrimitiveType) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *PrimitiveType) link(conv *Convert, inuse inuseLogic) TypeRef {
	return t
}

func (t *PrimitiveType) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	return newTypeInfo(t.Basic(), nullable, option, variadic, false, false, false), t
}

// RawJSType used when no conversion should take place and
// the raw underlying js.Value should be returned or inserted
// instead
type RawJSType struct {
}

var _ TypeRef = &RawJSType{}

func NewRawJSType() *RawJSType {
	return &RawJSType{}
}

func (t *RawJSType) Basic() BasicInfo {
	return BasicInfo{
		Idl:      "<rawjs>",
		Def:      "js.Value",
		Internal: "js.Value",
		Package:  BuiltInPackage,
		Template: "rawjs",
	}
}

func (t *RawJSType) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *RawJSType) link(conv *Convert, inuse inuseLogic) TypeRef {
	return t
}

func (t *RawJSType) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	return newTypeInfo(t.Basic(), nullable, option, variadic, false, false, false), t
}

func (t *RawJSType) NeedRelease() bool {
	return false
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
			Package:  BuiltInPackage,
			Def:      "def-sequence",
			Internal: "internal-sequence",
			Template: "sequence",
		},
	}
	return ret
}

func (t *SequenceType) Basic() BasicInfo {
	value := t.basic
	eb := t.Elem.Basic()
	_, prim := t.Elem.(*PrimitiveType)
	_, enum := t.Elem.(*Enum)
	_, cb := t.Elem.(*Callback)
	if prim || enum || cb {
		value.Def = "[]" + eb.Def
	} else {
		value.Def = "[]*" + eb.Def
	}
	return TransformBasic(t, value)
}

func (t *SequenceType) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *SequenceType) link(conv *Convert, inuse inuseLogic) TypeRef {
	inner := make(inuseLogic)
	t.Elem = t.Elem.link(conv, inner)
	if conv.HaveError {
		// if we have failing, we can't evaluate Basic() as e.g.
		// typeNameRef is returning faulty type when trigger an
		// error
		return t
	}
	return t
}

func (t *SequenceType) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	return newTypeInfo(t.Basic(), nullable, option, variadic, false, false, false), t
}

func (t *SequenceType) NeedRelease() bool {
	return t.Elem.NeedRelease()
}

type TypedArrayType struct {
	Elem  *PrimitiveType
	basic BasicInfo
}

var _ TypeRef = &TypedArrayType{}

func newTypedArrayType(primitive *PrimitiveType) *TypedArrayType {
	return &TypedArrayType{
		Elem: primitive,
		basic: BasicInfo{
			Idl:      "typed-array",
			Package:  BuiltInPackage,
			Def:      "js.Value",
			Internal: "typed-array",
			Template: "typedarray",
		},
	}
}

func (t *TypedArrayType) Basic() BasicInfo {
	return TransformBasic(t, t.basic)
}

func (t *TypedArrayType) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *TypedArrayType) link(conv *Convert, inuse inuseLogic) TypeRef {
	// assumes that PrimitiveType.link() doesn't do anything
	return t
}

func (t *TypedArrayType) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	return newTypeInfo(t.Basic(), nullable, option, variadic, false, false, false), t
}

func (t *TypedArrayType) NeedRelease() bool {
	return false
}

type typeNameRef struct {
	*Ref
	in         *ast.TypeName
	Underlying TypeRef
}

var _ TypeRef = &typeNameRef{}

func newTypeNameRef(in *ast.TypeName, ref *Ref) *typeNameRef {
	return &typeNameRef{
		in:  in,
		Ref: ref,
	}
}

func (t *typeNameRef) Basic() BasicInfo {
	panic("not supported by this type")
	// return t.Underlying.Basic()
}

func (t *typeNameRef) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *typeNameRef) link(conv *Convert, inuse inuseLogic) TypeRef {
	candidate := getIdlName(t.in.Name)
	if real, f := conv.Types[candidate]; f {
		t.Underlying = real.link(conv, inuse)
		return t.Underlying
	} else {
		conv.failing(t, "reference to unknown type '%s' (%s)", candidate, t.in.Name)
		return t
	}
}

func (t *typeNameRef) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	panic("not supported by this type")
	// return t.Underlying.Param(nullable, option, variadic)
}

func (t *typeNameRef) NeedRelease() bool {
	return t.Underlying.NeedRelease()
}

type UnionType struct {
	in    *ast.UnionType
	ref   *Ref
	name  string
	Types []TypeRef
	basic BasicInfo
	use   bool
}

var _ TypeRef = &UnionType{}

func newUnionType(in *ast.UnionType, exrTypes *extractTypes) *UnionType {
	ret := &UnionType{in: in, ref: createRef(in, exrTypes)}
	for _, t := range in.Types {
		ret.Types = append(ret.Types, convertType(t, exrTypes))
	}
	return ret
}

func (t *UnionType) Basic() BasicInfo {
	return TransformBasic(t, t.basic)
}

func (t *UnionType) DefaultParam() (info *TypeInfo, inner TypeRef) {
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
		Package:  BuiltInPackage,
		Def:      t.name + "Union",
		Internal: "union" + t.name,
		Template: "union",
	}
	return t
}

func (t *UnionType) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	return newTypeInfo(t.Basic(), nullable, option, variadic, true, false, false), t
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
	basic := BasicInfo{
		Idl:      "void",
		Package:  BuiltInPackage,
		Def:      "",
		Internal: "void",
		Template: "void",
	}
	return TransformBasic(t, basic)
}

func (t *voidType) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *voidType) link(conv *Convert, inuse inuseLogic) TypeRef {
	return t
}

func (t *voidType) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	return &TypeInfo{
		BasicInfo:   t.Basic(),
		InOut:       "",
		Input:       "",
		Output:      "",
		OutputInner: "",
		NeedRelease: false,
		Nullable:    false,
		Option:      false,
		Variadic:    false,
	}, t
}
