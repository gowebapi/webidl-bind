package types

import (
	"github.com/dennwc/webidl/ast"
)

type Dictionary struct {
	standardType
	source   *ast.Dictionary
	Inherits *Dictionary

	Members []*DictMember
}

// Dictionary need to implement Type
var _ Type = &Dictionary{}

type DictMember struct {
	nameAndLink
	Src      *ast.Member
	Type     TypeRef
	Required bool
}

func (t *extractTypes) convertDictionary(in *ast.Dictionary) (*Dictionary, bool) {
	t.assertTrue(len(in.Annotations) == 0, in, "unsupported annotations")
	ret := &Dictionary{
		standardType: standardType{
			base:        in.NodeBase(),
			name:        fromIdlName(t.main.setup.Package, in.Name, true),
			needRelease: false,
		},
		source: in,
	}
	for _, mi := range in.Members {
		mo := t.convertDictMember(mi)
		ret.Members = append(ret.Members, mo)
	}
	return ret, in.Partial
}

func (conv *extractTypes) convertDictMember(in *ast.Member) *DictMember {
	conv.assertTrue(!in.Readonly, in, "read only not allowed")
	conv.assertTrue(in.Attribute, in, "must be an attribute")
	conv.assertTrue(!in.Static, in, "static is not allowed")
	conv.assertTrue(!in.Const, in, "const is not allowed")
	conv.assertTrue(len(in.Annotations) == 0, in, "annotations are not supported")
	conv.assertTrue(len(in.Parameters) == 0, in, "parameters on member is not allowed (or not supported)")
	conv.assertTrue(len(in.Specialization) == 0, in, "specialization on member is not allowed (or not supported)")
	conv.assertTrue(!in.Required, in, "default value not implemented yet, report this as a bug :)")
	if in.Init != nil {
		conv.warning(in, "default value for dictionary not implemented yet")
	}
	return &DictMember{
		nameAndLink: nameAndLink{
			base: in.NodeBase(),
			name: fromMethodName(in.Name),
		},
		Src:      in,
		Type:     convertType(in.Type),
		Required: in.Required,
	}
}

func (t *Dictionary) GetAllTypeRefs(list []TypeRef) []TypeRef {
	for _, m := range t.Members {
		list = append(list, m.Type)
	}
	return list
}

func (t *Dictionary) NeedRelease() bool {
	need := false
	for _, v := range t.Members {
		need = need || v.Type.NeedRelease()
	}
	return need
}

func (t *Dictionary) TemplateName() (string, TemplateNameFlags) {
	return "dictionary", PointerTnFlag
}
