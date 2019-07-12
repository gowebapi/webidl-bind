package types

import (
	"fmt"

	"github.com/gowebapi/webidlparser/ast"
)

type Interface struct {
	standardType
	basic BasicInfo

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
	Type  TypeRef
	Value string
}

type IfVar struct {
	nameAndLink
	Type        TypeRef
	Static      bool
	Readonly    bool
	Stringifier bool
}

type IfMethod struct {
	nameAndLink
	Return TypeRef
	Static bool
	Params []*Parameter
}

type TypeConvert func(in TypeRef) TypeRef

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
				Return: newInterfaceType(ret),
				Params: params,
			}
		} else if a.Name == "OnGlobalScope" {
			ret.Global = true
		} else if _, f := ignoredInterfaceAnnotation[a.Name]; !f {
			t.warning(ref, "unsupported interface annotation '%s'", a.Name)
		}
	}
	for _, c := range in.CustomOps {
		switch c.Name {
		case "stringifier":
			t.queueProtocolInterfaceStringifier(ret.basic.Idl, ret.ref)
		default:
			t.failing(ret.ref, "unsupported custom operation '%s'", c.Name)
		}
	}
	if in.Iterable != nil {
		v := in.Iterable.Elem
		ref := createRef(in.Iterable, t)
		if in.Iterable.Key == nil {
			t.queueProtocolIterableOne(ret.basic.Idl, v, ref)
		} else {
			k := in.Iterable.Key
			t.queueProtocolIterableTwo(ret.basic.Idl, k, v, ref)
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
		Type:  convertType(in.Type, conv),
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
	conv.warningTrue(in.Specialization == "" || in.Specialization == "stringifier", ref, "var: unsupported specialization")
	conv.assertTrue(in.Init == nil, ref, "var: unsupported default value")
	conv.assertTrue(!in.Required, ref, "var: unsupported required attribute")
	// parser.Dump(os.Stdout, in)

	return &IfVar{
		nameAndLink: nameAndLink{
			ref:  ref,
			name: fromIdlToMethodName(in.Name),
		},
		Type:        convertType(in.Type, conv),
		Static:      in.Static,
		Readonly:    in.Readonly,
		Stringifier: in.Specialization == "stringifier",
	}
}

func (conv *extractTypes) convertInterfaceMethod(in *ast.Member) *IfMethod {
	name := in.Name
	ref := createRef(in, conv)
	if name == "" {
		if in.Specialization == "" {
		}
		switch in.Specialization {
		case "":
			conv.failing(ref, "empty method name")
			return nil
		case "stringifier":
			name = "toString"
		default:
			conv.warning(ref, "skipping method, no support for specialization '%s'", in.Specialization)
			return nil
		}
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
			name: fromIdlToMethodName(name),
		},
		Return: convertType(in.Type, conv),
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

func (t *Interface) lessThan(b *Interface) bool {
	return t.basic.lessThan(&b.basic)
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
		info.Input = "*" + info.Input + "Value"
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
	for _, v := range m.Consts {
		t.Consts = append(t.Consts, v.copy())
	}
	t.Vars = append(t.Vars, m.Vars...)
	t.StaticVars = append(t.StaticVars, m.StaticVars...)
	t.Method = append(t.Method, m.Method...)
	t.StaticMethod = append(t.StaticMethod, m.StaticMethod...)
}

func (t *Interface) TypeID() TypeID {
	return TypeInterface
}

func (t *Interface) TemplateCopy(targetInfo BasicInfo) *Interface {
	src := t
	ref := *src.standardType.ref
	dst := &Interface{
		standardType: standardType{
			inuse:       true,
			needRelease: src.standardType.needRelease,
			ref:         &ref,
		},
		basic:        targetInfo,
		Inherits:     src.Inherits,
		inheritsName: src.inheritsName,
		Global:       src.Global,
		Callback:     src.Callback,
		FunctionCB:   src.FunctionCB,
		ConstPrefix:  src.ConstPrefix,
		ConstSuffix:  src.ConstSuffix,
		Constructor:  src.Constructor.copy(),
	}
	dst.basic.Template = src.basic.Template
	for _, in := range src.Consts {
		dst.Consts = append(dst.Consts, in.copy())
	}
	for _, in := range src.Vars {
		dst.Vars = append(dst.Vars, in.copy())
	}
	for _, in := range src.StaticVars {
		dst.StaticVars = append(dst.StaticVars, in.copy())
	}
	for _, in := range src.Method {
		dst.Method = append(dst.Method, in.copy())
	}
	for _, in := range src.StaticMethod {
		dst.StaticMethod = append(dst.StaticMethod, in.copy())
	}
	return dst
}

func (t *Interface) ChangeType(typeConv TypeConvert) {
	src := t
	if t.Constructor != nil {
		t.Constructor.changeType(typeConv)
	}
	for _, value := range src.Consts {
		value.Type = typeConv(value.Type)
	}
	for _, value := range src.Vars {
		value.Type = typeConv(value.Type)
	}
	for _, value := range src.StaticVars {
		value.Type = typeConv(value.Type)
	}
	for _, value := range src.Method {
		value.changeType(typeConv)
	}
	for _, value := range src.StaticMethod {
		value.changeType(typeConv)
	}
}

func (t *IfConst) copy() *IfConst {
	dup := *t
	return &dup
}

func (t *IfConst) SetType(value TypeRef) string {
	return "const can't change type"
}

func (t *IfVar) copy() *IfVar {
	r := *t.ref
	return &IfVar{
		nameAndLink: nameAndLink{
			name: t.nameAndLink.name,
			ref:  &r,
		},
		Type:     t.Type,
		Static:   t.Static,
		Readonly: t.Readonly,
	}
}

func (t *IfVar) SetType(value TypeRef) string {
	t.Type = value
	return ""
}

func (t *IfMethod) copy() *IfMethod {
	if t == nil {
		return t
	}
	r := *t.ref
	dst := &IfMethod{
		nameAndLink: nameAndLink{
			name: t.nameAndLink.name,
			ref:  &r,
		},
		Return: t.Return,
		Static: t.Static,
	}
	for _, pin := range t.Params {
		dst.Params = append(dst.Params, pin.copy())
	}
	return dst
}

func (t *IfMethod) changeType(typeConv TypeConvert) {
	t.Return = typeConv(t.Return)
	for i := range t.Params {
		t.Params[i].Type = typeConv(t.Params[i].Type)
	}
}

func (t *IfMethod) SetType(value TypeRef) string {
	return "method can't change type"
}
