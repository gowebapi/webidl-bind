package types

import (
	"github.com/dennwc/webidl/ast"
)

// Enum type
type Enum struct {
	standardType
	source *ast.Enum
	Values []EnumValue
}

// EnumValue is a single enum value
type EnumValue struct {
	Idl string
	Go  string
}

func (t *extractTypes) convertEnum(in *ast.Enum) *Enum {
	t.assertTrue(len(in.Annotations) == 0, in, "unsupported annotation")
	ret := &Enum{
		standardType: standardType{
			base: in.NodeBase(),
			name: fromIdlName(t.main.setup.Package, in.Name),
		},
		source: in,
		Values: []EnumValue{},
	}

	for _, v := range in.Values {
		if b, ok := v.(*ast.BasicLiteral); ok {
			v := b.Value
			v = clipString(v)
			ret.Values = append(ret.Values, EnumValue{
				Idl: v,
				Go:  toCamelCase(v, true),
			})
		} else {
			t.failing(in, "unsupported literal: %T: %#V", v, v)
		}
	}
	return ret
}

func (t *Enum) GetAllTypeRefs(list []TypeRef) []TypeRef {
	return list
}
