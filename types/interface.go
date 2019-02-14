package types

import (
	"fmt"

	"github.com/gowebapi/webidlparser/ast"
)

type Interface struct {
	standardType
	source *ast.Interface
	basic  BasicInfo

	Inherits     *Interface
	inheritsName string

	// Global indicate that this "interface" is actually the
	// global scope of javascript
	Global bool

	// Callback specify that this is an interface that the
	// developer should implement.
	Callback bool

	// FunctionCB indicate that an extra function implementation
	// shall be added in final output
	FunctionCB bool

	// variable naming prefix and suffix for const variables
	ConstPrefix, ConstSuffix string

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
	Src   *ast.Member
	Type  TypeRef
	Value string
}

type IfVar struct {
	nameAndLink
	Src      *ast.Member
	Type     TypeRef
	Static   bool
	Readonly bool
}

type IfMethod struct {
	nameAndLink
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
			ref:         createRef(in, t),
			needRelease: false,
		},
		basic:        fromIdlToTypeName(t.main.setup.Package, in.Name, "interface"),
		source:       in,
		inheritsName: in.Inherits,
		Callback:     in.Callback,
		FunctionCB:   true,
	}
	ret.ConstSuffix = "_" + ret.basic.Def
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
		ref := createRef(a, t)
		if a.Name == "Constructor" {
			t.assertTrue(a.Value == "", ref, "constructor shall have parameters, not A=B")
			t.assertTrue(len(a.Values) == 0, ref, "constructor shall have parameters, not A=(a,b,c)")
			params := t.convertParams(a.Parameters)
			ret.Constructor = &IfMethod{
				nameAndLink: nameAndLink{
					ref:  ref,
					name: fromIdlToMethodName("New_" + ret.basic.Idl),
				},
				Static: true,
				SrcA:   a,
				Return: newInterfaceType(ret),
				Params: params,
			}
		} else if a.Name == "OnGlobalScope" {
			ret.Global = true
		} else if _, f := ignoredInterfaceAnnotation[a.Name]; !f {
			t.warning(ref, "unsupported interface annotation '%s'", a.Name)
		}
	}
	return ret, in.Partial
}

func (conv *extractTypes) convertInterfaceConst(in *ast.Member) *IfConst {
	ref := createRef(in, conv)
	conv.assertTrue(len(in.Annotations) == 0, ref, "const: unsupported annotation")
	conv.assertTrue(len(in.Parameters) == 0, ref, "const: unsupported parameters")
	conv.assertTrue(in.Specialization == "", ref, "const: unsupported specialization")
	conv.assertTrue(in.Init != nil, ref, "const: missing default value")

	value := ""
	if basic, ok := in.Init.(*ast.BasicLiteral); ok {
		value = basic.Value
	} else {
		conv.failing(ref, "const: unsupported default value")
	}
	return &IfConst{
		nameAndLink: nameAndLink{
			ref:  ref,
			name: fromIdlToMethodName(in.Name),
		},
		Src:   in,
		Type:  convertType(in.Type,conv ),
		Value: value,
	}
}

func (conv *extractTypes) convertInterfaceVar(in *ast.Member) *IfVar {
	ref := createRef(in, conv)
	for _, a := range in.Annotations {
		if _, f := ignoredVarAnnotations[a.Name]; f {
			continue
		}
		switch a.Name {
		case "TreatNullAs":
			// only for DOMString
			if a.Value != "EmptyString" {
				conv.failing(ref, "for TreatNullAs, only TreatNullAs=EmptyString is allowed according to specification")
			}
			conv.warning(ref, "unhandled TreatNullAs (null should be an empty string)")
		default:
			ref = createRef(a, conv)
			conv.warning(ref, "unhandled variable annotation '%s'", a.Name)
		}
	}
	conv.assertTrue(len(in.Parameters) == 0, ref, "var: unsupported parameters")
	conv.warningTrue(in.Specialization == "", ref, "var: unsupported specialization")
	conv.assertTrue(in.Init == nil, ref, "var: unsupported default value")
	conv.assertTrue(!in.Required, ref, "var: unsupported required attribute")
	// parser.Dump(os.Stdout, in)

	return &IfVar{
		nameAndLink: nameAndLink{
			ref:  ref,
			name: fromIdlToMethodName(in.Name),
		},
		Src:      in,
		Type:     convertType(in.Type,conv ),
		Static:   in.Static,
		Readonly: in.Readonly,
	}
}

func (conv *extractTypes) convertInterfaceMethod(in *ast.Member) *IfMethod {
	ref := createRef(in, conv)
	if in.Name == "" {
		if in.Specialization != "" {
			conv.warning(ref, "skipping method, no support for specialization '%s'", in.Specialization)
		} else {
			conv.failing(ref, "empty method name")
		}
		return nil
	}
	conv.warningTrue(in.Specialization == "", ref, "method: unsupported specialization (need to be implemented)")
	conv.assertTrue(in.Init == nil, ref, "method: unsupported default value")
	conv.assertTrue(!in.Required, ref, "method: unsupported required tag")
	// TODO add support for method annotations
	for _, a := range in.Annotations {
		if _, f := ignoredMethodAnnotation[a.Name]; f {
			continue
		}
		aref := createRef(a, conv)
		conv.warning(aref, "unsupported method annotation '%s'", a.Name)
	}

	return &IfMethod{
		nameAndLink: nameAndLink{
			ref:  ref,
			name: fromIdlToMethodName(in.Name),
		},
		Src:    in,
		Return: convertType(in.Type,conv ),
		Static: in.Static,
		Params: conv.convertParams(in.Parameters),
	}
}

func (t *Interface) Basic() BasicInfo {
	return TransformBasic(t, t.basic)
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

func (t *Interface) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	info = newTypeInfo(t.Basic(), nullable, option, variadic, true, t.Callback, false)
	if t.Callback {
		info.Input = "*" + info.InOut + "Value"
	}
	return info, t
}

func (t *Interface) SetBasic(value BasicInfo) {
	t.basic = value
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


func (t *Interface) TypeID() TypeID {
	return TypeInterface
}
