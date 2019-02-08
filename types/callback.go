package types

import (
	"github.com/gowebapi/webidlparser/ast"
)

type Callback struct {
	standardType
	basic      BasicInfo
	Return     TypeRef
	Parameters []*Parameter
	source     *ast.Callback
}

// Callback need to implement Type
var _ Type = &Callback{}

func (t *extractTypes) convertCallback(in *ast.Callback) *Callback {
	params := t.convertParams(in.Parameters)
	ret := &Callback{
		standardType: standardType{
			base:        in.NodeBase(),
			needRelease: false,
		},
		basic:      fromIdlToTypeName(t.main.setup.Package, in.Name, "callback"),
		source:     in,
		Return:     convertType(in.Return),
		Parameters: params,
	}
	return ret
}

func (t *Callback) Basic() BasicInfo {
	return t.basic
}

func (t *Callback) DefaultParam() (info *TypeInfo, inner TypeRef) {
	return t.Param(false, false, false)
}

func (t *Callback) key() string {
	return t.basic.Idl
}

func (t *Callback) link(conv *Convert, inuse inuseLogic) TypeRef {
	if t.inuse {
		return t
	}
	t.inuse = true

	t.Return = t.Return.link(conv, make(inuseLogic))
	for i := range t.Parameters {
		t.Parameters[i].Type = t.Parameters[i].Type.link(conv, make(inuseLogic))
	}
	return t
}

func (t *Callback) Param(nullable, option, variadic bool) (info *TypeInfo, inner TypeRef) {
	info, typ := newTypeInfo(t.basic, nullable, option, variadic, false, true, false), t
	info.Input = "*js.Callback"
	info.Pointer = true
	return info, typ
}

func (t *Callback) SetBasic(basic BasicInfo) {
	t.basic = basic
}
