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

func convertParam(in *ast.Parameter) (*Parameter, error) {
	if len(in.Annotations) > 0 {
		return nil, UnsupportedAnnotationErr
	}
	return &Parameter{
		in:       in,
		Name:     toCamelCase(in.Name, false),
		Type:     convertType(in.Type),
		Optional: in.Optional,
		Variadic: in.Variadic,
	}, nil
}
