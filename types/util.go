// contains types converted from AST
package types

import (
	"strings"
	"unicode"

	"github.com/gowebapi/webidlparser/ast"
)

// TypeName contains usage information about a type
type BasicInfo struct {
	// Idl name
	Idl string

	// Package name
	Package string

	// Def is short for definition of a type, e.g. Foo
	Def string

	// Internal name for used with methods and need to write some code
	Internal string

	// Template define a template name prefix/suffix
	Template string
}

// TypeName contains usage information about a type
type TypeInfo struct {
	BasicInfo

	// InOut is method input and output variable type, e.g. *Foo
	InOut string

	// Pointer is true if InOut is a pointer type
	Pointer bool

	// NeedRelease define if the type need a release handle
	NeedRelease bool

	// Nullable indicate that a null/nil value is a possibility
	Nullable bool

	// Optional input value
	Option bool

	// Variadic is variable number of input values
	Variadic bool
}

type MethodName struct {
	// Idl name
	Idl string

	// Def contains method name to use e.g. Foo
	Def string

	// Internal name for used with methods and need to write some code
	Internal string
}

type standardType struct {
	base        *ast.Base
	needRelease bool
	inuse       bool
}

type nameAndLink struct {
	base *ast.Base
	name MethodName
}

type changeTemplateType struct {
	template string
	real     TypeRef
}

type inuseLogic map[string]bool

var reservedKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "error": true, "for": true,
	"fallthrough": true, "func": true, "go": true, "goto": true, "if": true,
	"interface": true, "import": true, "map": true, "package": true,
	"range": true, "return": true, "select": true, "struct": true,
	"switch": true, "type": true, "var": true,
}

// clipString is removing any starting and ending '"' + spaces
func clipString(input string) string {
	if strings.HasPrefix(input, "\"") && strings.HasSuffix(input, "\"") {
		return strings.TrimSpace(input[1 : len(input)-1])
	}
	return input
}

func fromIdlToTypeName(pkg string, name string, tmpl string) BasicInfo {
	name = getIdlName(name)
	ret := BasicInfo{
		Package:  pkg,
		Idl:      name,
		Def:      fixLangName(toCamelCase(name, true)),
		Internal: fixLangName(toCamelCase(name, false)),
		Template: tmpl,
	}
	return ret
}

func fromIdlToMethodName(name string) MethodName {
	name = getIdlName(name)
	ret := MethodName{
		Idl:      name,
		Def:      fixLangName(toCamelCase(name, true)),
		Internal: fixLangName(toCamelCase(name, false)),
	}
	return ret
}

func getIdlName(input string) string {
	if strings.HasPrefix(input, "_") && len(input) > 1 {
		input = input[1:]
	}
	return input
}

func fixLangName(input string) string {
	if input == "" {
		return input
	}
	if _, f := reservedKeywords[input]; f {
		input = "_" + input
	}
	if len(input) > 1 && input[0] >= '0' && input[0] <= '9' {
		input = "_" + input
	}
	return input
}

func newTypeInfo(basic BasicInfo, nullable, option, variadic, pointer, disablePtr, release bool) *TypeInfo {
	if basic.Template == "" {
		panic("empty template name")
	}
	t := &TypeInfo{
		BasicInfo:   basic,
		InOut:       basic.Def,
		NeedRelease: release,
		Pointer:     (nullable || option || pointer) && !disablePtr,
		Nullable:    nullable,
		Option:      option,
		Variadic:    variadic,
	}
	if t.Pointer {
		t.InOut = "*" + t.InOut
	}
	if variadic {
		t.Def = "..." + t.Def
		t.InOut = "..." + t.InOut
	}
	return t
}

func toCamelCase(in string, upper bool) string {
	out := ""
	up := true
	for i, c := range in {
		if i == 0 && !upper {
			c = unicode.ToLower(c)
		} else if up {
			c = unicode.ToUpper(c)
		} else if c == '_' || c == '-' || c == ' ' || c == '\t' {
			up = true
			continue
		}
		out += string(c)
		up = false
	}
	return out
}

func IsVoid(t TypeRef) bool {
	_, isVoid := t.(*voidType)
	return isVoid
}

func (t *nameAndLink) NodeBase() *ast.Base {
	return t.base
}

func (t *nameAndLink) Name() *MethodName {
	return &t.name
}

func (t *standardType) NeedRelease() bool {
	return t.needRelease
}

func (t *standardType) NodeBase() *ast.Base {
	return t.base
}

func (t *standardType) InUse() bool {
	return t.inuse
}

func (t *inuseLogic) push(name string, ref ast.Node, conv *Convert) bool {
	_, ret := (*t)[name]
	if ret {
		conv.failing(ref, "circular typedef chain: %s: ", name)
		return false
	}
	(*t)[name] = true
	return true
}

func (t *inuseLogic) pop(name string) {
	delete(*t, name)
}

func ChangeTemplateName(on TypeRef, name string) TypeRef {
	return &changeTemplateType{
		real:     on,
		template: name,
	}
}

func (t *changeTemplateType) Basic() BasicInfo {
	in := t.real.Basic()
	in.Template = t.template
	return in
}

func (t *changeTemplateType) DefaultParam() (*TypeInfo, TypeRef) {
	info, ref := t.real.DefaultParam()
	copy := *info
	copy.Template = t.template
	return &copy, ref
}

func (t *changeTemplateType) NeedRelease() bool {
	return t.real.NeedRelease()
}

func (t *changeTemplateType) link(conv *Convert, inuse inuseLogic) TypeRef {
	panic("unsupported")
}

func (t *changeTemplateType) Param(nullable, optional, variadic bool) (*TypeInfo, TypeRef) {
	info, ref := t.real.Param(nullable, optional, variadic)
	copy := *info
	copy.Template = t.template
	return &copy, ref
}
