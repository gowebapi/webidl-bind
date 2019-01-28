package types

import (
	"fmt"

	"github.com/dennwc/webidl/ast"
)

type TypeRef interface {
	link(conv *Convert)

	// the type is doing some allocation that needs manual release.
	NeedRelease() bool

	// what template suffix to use
	TemplateName() (string, TemplateNameFlags)
}

// TemplateNameFlags highlight inner structure. A bitwise value.
type TemplateNameFlags int

const (
	NoTnFlag  TemplateNameFlags = 0
	AnyTnFlag                   = (1 << iota)
	PointerTnFlag
	NullableTnFlag
)

func convertType(in ast.Type) TypeRef {
	var ret TypeRef
	switch in := in.(type) {
	case *ast.TypeName:
		switch in.Name {
		case "void":
			ret = newVoidType(in)
		case "DOMString":
			ret = newPrimitiveType(in.Name, "string", "String")
		default:
			ret = newTypeNameRef(in)
		}
	case *ast.AnyType:
		panic("support not implemented")
	case *ast.SequenceType:
		panic("support not implemented")
	case *ast.RecordType:
		panic("support not implemented")
	case *ast.ParametrizedType:
		panic("support not implemented")
	case *ast.UnionType:
		panic("support not implemented")
	case *ast.NullableType:
		panic("support not implemented")
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

func (t *InterfaceType) TemplateName() (string, TemplateNameFlags) {
	return "interface-type", NoTnFlag
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

func (t *PrimitiveType) TemplateName() (string, TemplateNameFlags) {
	return "primitive", NoTnFlag
}

type TypeNameRef struct {
	in         *ast.TypeName
	Name       Name
	Underlying Type
}

var _ TypeRef = &TypeNameRef{}

func newTypeNameRef(in *ast.TypeName) *TypeNameRef {
	return &TypeNameRef{
		in: in,
	}
}

func (t *TypeNameRef) link(conv *Convert) {
	candidate := fromIdlName("", t.in.Name, false).Idl
	if real, f := conv.Types[candidate]; f {
		t.Name = real.Name()
		t.Underlying = real
	} else {
		conv.failing(t.in, "reference to unknown type '%s' (%s)", candidate, t.in.Name)
	}
}

func (t *TypeNameRef) NeedRelease() bool {
	return t.Underlying.NeedRelease()
}

func (t *TypeNameRef) TemplateName() (string, TemplateNameFlags) {
	return t.Underlying.TemplateName()
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

func (t VoidType) TemplateName() (string, TemplateNameFlags) {
	return "void", NoTnFlag
}
