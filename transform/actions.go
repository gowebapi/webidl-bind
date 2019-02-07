package transform

import (
	"strings"

	"github.com/gowebapi/webidlgenerator/types"
)

type action interface {
	ExecuteCallback(instance *types.Callback, trans *Transform)
	ExecuteInterface(instance *types.Interface, targets map[string]renameTarget, trans *Transform)
}

type property struct {
	Name  string
	Value string
	Ref   ref
}

type rename struct {
	Name  string
	Value string
	Ref   ref
}

type renameTarget interface {
	Name() *types.MethodName
}

func (t *property) ExecuteCallback(instance *types.Callback, trans *Transform) {
	if f, ok := callbackProperties[t.Name]; ok {
		f(instance, t.Value)
	} else {
		trans.messageError(t.Ref, "unknow property '%s', valid are: %s",
			t.Name, strings.Join(callbackPropertyNames, ", "))
	}
}

func (t *property) ExecuteInterface(value *types.Interface, targets map[string]renameTarget, trans *Transform) {
	if f, ok := interfaceProperties[t.Name]; ok {
		f(value, t.Value)
	} else {
		trans.messageError(t.Ref, "unknow property '%s', valid are: %s",
			t.Name, strings.Join(interfacePropertyNames, ", "))
	}
}

func (t *rename) ExecuteCallback(instance *types.Callback, trans *Transform) {
	trans.messageError(t.Ref, "callback doesn't have any attributes or methods that can be renamed")
}

func (t *rename) ExecuteInterface(value *types.Interface, targets map[string]renameTarget, trans *Transform) {
	if target, found := targets[t.Name]; found {
		target.Name().Def = t.Value
	} else {
		trans.messageError(t.Ref, "unknown rename target '%s'", t.Name)
	}
}
