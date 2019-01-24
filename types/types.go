package types

import (
	"fmt"

	"github.com/dennwc/webidl/ast"
)

type TypeRef interface {
	link(conv *Convert)
}

func convertType(in ast.Type) TypeRef {
	var ret TypeRef
	switch in := in.(type) {
	case *ast.TypeName:
		switch in.Name {
		case "void":
			ret = &VoidType{in: in}
		case "DOMString":
			ret = &PrimitiveType{
				Idl:  in.Name,
				Lang: "string",
			}
		default:
			ret = &TypeNameRef{in: in}
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

type PrimitiveType struct {
	Idl  string
	Lang string
}

func (t *PrimitiveType) link(conv *Convert) {

}

type TypeNameRef struct {
	in         *ast.TypeName
	Name       Name
	Underlying Type
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

type VoidType struct {
	in *ast.TypeName
}

func (t *VoidType) link(conv *Convert) {

}
