package types

import (
	"github.com/dennwc/webidl/ast"
)

type Parameter struct {
	in       *ast.Parameter
	Type     TypeRef
	Optional bool
	Variadic bool
	Name     string
}

func (t *extractTypes) convertParam(in *ast.Parameter) *Parameter {
	t.assertTrue(len(in.Annotations) == 0, in, "unsupported annotation")
	return &Parameter{
		in:       in,
		Name:     toCamelCase(in.Name, false),
		Type:     convertType(in.Type),
		Optional: in.Optional,
		Variadic: in.Variadic,
	}
}
