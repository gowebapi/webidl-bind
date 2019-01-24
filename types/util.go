// contains types converted from AST
package types

import (
	"strings"
	"unicode"

	"github.com/dennwc/webidl/ast"
)

type standardType struct {
	base *ast.Base
	name Name
}

func clipString(input string) string {
	if strings.HasPrefix(input, "\"") && strings.HasSuffix(input, "\"") {
		return strings.TrimSpace(input[1 : len(input)-1])
	}
	return input
}

func fromIdlName(pkg string, name string) Name {
	if strings.HasPrefix(name, "_") && len(name) > 1 {
		name = name[1:]
	}
	return Name{
		Package: pkg,
		Idl:     name,
		Public:  toCamelCase(name, true),
		Local:   toCamelCase(name, false),
	}
}

func fromMethodName(name string) Name {
	if strings.HasPrefix(name, "_") && len(name) > 1 {
		name = name[1:]
	}
	return Name{
		Package: "",
		Idl:     name,
		Public:  toCamelCase(name, true),
		Local:   toCamelCase(name, false),
	}
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

func (t *standardType) NodeBase() *ast.Base {
	return t.base
}

func (t *standardType) Name() Name {
	return t.name
}
