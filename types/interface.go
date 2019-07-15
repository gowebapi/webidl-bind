package types

import (
	"fmt"

	"github.com/gowebapi/webidlparser/ast"
)

// TODO: A maplike interface and its inherited interfaces must
// not have an iterable declaration, a setlike declaration, or
// an indexed property getter.

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

	Consts         []*IfConst
	Vars           []*IfVar
	StaticVars     []*IfVar
	Method         []*IfMethod
	StaticMethod   []*IfMethod
	Specialization []*IfMethod

	// indicate that this interface have replacable methods
	haveReplacableMethods bool

	// SpecProperty is used by transform step to assign names
	// for getters, setters and deleters
	SpecProperty map[SpecializationType]string
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

	// ReplaceOnOverride indicate that this method can be replaced
	// by another method and shall not be overrided.
	replaceOnOverride bool

	// Specialization indicate if this is a getter, setter or deleter
	Specialization SpecializationType
}

type TypeConvert func(in TypeRef) TypeRef

type SpecializationType int

const (
	// Nothing assigned
	SpecNone SpecializationType = iota
	// getter with integer index
	SpecIndexGetter
	SpecIndexSetter
	SpecKeyGetter
	SpecKeySetter
	SpecKeyDeleter
)

var ignoredInterfaceAnnotation = map[string]bool{
	"Exposed":                           true,
	"LegacyUnenumerableNamedProperties": true,
	"HTMLConstructor":                   true,
}

