package types

import (
	"fmt"

	"github.com/dennwc/webidl/ast"
)

type Interface struct {
	standardType
	source *ast.Interface
	basic  BasicInfo

	Inherits     *Interface
	inheritsName string

	Constructor *IfMethod

	Consts       []*IfConst
	Vars         []*IfVar
	StaticVars   []*IfVar
	Method       []*IfMethod
	StaticMethod []*IfMethod
}

// Interface need to implement Type
var _ Type = &Interface{}

type IfConst struct {
	nameAndLink
	name  MethodName
	Src   *ast.Member
	Type  TypeRef
	Value string
}

type IfVar struct {
	nameAndLink
	name     MethodName
	Src      *ast.Member
	Type     TypeRef
	Static   bool
	Readonly bool
}

type IfMethod struct {
	nameAndLink
	name   MethodName
	Src    *ast.Member
	SrcA   *ast.Annotation
	Return TypeRef
	Static bool
	Params []*Parameter
}

var ignoredInterfaceAnnotation = map[string]bool{
	"Exposed":                           true,
	"LegacyUnenumerableNamedProperties": true,
	"HTMLConstructor":                   true,
}

var ignoredMethodAnnotation = map[string]bool{
	"CEReactions": true, "NewObject": true,
	"Unscopable": true,
}

var ignoredVarAnnotations = map[string]bool{
	"Unforgeable": true, "Replaceable": true,
	"SameObject": true, "CEReactions": true,
	"PutForwards": true, "Unscopable": true,
}

func (t *extractTypes) convertInterface(in *ast.Interface) (*Interface, bool) {
	ret := &Interface{
		standardType: standardType{
			base:        in.NodeBase(),
			needRelease: false,
		},
		basic:        fromIdlToTypeName(t.main.setup.Package, in.Name, "interface"),
		source:       in,
		inheritsName: in.Inherits,
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
		if a.Name == "Constructor" {
			t.assertTrue(a.Value == "", a, "constructor shall have parameters, not A=B")
			t.assertTrue(len(a.Values) == 0, a, "constructor shall have parameters, not A=(a,b,c)")
			params := t.convertParams(a.Parameters)
			ret.Constructor = &IfMethod{
				nameAndLink: nameAndLink{
					base: a.NodeBase(),
				},
				name:   fromIdlToMethodName("New_" + ret.basic.Idl),
				Static: true,
				SrcA:   a,
				Return: newInterfaceType(ret),
				Params: params,
			}
		} else if _, f := ignoredInterfaceAnnotation[a.Name]; !f {
			t.warning(a, "unsupported interface annotation '%s'", a.Name)
		}
	}
	return ret, in.Partial
}

func (conv *extractTypes) convertInterfaceConst(in *ast.Member) *IfConst {
	conv.assertTrue(len(in.Annotations) == 0, in, "const: unsupported annotation")
	conv.assertTrue(len(in.Parameters) == 0, in, "const: unsupported parameters")
	conv.assertTrue(in.Specialization == "", in, "const: unsupported specialization")
	conv.assertTrue(in.Init != nil, in, "const: missing default value")

	value := ""
	if basic, ok := in.Init.(*ast.BasicLiteral); ok {
		value = basic.Value
	} else {
		conv.failing(in, "const: unsupported default value")
	}
	return &IfConst{
		nameAndLink: nameAndLink{
			base: in.NodeBase(),
		},
		name:  fromIdlToMethodName(in.Name),
		Src:   in,
		Type:  convertType(in.Type),
		Value: value,
	}
}

func (conv *extractTypes) convertInterfaceVar(in *ast.Member) *IfVar {
	for _, a := range in.Annotations {
		if _, f := ignoredVarAnnotations[a.Name]; f {
			continue
		}
		switch a.Name {
		case "TreatNullAs":
			// only for DOMString
			if a.Value != "EmptyString" {
				conv.failing(in, "for TreatNullAs, only TreatNullAs=EmptyString is allowed according to specification")
			}
			conv.warning(in, "unhandled TreatNullAs (null should be an empty string)")
		default:
			conv.warning(a, "unhandled variable annotation '%s'", a.Name)
		}
	}
	conv.assertTrue(len(in.Parameters) == 0, in, "var: unsupported parameters")
	conv.warningTrue(in.Specialization == "", in, "var: unsupported specialization")
	conv.assertTrue(in.Init == nil, in, "var: unsupported default value")
	conv.assertTrue(!in.Required, in, "var: unsupported required attribute")
	// parser.Dump(os.Stdout, in)

	return &IfVar{
		nameAndLink: nameAndLink{
			base: in.NodeBase(),
		},
		name:     fromIdlToMethodName(in.Name),
		Src:      in,
		Type:     convertType(in.Type),
		Static:   in.Static,
		Readonly: in.Readonly,
	}
}

