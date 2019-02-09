package transform

import (
	"strings"

	"github.com/gowebapi/webidlgenerator/types"
)

type action interface {
	IsGlobal() bool
	ExecuteCallback(instance *types.Callback, trans *Transform)
	ExecuteDictionary(instance *types.Dictionary, trans *Transform)
	ExecuteEnum(instance *types.Enum, trans *Transform)
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

func (t *property) ExecuteDictionary(instance *types.Dictionary, trans *Transform) {
	if f, ok := dictionaryProperties[t.Name]; ok {
		f(instance, t.Value)
	} else {
		trans.messageError(t.Ref, "unknow property '%s', valid are: %s",
			t.Name, strings.Join(dictionaryPropertyNames, ", "))
	}
}

func (t *property) ExecuteEnum(instance *types.Enum, trans *Transform) {
	if f, ok := enumProperties[t.Name]; ok {
		f(instance, t.Value)
	} else {
		trans.messageError(t.Ref, "unknow property '%s', valid are: %s",
			t.Name, strings.Join(enumPropertyNames, ", "))
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

func (t property) IsGlobal() bool {
	value, found := globalProperties[t.Name]
	return value && found
}

func (t *rename) ExecuteCallback(instance *types.Callback, trans *Transform) {
	trans.messageError(t.Ref, "callback doesn't have any attributes or methods that can be renamed")
}

func (t *rename) ExecuteDictionary(value *types.Dictionary, trans *Transform) {
	trans.messageError(t.Ref, "dictionary doesn't have any attributes or methods that can be renamed")
}

func (t *rename) ExecuteEnum(value *types.Enum, trans *Transform) {
	trans.messageError(t.Ref, "enum doesn't have any attributes or methods that can be renamed")
}

func (t *rename) ExecuteInterface(value *types.Interface, targets map[string]renameTarget, trans *Transform) {
	if target, found := targets[t.Name]; found {
		target.Name().Def = t.Value
	} else {
		trans.messageError(t.Ref, "unknown rename target '%s'", t.Name)
	}
}

func (t *rename) IsGlobal() bool {
	return false
}
