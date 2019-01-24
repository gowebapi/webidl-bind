package types

import (
	"github.com/dennwc/webidl/ast"
)

type Callback struct {
	standardType
	Return     TypeRef
	Parameters []*Parameter
	source     *ast.Callback
}

func (t *extractTypes) convertCallback(in *ast.Callback) *Callback {
	params := []*Parameter{}
	for _, pi := range in.Parameters {
		po := t.convertParam(pi)
		params = append(params, po)
	}
	ret := &Callback{
		standardType: standardType{
			base: in.NodeBase(),
			name: fromIdlName(t.main.setup.Package, in.Name),
		},
		source:     in,
		Return:     convertType(in.Return),
		Parameters: params,
	}
	return ret
}

func (t *Callback) GetAllTypeRefs(list []TypeRef) []TypeRef {
	list = append(list, t.Return)
	for _, p := range t.Parameters {
		list = append(list, p.Type)
	}
	return list
}
