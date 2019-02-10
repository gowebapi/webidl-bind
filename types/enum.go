package types

import (
	"fmt"

	"github.com/gowebapi/webidlparser/ast"
)

// Enum type
type Enum struct {
	standardType
	basic  BasicInfo
	source *ast.Enum
	Values []EnumValue

	// target language prefix and suffix for enum values
	Prefix, Suffix string
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
	ret.Suffix = ret.basic.Def

	for i, v := range in.Values {
		if b, ok := v.(*ast.BasicLiteral); ok {
			idl := b.Value
			idl = clipString(idl)
			lang := idl
			if lang == "" {
				lang = fmt.Sprintf("empty_string_%d", i)
			}
			ret.Values = append(ret.Values, EnumValue{
				Idl: idl,
				Go:  fixLangName(toCamelCase(lang, true)),
			})
		} else {
			t.failing(in, "unsupported literal: %T: %#V", v, v)
		}
	}
	return ret
}

func (t *Enum) Basic() BasicInfo {
	return TransformBasic(t, t.basic)
}

func (t *Enum) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *Enum) key() string {
	return t.basic.Idl
}

func (t *Enum) link(conv *Convert, inuse inuseLogic) TypeRef {
	t.inuse = true
	return t
}

func (t *Enum) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	return newTypeInfo(t.Basic(), nullable, option, variadic, false, false, false), t
}

func (t *Enum) SetBasic(basic BasicInfo) {
	t.basic = basic
}
