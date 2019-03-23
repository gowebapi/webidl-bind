package transform

import (
	"regexp"
	"strings"

	"github.com/gowebapi/webidl-bind/types"
)

type action interface {
	OperateOn() scopeMode
	ExecuteCallback(instance *types.Callback, notify notifyMsg)
	ExecuteDictionary(instance *types.Dictionary, targets map[string]renameTarget, notify notifyMsg)
	ExecuteEnum(instance *types.Enum, targets map[string]renameTarget, notify notifyMsg)
	ExecuteInterface(instance *types.Interface, targets map[string]renameTarget, notify notifyMsg)
	ExecuteStatus(instance *SpecStatus, notify notifyMsg)
	Reference() ref
}

type notifyMsg interface {
	messageError(ref ref, format string, args ...interface{})
}

type scopeMode int

const (
	scopeGlobal scopeMode = iota
	scopeFile
	scopeType
)

// propary change on interface/enum/etc, like package name
type property struct {
	Name  string
	Value string
	Ref   ref
}

func (t *property) ExecuteCallback(instance *types.Callback, notify notifyMsg) {
	if f, ok := callbackProperties[t.Name]; ok {
		if msg := f(instance, t.Value); msg != "" {
			notify.messageError(t.Ref, msg)
		}
	} else {
		notify.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(callbackPropertyNames, ", "))
	}
}

func (t *property) ExecuteDictionary(instance *types.Dictionary, targets map[string]renameTarget, notify notifyMsg) {
	if f, ok := dictionaryProperties[t.Name]; ok {
		if msg := f(instance, t.Value); msg != "" {
			notify.messageError(t.Ref, msg)
		}
	} else {
		notify.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(dictionaryPropertyNames, ", "))
	}
}

func (t *property) ExecuteEnum(instance *types.Enum, targets map[string]renameTarget, notify notifyMsg) {
	if f, ok := enumProperties[t.Name]; ok {
		if msg := f(instance, t.Value); msg != "" {
			notify.messageError(t.Ref, msg)
		}
	} else {
		notify.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(enumPropertyNames, ", "))
	}
}

func (t *property) ExecuteInterface(instance *types.Interface, targets map[string]renameTarget, notify notifyMsg) {
	if f, ok := interfaceProperties[t.Name]; ok {
		if msg := f(instance, t.Value); msg != "" {
			notify.messageError(t.Ref, msg)
		}
	} else {
		notify.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(interfacePropertyNames, ", "))
	}
}

func (t *property) ExecuteStatus(instance *SpecStatus, notify notifyMsg) {
	if f, ok := fileProperties[t.Name]; ok {
		if msg := f(instance, t.Value); msg != "" {
			notify.messageError(t.Ref, msg)
		}
	} else {
		notify.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(filePropertyNames, ", "))
	}
}

func (t property) OperateOn() scopeMode {
	_, found := globalProperties[t.Name]
	if found {
		return scopeGlobal
	}
	_, found = fileProperties[t.Name]
	if found {
		return scopeFile
	}
	return scopeType
}

func (t property) Reference() ref {
	return t.Ref
}

// rename a method or attribute name
type rename struct {
	Name  string
	Value string
	Ref   ref
}

type renameTarget interface {
	Name() *types.MethodName
	SetType(value types.TypeRef) string
}

func (t *rename) ExecuteCallback(instance *types.Callback, notify notifyMsg) {
	notify.messageError(t.Ref, "callback doesn't have any attributes or methods that can be renamed")
}

func (t *rename) ExecuteDictionary(value *types.Dictionary, targets map[string]renameTarget, notify notifyMsg) {
	genericRename(t.Name, t.Value, t.Ref, targets, notify)
}

func (t *rename) ExecuteEnum(value *types.Enum, targets map[string]renameTarget, notify notifyMsg) {
	genericRename(t.Name, t.Value, t.Ref, targets, notify)
}

func (t *rename) ExecuteInterface(value *types.Interface, targets map[string]renameTarget, notify notifyMsg) {
	genericRename(t.Name, t.Value, t.Ref, targets, notify)
}

func (t *rename) ExecuteStatus(instance *SpecStatus, notify notifyMsg) {
	panic("unsupported")
}

