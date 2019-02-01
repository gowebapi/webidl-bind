package types

import (
	"github.com/dennwc/webidl/ast"
)

type typeDef struct {
	standardType
	basic BasicInfo
	Type  TypeRef
}

var _ Type = &typeDef{}

func (t *extractTypes) convertTypeDef(in *ast.Typedef) *typeDef {
	ret := typeDef{
		standardType: standardType{
			base:        in.NodeBase(),
			needRelease: false,
		},
		basic: fromIdlToTypeName("", in.Name, "typedef"),
		Type:  convertType(in.Type),
	}
	for _, a := range in.Annotations {
		t.warning(a, "typedef: unsupported annotation '%s'", a.Name)
	}
	return &ret
}

func (t *typeDef) Basic() BasicInfo {
	return t.basic
}

func (t *typeDef) DefaultParam() *TypeInfo {
	return t.Param(false, false, false)
}

func (t *typeDef) GetAllTypeRefs(list []TypeRef) []TypeRef {
	list = append(list, t.Type)
	return list
}

func (t *typeDef) key() string {
	return t.basic.Idl
}

func (t *typeDef) Param(nullable, option, vardict bool) *TypeInfo {
	return t.Type.Param(nullable, option, vardict)
}