var ignoredMethodAnnotation = map[string]bool{
	"CEReactions": true, "NewObject": true, "Unscopable": true,
	"SecureContext": true, "Exposed": true, "SameObject": true,
	"Default": true, "Unforgeable": true,
	"WebGLHandlesContextLoss": true,
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
		SpecProperty: make(map[SpecializationType]string),
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
			mo, spec := t.convertInterfaceMethod(mi)
			if mo != nil {
				ret.StaticMethod = append(ret.StaticMethod, mo)
			}
			if spec != nil {
				t.failing(spec.ref, "specialization is not supported for static methods")
			}
		} else {
			mo, spec := t.convertInterfaceMethod(mi)
			if mo != nil {
				ret.haveReplacableMethods = ret.haveReplacableMethods || mo.replaceOnOverride
				ret.Method = append(ret.Method, mo)
			}
			if spec != nil {
				ret.Specialization = append(ret.Specialization, spec)
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
	for idx, pattern := range in.Patterns {
		ref := createRef(pattern, t)
		if idx >= 1 {
			t.failing(ref, "an interface may only have one of iterable, maplike or setlike.")
			break
		}
		switch pattern.Type {
		case ast.Iterable:
			v := pattern.Elem
			if pattern.Key == nil {
				t.queueProtocolIterableOne(ret.basic.Idl, v, ref)
			} else {
				k := pattern.Key
				t.queueProtocolIterableTwo(ret.basic.Idl, k, v, ref)
			}
		case ast.AsyncIterable:
			t.failing(createRef(pattern, t), "async iterable is not implemented")
		case ast.Maplike:
			t.queueProtocolMaplike(ret.basic.Idl, pattern.ReadOnly, pattern.Key, pattern.Elem, ref)
		case ast.Setlike:
			t.queueProtocolSetlike(ret.basic.Idl, pattern.ReadOnly, pattern.Elem, ref)
		default:
			panic(fmt.Sprint("unknown pattern: ", pattern.Type))
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

func (conv *extractTypes) convertInterfaceMethod(in *ast.Member) (*IfMethod, *IfMethod) {
	ref := createRef(in, conv)
	conv.assertTrue(in.Init == nil, ref, "method: unsupported default value")
	conv.assertTrue(!in.Required, ref, "method: unsupported required tag")

	name := in.Name
	value := &IfMethod{
		nameAndLink: nameAndLink{
			ref: ref,
			// name: assigned below
		},
		Return: convertType(in.Type, conv),
		Static: in.Static,
		Params: conv.convertParams(in.Parameters),
	}

	// process annotations
	// TODO add support for method annotations
	for _, a := range in.Annotations {
		if a.Name == "ReplaceOnOverride" {
			value.replaceOnOverride = true
		} else if _, f := ignoredMethodAnnotation[a.Name]; !f {
			aref := createRef(a, conv)
			conv.warning(aref, "unsupported method annotation '%s'", a.Name)
		}
	}

	// process specialization
	method := value
	var specialization *IfMethod
	switch in.Specialization {
	case "":
		if name == "" {
			conv.failing(ref, "empty method name")
			return nil, nil
		}
	case "stringifier":
		if name == "" {
			name = "toString"
		}
	case "getter", "setter", "deleter":
		specialization = value
		if name == "" {
			method = nil
		}
		spec, msg := value.identifySpecializationType(in.Specialization)
		if msg != "" {
			conv.failing(ref, msg)
		}
		value.Specialization = spec
	default:
		conv.warning(ref, "skipping method, no support for specialization '%s'", in.Specialization)
		return nil, nil
	}
	value.name = fromIdlToMethodName(name)
	return method, specialization
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
	for _, m := range t.Specialization {
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
	t.Consts = mergeConstants(t.Consts, m.Consts)
	t.Vars = mergeVariables(t.Vars, m.Vars)
	t.StaticVars = mergeVariables(t.StaticVars, m.StaticVars)
	t.Method = mergeMethods(t.Method, m.Method)
	t.StaticMethod = mergeMethods(t.StaticMethod, m.StaticMethod)
	t.Specialization = mergeMethods(t.Specialization, m.Specialization)
	t.haveReplacableMethods = t.haveReplacableMethods || m.haveReplacableMethods
}

func (t *Interface) mergeMixin(m *mixin, conv *Convert) {
	t.Consts = mergeConstants(t.Consts, m.Consts)
	t.Vars = mergeVariables(t.Vars, m.Vars)
	t.StaticVars = mergeVariables(t.StaticVars, m.StaticVars)
	t.Method = mergeMethods(t.Method, m.Method)
	t.StaticMethod = mergeMethods(t.StaticMethod, m.StaticMethod)
	t.Specialization = mergeMethods(t.Specialization, m.Specialization)
	t.haveReplacableMethods = t.haveReplacableMethods || m.haveReplacableMethods
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
		Constructor:  src.Constructor.Copy(),
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
		dst.Method = append(dst.Method, in.Copy())
	}
	for _, in := range src.StaticMethod {
		dst.StaticMethod = append(dst.StaticMethod, in.Copy())
	}
	for _, in := range src.Specialization {
		dst.Specialization = append(dst.Specialization, in.Copy())
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
	for _, value := range src.Specialization {
		value.changeType(typeConv)
	}
}

func cleanupReplaceMethods(list []*IfMethod) []*IfMethod {
	// first we search unique names
	countNames := make(map[string]int)
	for _, m := range list {
		key := m.name.Idl
		countNames[key] = countNames[key] + 1
	}
	// all replacable that isn't unique should be removed
	out := make([]*IfMethod, 0, len(list))
	for _, m := range list {
		key := m.name.Idl
		count := countNames[key]
		if m.replaceOnOverride && count > 1 {
			continue
		}
		out = append(out, m)
	}
	return out
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
		Type:        t.Type,
		Static:      t.Static,
		Readonly:    t.Readonly,
		Stringifier: t.Stringifier,
	}
}

func (t *IfVar) SetType(value TypeRef) string {
	t.Type = value
	return ""
}

func (t *IfMethod) Copy() *IfMethod {
	if t == nil {
		return t
	}
	r := *t.ref
	dst := &IfMethod{
		nameAndLink: nameAndLink{
			name: t.nameAndLink.name,
			ref:  &r,
		},
		Return:            t.Return,
		Static:            t.Static,
		replaceOnOverride: t.replaceOnOverride,
		Specialization:    t.Specialization,
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

func (t *IfMethod) identifySpecializationType(forWhat string) (SpecializationType, string) {
	for _, p := range t.Params {
		if p.Variadic || p.Optional {
			return SpecNone, "special operands can't have optional or variadic arguments"
		}
	}
	if forWhat == "getter" {
		if len(t.Params) != 1 {
			return SpecNone, "getter must have exact one argument"
		}
		if isUnsignedInt(t.Params[0].Type) {
			return SpecIndexGetter, ""
		} else if IsString(t.Params[0].Type) {
			return SpecKeyGetter, ""
		}
		return SpecNone, "unknown type for getter argument. only unsigned integer or DOMString is supported"
	} else if forWhat == "setter" {
		if len(t.Params) != 2 {
			return SpecNone, "setter must have exactly two arguments"
		}
		if isUnsignedInt(t.Params[0].Type) {
			return SpecIndexSetter, ""
		} else if IsString(t.Params[0].Type) {
			return SpecKeySetter, ""
		}
		return SpecNone, "unknown type for setter argument, only unsigned integer or DOMString is supported"
	} else if forWhat == "deleter" {
		if len(t.Params) != 1 {
			return SpecNone, "deleter must have exact one argument"
		}
		if IsString(t.Params[0].Type) {
			return SpecKeyDeleter, ""
		}
		return SpecNone, "unknown type for deleter argument. only DOMString is supported"
	}
	panic(forWhat)
	// return SpecNone, ""
}

func (t *IfMethod) SetType(value TypeRef) string {
	return "method can't change type"
}

func mergeConstants(dst, src []*IfConst) []*IfConst {
	for _, v := range src {
		dst = append(dst, v.copy())
	}
	return dst
}

func mergeVariables(dst, src []*IfVar) []*IfVar {
	for _, v := range src {
		dst = append(dst, v.copy())
	}
	return dst
}

func mergeMethods(dst, src []*IfMethod) []*IfMethod {
	for _, v := range src {
		dst = append(dst, v.Copy())
	}
	return dst
}
