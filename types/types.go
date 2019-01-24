package types

import (
	"fmt"

	"github.com/dennwc/webidl/ast"
)

type TypeRef interface {
	link(conv *Convert)
}

func convertType(in ast.Type) TypeRef {
	switch in := in.(type) {
	case *ast.TypeName:
		switch in.Name {
		case "void":
			return &VoidType{in: in}
		}
		return &TypeNameRef{in: in}
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
	default:
		msg := fmt.Sprint("unknown type %T", in)
		panic(msg)
	}
}

type VoidType struct {
	in *ast.TypeName
}

func (t *VoidType) link(conv *Convert) {

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
		conv.failing(t.in.NodeBase(), "reference to unknown type '%s' (%s)", candidate, t.in.Name)
	}
}
