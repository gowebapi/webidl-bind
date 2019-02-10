package types

import (
	"github.com/gowebapi/webidlparser/ast"
)

type Parameter struct {
	in       *ast.Parameter
	ref      *Ref
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
	ref := createRef(in, t)
	t.warningTrue(len(in.Annotations) == 0, ref, "parameter: unsupported annotation")
	name := getIdlName(in.Name)
	return &Parameter{
		in:       in,
		ref:      ref,
		Name:     fixLangName(toCamelCase(name, false)),
		Type:     convertType(in.Type, t),
		Optional: in.Optional,
		Variadic: in.Variadic,
	}
}