func (conv *extractTypes) convertInterfaceMethod(in *ast.Member) *IfMethod {
	if in.Name == "" {
		if in.Specialization != "" {
			conv.warning(in, "skipping method, no support for specialization '%s'", in.Specialization)
		} else {
			conv.failing(in, "empty method name")
		}
		return nil
	}
	conv.warningTrue(in.Specialization == "", in, "method: unsupported specialization (need to be implemented)")
	conv.assertTrue(in.Init == nil, in, "method: unsupported default value")
	conv.assertTrue(!in.Required, in, "method: unsupported required tag")
	// TODO add support for method annotations
	for _, a := range in.Annotations {
		if _, f := ignoredMethodAnnotation[a.Name]; f {
			continue
		}
		conv.warning(a, "unsupported method annotation '%s'", a.Name)
	}

	return &IfMethod{
		nameAndLink: nameAndLink{
			base: in.NodeBase(),
		},
		name:   fromIdlToMethodName(in.Name),
		Src:    in,
		Return: convertType(in.Type),
		Static: in.Static,
		Params: conv.convertParams(in.Parameters),
	}
}

func (t *Interface) Basic() BasicInfo {
	return t.basic
}

func (t *Interface) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *Interface) key() string {
	return t.basic.Idl
}

func (t *Interface) link(conv *Convert, inuse inuseLogic) TypeRef {
	if t.inuse {
		return t
	}
	t.inuse = true

	if t.inheritsName != "" {
		if parent, ok := conv.Types[t.inheritsName]; ok {
			if ip, ok := parent.(*Interface); ok {
				t.Inherits = ip
			} else {
				conv.failing(t, "inherits '%s' that is not an interface but a %T", t.inheritsName, parent)
			}
		} else {
			conv.failing(t, "inherits unknown '%s'", t.inheritsName)
		}
	}

	if t.Constructor != nil {
		t.Constructor.Return = t.Constructor.Return.link(conv, make(inuseLogic))
		for _, p := range t.Constructor.Params {
			p.Type = p.Type.link(conv, make(inuseLogic))
		}
	}
	for _, m := range t.Consts {
		m.Type = m.Type.link(conv, make(inuseLogic))
	}
	for _, m := range t.Vars {
		m.Type = m.Type.link(conv, make(inuseLogic))
	}
	for _, m := range t.StaticVars {
		m.Type = m.Type.link(conv, make(inuseLogic))
	}
	for _, m := range t.Method {
		m.Return = m.Return.link(conv, make(inuseLogic))
		for _, p := range m.Params {
			p.Type = p.Type.link(conv, make(inuseLogic))
		}
	}
	for _, m := range t.StaticMethod {
		m.Return = m.Return.link(conv, make(inuseLogic))
		for _, p := range m.Params {
			p.Type = p.Type.link(conv, make(inuseLogic))
		}
	}
	return t
}

func (t *Interface) Param(nullable, option, vardict bool) (info *TypeInfo, inner TypeRef) {
	return newTypeInfo(t.basic, nullable, option, vardict, false, false, false), t
}

func (t *Interface) merge(m *Interface, conv *Convert) {
	t.Consts = append(t.Consts, m.Consts...)
	t.Vars = append(t.Vars, m.Vars...)
	t.StaticVars = append(t.StaticVars, m.StaticVars...)
	t.Method = append(t.Method, m.Method...)
	t.StaticMethod = append(t.StaticMethod, m.StaticMethod...)
}

func (t *Interface) mergeMixin(m *mixin, conv *Convert) {
	t.Consts = append(t.Consts, m.Consts...)
	t.Vars = append(t.Vars, m.Vars...)
	t.StaticVars = append(t.StaticVars, m.StaticVars...)
	t.Method = append(t.Method, m.Method...)
	t.StaticMethod = append(t.StaticMethod, m.StaticMethod...)
}

func (t *IfConst) Name() MethodName {
	return t.name
}

func (t *IfVar) Name() MethodName {
	return t.name
}

func (t *IfMethod) Name() MethodName {
	return t.name
}
