package types

import (
	"fmt"

	"github.com/dennwc/webidl/ast"
)

type typeDef struct {
	standardType
	basic BasicInfo
	Type  TypeRef
	name  string
}

var _ Type = &typeDef{}

func (t *extractTypes) convertTypeDef(in *ast.Typedef) *typeDef {
	ret := typeDef{
		standardType: standardType{
			base:        in.NodeBase(),
			needRelease: false,
		},
		basic: fromIdlToTypeName("", in.Name, "typedef"),
		Type:  convertType(in.Type),
		name:  in.Name,
	}
	for _, a := range in.Annotations {
		t.warning(a, "typedef: unsupported annotation '%s'", a.Name)
	}
	return &ret
}

func (t *typeDef) Basic() BasicInfo {
	panic("not supported for this type")
	// return t.basic
}

func (t *typeDef) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *typeDef) key() string {
	return t.basic.Idl
}

func (t *typeDef) link(conv *Convert, inuse inuseLogic) TypeRef {
	if inuse.push(t.name, t.base, conv) {
		t.Type = t.Type.link(conv, inuse)
		inuse.pop(t.name)
		return t.Type
	}
	fmt.Println("DEBUG: ", t.name, inuse)

	// at this point the source code should be consider faulty.
	// create a test that is checking that we do get an error
	// message at this point.
	panic("untested code, remove this panic")
	// if we are failing, we just return something
	// return t.Type
}

func (t *typeDef) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	panic("not supported for this type")
	// return t.Type.Param(nullable, option, variadic)
}
