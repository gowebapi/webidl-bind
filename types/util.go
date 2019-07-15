// contains types converted from AST
package types

import (
	"bytes"
	"fmt"
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

func (a *BasicInfo) lessThan(b *BasicInfo) bool {
	if a.Package != b.Package {
		return a.Package < b.Package
	}
	if a.Def != b.Def {
		return a.Def < b.Def
	}
	return a.Idl < b.Idl
}

// TypeName contains usage information about a type
type TypeInfo struct {
	BasicInfo

	// Input is method parameter type
	Input string

	// VarIn used in variable definition in Input cases
	VarIn string

	// VarInInner is the inner type of a variadic/sequence in Input cases
	VarInInner string

	// Output define type out from a method
	Output string

	// VarOut used in variable definition in Output cases
	VarOut string

	// VarOut is the intter type of variadic/sequence in Output cases
	VarOutInner string

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

// Reference in input file
type Ref struct {
	Filename      string
	Line          int
	TransformFile string
}

type GetRef interface {
	SourceReference() *Ref
}

type standardType struct {
	ref         *Ref
	needRelease bool
	inuse       bool
}

type nameAndLink struct {
	ref  *Ref
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

// convertIntoValidVariableName will convert any non-variable
// accepted chars will be turned into _ instead.
func convertIntoValidVariableName(input string) string {
	out := ""
	for _, c := range input {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'A' && c <= 'Z':
		case c >= 'a' && c <= 'z':
		case c == '_':
		default:
			c = '_'
		}
		out += string(c)
	}
	return out
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
		Input:       basic.Def,
		Output:      basic.Def,
		VarIn:       basic.Def,
		VarInInner:  basic.Def,
		VarOut:      basic.Def,
		VarOutInner: basic.Def,
		NeedRelease: release,
		Pointer:     (nullable || option || pointer) && !disablePtr,
		Nullable:    nullable,
		Option:      option,
		Variadic:    variadic,
	}
	if t.Pointer {
		t.Input = "*" + t.Input
		t.Output = "*" + t.Output
		t.VarIn = "*" + t.VarIn
		t.VarInInner = "*" + t.VarInInner
		t.VarOut = "*" + t.VarOut
		t.VarOutInner = "*" + t.VarOutInner
	}
	if variadic {
		t.Input = "..." + t.Input
		t.VarIn = "[]" + t.VarIn
		t.Output = "[]" + t.Output
		t.VarOut = "[]" + t.VarOut
	}
	return t
}

// toCamelCase is convert a constant into camel case
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

// insertLineNumber is used to debug print internal type specification
// and is adding line number to every line
func insertLineNumber(content []byte) []byte {
	split := bytes.Split(content, []byte("\n"))
	out := make([]byte, 0, len(content)+5*len(split))
	nl := []byte("\n")
	for idx, line := range split {
		number := fmt.Sprintf("%03d ", idx+1)
		out = append(out, []byte(number)...)
		out = append(out, line...)
		out = append(out, nl...)
	}
	return out
}

func IsString(t TypeRef) bool {
	p, ok := t.(*PrimitiveType)
	if ok {
		if p.Lang == "string" {
			return true
		}
	}
	return false
}

func IsVoid(t TypeRef) bool {
	_, isVoid := t.(*voidType)
	return isVoid
}

func isUnsignedInt(t TypeRef) bool {
	if prim, ok := t.(*PrimitiveType); ok {
		if prim.Lang == "uint" {
			return true
		}
	}
	return false
}

func createRef(in ast.Node, et *extractTypes) *Ref {
	return &Ref{
		Filename: et.main.setup.Filename,
		Line:     in.NodeBase().Line + et.lineOffset,
	}
}

func (t *nameAndLink) SourceReference() *Ref {
	return t.ref
}

func (t *nameAndLink) Name() *MethodName {
	return &t.name
}

func (t *nameAndLink) SetName(value *MethodName) {
	t.name = *value
}

func (t *Ref) sourceLessThan(other *Ref) bool {
	if t.Filename != other.Filename {
		return t.Filename < other.Filename
	}
	return t.Line < other.Line
}

func (t *Ref) SourceReference() *Ref {
	return t
}

func (t *Ref) String() string {
	return fmt.Sprintf("%s:%d", t.Filename, t.Line)
}

func (t *standardType) NeedRelease() bool {
	return t.needRelease
}

func (t *standardType) InUse() bool {
	return t.inuse
}

func (t *standardType) SetInUse(value bool) {
	t.inuse = value
}

func (t *standardType) SourceReference() *Ref {
	return t.ref
}

func (t *inuseLogic) push(name string, ref GetRef, conv *Convert) bool {
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
