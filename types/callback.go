package types

import (
	"fmt"

	"github.com/dennwc/webidl/ast"
)

type Callback struct {
	name       Name
	Return     TypeRef
	Parameters []*Parameter
	source     *ast.Callback
}

func ConvertCallback(in *ast.Callback, setup *Setup) (*Callback, error) {
	params := []*Parameter{}
	for _, pi := range in.Parameters {
		po, err := convertParam(pi)
		if err != nil {
			return nil, fmt.Errorf("unable to convert parameter: %s: %s", pi.Name, err)
		}
		params = append(params, po)
	}
	ret := &Callback{
		source:     in,
		name:       fromIdlName(setup.Package, in.Name),
		Return:     convertType(in.Return),
		Parameters: params,
	}
	return ret, nil
}

func (cb *Callback) Name() Name {
	return cb.name
}

func (cb *Callback) Base() *ast.Base {
	return cb.source.NodeBase()
}

func (t *Callback) GetAllTypeRefs(list []TypeRef) []TypeRef {
	list = append(list, t.Return)
	for _, p := range t.Parameters {
		list = append(list, p.Type)
	}
	return list
}
