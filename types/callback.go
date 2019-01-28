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

// Callback need to implement Type
var _ Type = &Callback{}

func (t *extractTypes) convertCallback(in *ast.Callback) *Callback {
	params := t.convertParams(in.Parameters)
	ret := &Callback{
		standardType: standardType{
			base:        in.NodeBase(),
			name:        fromIdlName(t.main.setup.Package, in.Name, false),
			needRelease: true,
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

func (t *Callback) TemplateName() (string, TemplateNameFlags) {
	return "callback", NoTnFlag
}
