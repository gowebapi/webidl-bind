package types

import (
	"fmt"

	"github.com/gowebapi/webidlparser/ast"
)

type mixin struct {
	source *ast.Mixin
	Name   string
	ref    *Ref

	Consts       []*IfConst
	Vars         []*IfVar
	StaticVars   []*IfVar
	Method       []*IfMethod
	StaticMethod []*IfMethod
}

// var _ Type = &Mixin{}

type includes struct {
	source *ast.Includes
	// Name is the target
	Name string
	// Source points to a mixin
	Source string
	ref    *Ref
}

func (t *extractTypes) convertMixin(in *ast.Mixin) (*mixin, bool) {
	ret := &mixin{
		source: in,
		Name:   in.Name,
		ref:    createRef(in, t),
	}
	for _, raw := range in.Members {
		mi, ok := raw.(*ast.Member)
		if !ok {
			panic(fmt.Sprintf("unsupported %T", raw))
		}
		if mi.Const {
			mo := t.convertInterfaceConst(mi)
			ret.Consts = append(ret.Consts, mo)
		} else if mi.Attribute && mi.Static {
			mo := t.convertInterfaceVar(mi)
			ret.StaticVars = append(ret.StaticVars, mo)
		} else if mi.Attribute {
			mo := t.convertInterfaceVar(mi)
			ret.Vars = append(ret.Vars, mo)
		} else if mi.Static {
			mo := t.convertInterfaceMethod(mi)
			if mo != nil {
				ret.StaticMethod = append(ret.StaticMethod, mo)
			}
		} else {
			mo := t.convertInterfaceMethod(mi)
			if mo != nil {
				ret.Method = append(ret.Method, mo)
			}
		}
	}
	for _, a := range in.Annotations {
		aref := createRef(a, t)
		t.warning(aref, "unsupported interface annotation '%s'", a.Name)
	}
	return ret, in.Partial
}

func (t *extractTypes) convertIncludes(in *ast.Includes) *includes {
	ret := &includes{
		source: in,
		Name:   in.Name,
		Source: in.Source,
		ref:    createRef(in, t),
	}
	return ret
}

func (t *mixin) merge(m *mixin, conv *Convert) {
	t.Consts = append(t.Consts, m.Consts...)
	t.Vars = append(t.Vars, m.Vars...)
	t.StaticVars = append(t.StaticVars, m.StaticVars...)
	t.Method = append(t.Method, m.Method...)
	t.StaticMethod = append(t.StaticMethod, m.StaticMethod...)
}

func (t *mixin) SourceReference() *Ref {
	return t.ref
}

func (t *includes) SourceReference() *Ref {
	return t.ref
}
