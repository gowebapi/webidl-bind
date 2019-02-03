// contains types converted from AST
package types

import (
	"strings"
	"unicode"

	"github.com/dennwc/webidl/ast"
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

	// Vardict is variable number of input values
	Vardict bool
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
}

type inuseLogic map[string]bool

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
		Def:      toCamelCase(name, true),
		Internal: toCamelCase(name, false),
		Template: tmpl,
	}
	return ret
}

func fromIdlToMethodName(name string) MethodName {
	name = getIdlName(name)
	ret := MethodName{
		Idl:      name,
		Def:      toCamelCase(name, true),
		Internal: toCamelCase(name, false),
	}
	return ret
}

func getIdlName(input string) string {
	if strings.HasPrefix(input, "_") && len(input) > 1 {
		input = input[1:]
	}
	return input
}

func newTypeInfo(basic BasicInfo, nullable, option, vardict, pointer, disablePtr, release bool) *TypeInfo {
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
		Vardict:     vardict,
	}
	if t.Pointer {
		t.InOut = "*" + t.InOut
	}
	if vardict {
		t.Def = "..." + t.Def
		t.InOut = "..." + t.InOut
	}
	return t
}

func stringToConst(input string) string {
	return toCamelCase(clipString(input), true)
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
