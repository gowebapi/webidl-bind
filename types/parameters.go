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

func (t *extractTypes) convertParams(list []*ast.Parameter) []*Parameter {
	params := []*Parameter{}
	for _, pi := range list {
		po := t.convertParam(pi)
		params = append(params, po)
	}
	return params
}

func (t *extractTypes) convertParam(in *ast.Parameter) *Parameter {
	t.warningTrue(len(in.Annotations) == 0, in, "parameter: unsupported annotation")
	return &Parameter{
		in:       in,
		Name:     toCamelCase(in.Name, false),
		Type:     convertType(in.Type),
		Optional: in.Optional,
		Variadic: in.Variadic,
	}
}
