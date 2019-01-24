package types

import (
	"github.com/dennwc/webidl/ast"
)

// Enum type
type Enum struct {
	name   Name
	source *ast.Enum
	Values []EnumValue
}

// EnumValue is a single enum value
type EnumValue struct {
	Idl string
	Go  string
}

func ConvertEnum(in *ast.Enum, setup *Setup) (*Enum, error) {
	if len(in.Annotations) > 0 {
		return nil, UnsupportedAnnotationErr
	}
	ret := &Enum{
		source: in,
		name:   fromIdlName(setup.Package, in.Name),
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
			return nil, UnsupportedLiteralErr
		}
	}

	return ret, nil
}

func (t *Enum) Base() *ast.Base {
	return t.source.NodeBase()
}

func (t *Enum) Name() Name {
	return t.name
}

func (t *Enum) GetAllTypeRefs(list []TypeRef) []TypeRef {
	return list
}