func genericRename(name, value string, ref ref, targets map[string]renameTarget, notify notifyMsg) {
	if target, found := targets[name]; found {
		target.Name().Def = value
	} else {
		notify.messageError(ref, "unknown rename target '%s'", name)
	}
}

func (t *rename) OperateOn() scopeMode {
	return scopeType
}

func (t rename) Reference() ref {
	return t.Ref
}

// do a command on multiple types at one
type globalRegExp struct {
	Match *regexp.Regexp
	Type  matchType
	What  action
	Ref   ref
}

func (t *globalRegExp) OperateOn() scopeMode {
	return scopeGlobal
}

func (t globalRegExp) Reference() ref {
	return t.Ref
}

func (t *globalRegExp) ExecuteCallback(instance *types.Callback, notify notifyMsg) {
	t.What.ExecuteCallback(instance, notify)
}

func (t *globalRegExp) ExecuteDictionary(value *types.Dictionary, targets map[string]renameTarget, notify notifyMsg) {
	t.What.ExecuteDictionary(value, targets, notify)
}

func (t *globalRegExp) ExecuteEnum(value *types.Enum, targets map[string]renameTarget, notify notifyMsg) {
	t.What.ExecuteEnum(value, targets, notify)
}

func (t *globalRegExp) ExecuteInterface(value *types.Interface, targets map[string]renameTarget, notify notifyMsg) {
	t.What.ExecuteInterface(value, targets, notify)
}

func (t *globalRegExp) ExecuteStatus(instance *SpecStatus, notify notifyMsg) {
	panic("unsupported")
}

type changeType struct {
	Name  string
	RawJS string
	Ref   ref
}

func (t *changeType) OperateOn() scopeMode {
	return scopeType
}

func (t changeType) Reference() ref {
	return t.Ref
}

func (t *changeType) ExecuteCallback(instance *types.Callback, notify notifyMsg) {
	notify.messageError(t.Ref, "type change in callback in not yet implmented")
}

func (t *changeType) ExecuteDictionary(value *types.Dictionary, targets map[string]renameTarget, notify notifyMsg) {
	on, found := targets[t.Name]
	if !found {
		notify.messageError(t.Ref, "unknown reference")
		return
	}
	raw := types.NewRawJSType()
	if msg := on.SetType(raw); msg != "" {
		notify.messageError(t.Ref, "type change error: %s", msg)
	}
}

func (t *changeType) ExecuteEnum(value *types.Enum, targets map[string]renameTarget, notify notifyMsg) {
	notify.messageError(t.Ref, "type change for enum is not supported")
}

func (t *changeType) ExecuteInterface(value *types.Interface, targets map[string]renameTarget, notify notifyMsg) {
	on, found := targets[t.Name]
	if !found {
		notify.messageError(t.Ref, "unknown reference")
		return
	}
	raw := types.NewRawJSType()
	if msg := on.SetType(raw); msg != "" {
		notify.messageError(t.Ref, "type change error: %s", msg)
	}
}

func (t *changeType) ExecuteStatus(instance *SpecStatus, notify notifyMsg) {
	panic("unsupported")
}

type idlconst struct {
	Ref ref
}

func (t *idlconst) OperateOn() scopeMode {
	return scopeType
}

func (t idlconst) Reference() ref {
	return t.Ref
}

func (t *idlconst) ExecuteCallback(instance *types.Callback, notify notifyMsg) {
	notify.messageError(t.Ref, "not supported on this type")
}

func (t *idlconst) ExecuteDictionary(value *types.Dictionary, targets map[string]renameTarget, notify notifyMsg) {
	notify.messageError(t.Ref, "not supported on this type")
}

func (t *idlconst) ExecuteEnum(value *types.Enum, targets map[string]renameTarget, notify notifyMsg) {
	notify.messageError(t.Ref, "not supported on this type")
}

func (t *idlconst) ExecuteInterface(value *types.Interface, targets map[string]renameTarget, notify notifyMsg) {
	for _, c := range value.Consts {
		m := c.Name()
		idl := m.Idl
		m.Def = strings.ToUpper(idl[:1]) + idl[1:]
		c.SetName(m)
	}
}

func (t *idlconst) ExecuteStatus(instance *SpecStatus, notify notifyMsg) {
	panic("unsupported")
}
