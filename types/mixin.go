package types

import (
	"fmt"

	"github.com/dennwc/webidl/ast"
)

type Mixin struct {
	source *ast.Mixin
	Name   string

	Consts       []*IfConst
	Vars         []*IfVar
	StaticVars   []*IfVar
	Method       []*IfMethod
	StaticMethod []*IfMethod
}

// var _ Type = &Mixin{}

type Includes struct {
	source *ast.Includes
}

func (t *extractTypes) convertMixin(in *ast.Mixin) (*Mixin, bool) {
	ret := &Mixin{
		source: in,
		Name:   in.Name,
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
			ret.StaticMethod = append(ret.StaticMethod, mo)
		} else {
			mo := t.convertInterfaceMethod(mi)
			ret.Method = append(ret.Method, mo)
		}
	}
	for _, a := range in.Annotations {
		t.warning(a, "unsupported interface annotation '%s'", a.Name)
	}
	return ret, in.Partial
}

func (t *extractTypes) convertIncludes(in *ast.Includes) *Includes {
	ret := &Includes{
		source: in,
	}
	return ret
}
