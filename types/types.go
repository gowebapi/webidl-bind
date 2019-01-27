package types

import (
	"fmt"

	"github.com/dennwc/webidl/ast"
)

type TypeRef interface {
	link(conv *Convert)
	NeedRelease() bool
}

func convertType(in ast.Type) TypeRef {
	var ret TypeRef
	switch in := in.(type) {
	case *ast.TypeName:
		switch in.Name {
		case "void":
			ret = newVoidType(in)
		case "DOMString":
			ret = newPrimitiveType(in.Name, "string")
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

type PrimitiveType struct {
	basicType
	Idl  string
	Lang string
}

var _ TypeRef = &PrimitiveType{}

func newPrimitiveType(idl, lang string) *PrimitiveType {
	return &PrimitiveType{
		basicType: basicType{
			needRelease: false,
		},
		Idl:  idl,
		Lang: lang,
	}
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
	candidate := fromIdlName("", t.in.Name).Idl
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
