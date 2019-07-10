package types

import (
	"github.com/gowebapi/webidlparser/ast"
)

type Parameter struct {
	ref      *Ref
	Type     TypeRef
	Optional bool
	Variadic bool
	Name     string
}

func (p *Parameter) copy() *Parameter {
	r := *p.ref
	dst := &Parameter{
		ref:      &r,
		Type:     p.Type,
		Optional: p.Optional,
		Variadic: p.Variadic,
		Name:     p.Name,
	}
	return dst
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
		ref:      ref,
		Name:     fixLangName(toCamelCase(name, false)),
		Type:     convertType(in.Type, t),
		Optional: in.Optional,
		Variadic: in.Variadic,
	}
}
