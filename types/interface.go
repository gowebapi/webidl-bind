package types

import (
	"fmt"

	"github.com/dennwc/webidl/ast"
)

type Interface struct {
	standardType
	source   *ast.Interface
	Inherits *Interface

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
	"HTMLConstructor": true, 
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
			name:        fromIdlName(t.main.setup.Package, in.Name, false),
			needRelease: false,
		},
		source: in,
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
			ret.StaticMethod = append(ret.StaticMethod, mo)
		} else {
			mo := t.convertInterfaceMethod(mi)
			ret.Method = append(ret.Method, mo)
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
					name: fromMethodName("New_" + ret.Name().Idl),
				},
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
			name: fromMethodName(in.Name),
		},
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
			name: fromMethodName(in.Name),
		},
		Src:      in,
		Type:     convertType(in.Type),
		Static:   in.Static,
		Readonly: in.Readonly,
	}
}

func (conv *extractTypes) convertInterfaceMethod(in *ast.Member) *IfMethod {
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
			name: fromMethodName(in.Name),
		},
		Src:    in,
		Return: convertType(in.Type),
		Static: in.Static,
		Params: conv.convertParams(in.Parameters),
	}
}

func (t *Interface) GetAllTypeRefs(list []TypeRef) []TypeRef {
	if t.Constructor != nil {
		list = append(list, t.Constructor.Return)
		for _, p := range t.Constructor.Params {
			list = append(list, p.Type)
		}
	}
	for _, m := range t.Consts {
		list = append(list, m.Type)
	}
	for _, m := range t.Vars {
		list = append(list, m.Type)
	}
	for _, m := range t.StaticVars {
		list = append(list, m.Type)
	}
	for _, m := range t.Method {
		list = append(list, m.Return)
		for _, p := range m.Params {
			list = append(list, p.Type)
		}
	}
	for _, m := range t.StaticMethod {
		list = append(list, m.Return)
		for _, p := range m.Params {
			list = append(list, p.Type)
		}
	}
	return list
}

func (t *Interface) TemplateName() (string, TemplateNameFlags) {
	return "interface", NoTnFlag
}
