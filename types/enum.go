package types

import (
	"github.com/dennwc/webidl/ast"
)

// Enum type
type Enum struct {
	standardType
	basic  BasicInfo
	source *ast.Enum
	Values []EnumValue
}

// Enum need to implement Type
var _ Type = &Enum{}

// EnumValue is a single enum value
type EnumValue struct {
	Idl string
	Go  string
}

func (t *extractTypes) convertEnum(in *ast.Enum) *Enum {
	t.assertTrue(len(in.Annotations) == 0, in, "unsupported annotation")
	ret := &Enum{
		standardType: standardType{
			base:        in.NodeBase(),
			needRelease: false,
		},
		basic:  fromIdlToTypeName(t.main.setup.Package, in.Name, "enum"),
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

func (t *Enum) Basic() BasicInfo {
	return t.basic
}

func (t *Enum) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *Enum) key() string {
	return t.basic.Idl
}

func (t *Enum) link(conv *Convert, inuse inuseLogic) TypeRef {
	t.inuse = true
	return t
}

func (t *Enum) Param(nullable, option, vardict bool) *TypeInfo {
	return newTypeInfo(t.basic, nullable, option, vardict, false, false, false)
}
