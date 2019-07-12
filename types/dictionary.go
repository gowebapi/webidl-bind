package types

import (
	"github.com/gowebapi/webidlparser/ast"
)

type Dictionary struct {
	standardType
	basic        BasicInfo
	Inherits     *Dictionary
	inheritsName string

	Members []*DictMember
}

// Dictionary need to implement Type
var _ Type = &Dictionary{}

type DictMember struct {
	nameAndLink
	Type     TypeRef
	Required bool
}

func (t *extractTypes) convertDictionary(in *ast.Dictionary) (*Dictionary, bool) {
	ref := createRef(in, t)
	t.warningTrue(len(in.Annotations) == 0, ref, "unsupported annotations")
	// t.assertTrue(in.Inherits == "", ref , "unsupported dictionary inherites of %s", in.Inherits)
	ret := &Dictionary{
		standardType: standardType{
			ref:         ref,
			needRelease: false,
		},
		basic:        fromIdlToTypeName(t.main.setup.Package, in.Name, "dictionary"),
		inheritsName: in.Inherits,
	}
	for _, mi := range in.Members {
		mo := t.convertDictMember(mi)
		ret.Members = append(ret.Members, mo)
	}
	return ret, in.Partial
}

func (conv *extractTypes) convertDictMember(in *ast.Member) *DictMember {
	ref := createRef(in, conv)
	conv.assertTrue(!in.Readonly, ref, "read only not allowed")
	conv.assertTrue(in.Attribute, ref, "must be an attribute")
	conv.assertTrue(!in.Static, ref, "static is not allowed")
	conv.assertTrue(!in.Const, ref, "const is not allowed")
	conv.assertTrue(len(in.Parameters) == 0, ref, "parameters on member is not allowed (or not supported)")
	conv.assertTrue(len(in.Specialization) == 0, ref, "specialization on member is not allowed (or not supported)")
	conv.warningTrue(!in.Required, ref, "required value not implemented yet, report this as a bug :)")
	for _, a := range in.Annotations {
		ref := createRef(a, conv)
		conv.warning(ref, "dictionary member: annotation '%s' is not supported", a.Name)
	}
	if in.Init != nil {
		conv.warning(ref, "dictionary: default value for dictionary not implemented yet")
		// parser.Dump(os.Stdout, in)
	}
	return &DictMember{
		nameAndLink: nameAndLink{
			ref:  createRef(in, conv),
			name: fromIdlToMethodName(in.Name),
		},
		Type:     convertType(in.Type, conv),
		Required: in.Required,
	}
}

func (t *Dictionary) Basic() BasicInfo {
	return TransformBasic(t, t.basic)
}

func (t *Dictionary) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *Dictionary) key() string {
	return t.basic.Idl
}

func (t *Dictionary) lessThan(b *Dictionary) bool {
	return t.basic.lessThan(&b.basic)
}

func (t *Dictionary) link(conv *Convert, inuse inuseLogic) TypeRef {
	if t.inuse {
		return t
	}
	t.inuse = true
	if t.inheritsName != "" {
		if candidate, ok := conv.Types[t.inheritsName]; ok {
			inner := make(inuseLogic)
			typeRef := candidate.link(conv, inner)
			if parent, ok := typeRef.(*Dictionary); ok {
				// found parent that is a dictionary
				// copy members
				out := make([]*DictMember, 0)
				for _, in := range parent.Members {
					out = append(out, in.copy())
				}
				t.Members = append(out, t.Members...)
			} else {
				conv.failing(t, "expected inherit to be a dictionary, not %T", typeRef)
			}
		} else {
			conv.failing(t, "unable to find inherit '%s'", t.inheritsName)
		}
	}
	inner := make(inuseLogic)
	for _, m := range t.Members {
		m.Type = m.Type.link(conv, inner)
	}
	return t
}

func (t *Dictionary) merge(partial *Dictionary, conv *Convert) {
	conv.assertTrue(partial.inheritsName == "", partial, "unsupported dictionary inherites on partial")
	// TODO member elemination logic with duplicate is detected
	t.Members = append(t.Members, partial.Members...)
}

func (t *Dictionary) NeedRelease() bool {
	need := false
	for _, v := range t.Members {
		need = need || v.Type.NeedRelease()
	}
	return need
}

func (t *Dictionary) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	return newTypeInfo(t.Basic(), nullable, option, variadic, true, false, false), t
}

func (t *Dictionary) SetBasic(basic BasicInfo) {
	t.basic = basic
}

func (t *Dictionary) TypeID() TypeID {
	return TypeDictionary
}

func (t *Dictionary) templateCopy(targetInfo BasicInfo) *Dictionary {
	src := t
	ref := *src.standardType.ref
	dst := &Dictionary{
		standardType: standardType{
			inuse:       true,
			needRelease: src.standardType.needRelease,
			ref:         &ref,
		},
		basic: targetInfo,

		Inherits:     src.Inherits,
		inheritsName: src.inheritsName,
	}
	dst.basic.Template = src.basic.Template
	for _, m := range src.Members {
		dst.Members = append(dst.Members, m.copy())
	}
	return dst
}

func (t *Dictionary) changeType(typeConv TypeConvert) {
	for _, m := range t.Members {
		m.Type = typeConv(m.Type)
	}
}

func (t *DictMember) copy() *DictMember {
	// TODO does the type need to be deep copied?
	return &DictMember{
		nameAndLink: t.nameAndLink,
		Type:        t.Type,
		Required:    t.Required,
	}
}

func (t *DictMember) SetType(value TypeRef) string {
	t.Type = value
	return ""
}
