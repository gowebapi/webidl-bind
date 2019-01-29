package types

import (
	"github.com/dennwc/webidl/ast"
)

type typeDef struct {
	standardType
	Type TypeRef
}

var _ Type = &typeDef{}

func (t *extractTypes) convertTypeDef(in *ast.Typedef) *typeDef {
	ret := typeDef{
		standardType: standardType{
			base:        in.NodeBase(),
			name:        fromIdlName("", in.Name, false),
			needRelease: false,
		},
		Type: convertType(in.Type),
	}
	for _, a := range in.Annotations {
		t.warning(a, "typedef: unsupported annotation '%s'", a.Name)
	}
	return &ret
}

func (t *typeDef) GetAllTypeRefs(list []TypeRef) []TypeRef {
	list = append(list, t.Type)
	return list
}

func (t *typeDef) TemplateName() (string, TemplateNameFlags) {
	panic("typeDef should be expanded before generation step")
}
