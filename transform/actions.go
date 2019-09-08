package transform

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gowebapi/webidl-bind/types"
)

type action interface {
	OperateOn() scopeMode
	ExecuteCallback(instance *types.Callback, data *actionData)
	ExecuteDictionary(instance *types.Dictionary, data *actionData)
	ExecuteEnum(instance *types.Enum, data *actionData)
	ExecuteInterface(instance *types.Interface, data *actionData)
	ExecuteStatus(instance *SpecStatus, data *actionData)
	Reference() ref
}

type actionData struct {
	notify    notifyMsg
	targets   map[string]renameTarget
	conv      *types.Convert
	eventMap  map[string]struct{}
	eventAttr []arg
	lastGroup string
}

func (ad *actionData) nextType(value types.Type) {
	group := groupName(value.SourceReference().Filename)
	if group != ad.lastGroup {
		ad.eventAttr = nil
	}
	ad.targets = nil
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

type abstractAction struct {
	Ref ref
}

func (t *abstractAction) Reference() ref {
	return t.Ref
}

func (t *abstractAction) ExecuteCallback(instance *types.Callback, data *actionData) {
	panic("unsupported")
}

func (t *abstractAction) ExecuteDictionary(instance *types.Dictionary, data *actionData) {
	panic("unsupported")
}

func (t *abstractAction) ExecuteEnum(instance *types.Enum, data *actionData) {
	panic("unsupported")
}

func (t *abstractAction) ExecuteInterface(instance *types.Interface, data *actionData) {
	panic("unsupported")
}

func (t *abstractAction) ExecuteStatus(instance *SpecStatus, data *actionData) {
	panic("unsupported")
}

// propary change on interface/enum/etc, like package name
type property struct {
	Name  string
	Value string
	Ref   ref
}

func (t *property) ExecuteCallback(instance *types.Callback, data *actionData) {
	if f, ok := callbackProperties[t.Name]; ok {
		if msg := f.Set(instance, t.Value); msg != "" {
			data.notify.messageError(t.Ref, msg)
		}
	} else {
		data.notify.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(callbackPropertyNames, ", "))
	}
}

func (t *property) ExecuteDictionary(instance *types.Dictionary, data *actionData) {
	if f, ok := dictionaryProperties[t.Name]; ok {
		if msg := f.Set(instance, t.Value); msg != "" {
			data.notify.messageError(t.Ref, msg)
		}
	} else {
		data.notify.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(dictionaryPropertyNames, ", "))
	}
}

func (t *property) ExecuteEnum(instance *types.Enum, data *actionData) {
	if f, ok := enumProperties[t.Name]; ok {
		if msg := f.Set(instance, t.Value); msg != "" {
			data.notify.messageError(t.Ref, msg)
		}
	} else {
		data.notify.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(enumPropertyNames, ", "))
	}
}

func (t *property) ExecuteInterface(instance *types.Interface, data *actionData) {
	if f, ok := interfaceProperties[t.Name]; ok {
		if msg := f.Set(instance, t.Value); msg != "" {
			data.notify.messageError(t.Ref, msg)
		}
	} else {
		data.notify.messageError(t.Ref, "unknown property '%s', valid are: %s",
			t.Name, strings.Join(interfacePropertyNames, ", "))
	}
}

func (t *property) ExecuteStatus(instance *SpecStatus, data *actionData) {
	if f, ok := fileProperties[t.Name]; ok {
		if msg := f.Set(instance, t.Value); msg != "" {
			data.notify.messageError(t.Ref, msg)
		}
	} else {
		data.notify.messageError(t.Ref, "unknown property '%s', valid are: %s",
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
	abstractAction
	Name  string
	Value string
}

type renameTarget interface {
	Name() *types.MethodName
	GetType() types.TypeRef
	SetType(value types.TypeRef) string
}

func (t *rename) ExecuteCallback(instance *types.Callback, data *actionData) {
	data.notify.messageError(t.Ref, "callback doesn't have any attributes or methods that can be renamed")
}

func (t *rename) ExecuteDictionary(value *types.Dictionary, data *actionData) {
	genericRename(t.Name, t.Value, t.Ref, data.targets, data.notify)
}

func (t *rename) ExecuteEnum(value *types.Enum, data *actionData) {
	genericRename(t.Name, t.Value, t.Ref, data.targets, data.notify)
}

func (t *rename) ExecuteInterface(value *types.Interface, data *actionData) {
	genericRename(t.Name, t.Value, t.Ref, data.targets, data.notify)
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

// do a command on multiple types at one
type globalRegExp struct {
	abstractAction
	Match *regexp.Regexp
	Type  matchType
	What  action
}

func (t *globalRegExp) OperateOn() scopeMode {
	return scopeGlobal
}

func (t *globalRegExp) ExecuteCallback(instance *types.Callback, data *actionData) {
	t.What.ExecuteCallback(instance, data)
}

func (t *globalRegExp) ExecuteDictionary(value *types.Dictionary, data *actionData) {
	t.What.ExecuteDictionary(value, data)
}

func (t *globalRegExp) ExecuteEnum(value *types.Enum, data *actionData) {
	t.What.ExecuteEnum(value, data)
}

func (t *globalRegExp) ExecuteInterface(value *types.Interface, data *actionData) {
	t.What.ExecuteInterface(value, data)
}

type changeType struct {
	abstractAction
	Name  string
	RawJS string
}

func (t *changeType) OperateOn() scopeMode {
	return scopeType
}

func (t *changeType) ExecuteCallback(instance *types.Callback, data *actionData) {
	data.notify.messageError(t.Ref, "type change in callback in not yet implmented")
}

func (t *changeType) ExecuteDictionary(value *types.Dictionary, data *actionData) {
	on, found := data.targets[t.Name]
	if !found {
		data.notify.messageError(t.Ref, "unknown reference")
		return
	}
	idl := on.GetType().Basic().Idl
	raw := types.NewRawJSType(idl)
	if msg := on.SetType(raw); msg != "" {
		data.notify.messageError(t.Ref, "type change error: %s", msg)
	}
}

func (t *changeType) ExecuteEnum(value *types.Enum, data *actionData) {
	data.notify.messageError(t.Ref, "type change for enum is not supported")
}

func (t *changeType) ExecuteInterface(value *types.Interface, data *actionData) {
	on, found := data.targets[t.Name]
	if !found {
		data.notify.messageError(t.Ref, "unknown reference")
		return
	}
	idl := on.GetType().Basic().Idl
	raw := types.NewRawJSType(idl)
	if msg := on.SetType(raw); msg != "" {
		data.notify.messageError(t.Ref, "type change error: %s", msg)
	}
}

type idlconst struct {
	abstractAction
}

func (t *idlconst) OperateOn() scopeMode {
	return scopeType
}

func (t *idlconst) ExecuteCallback(instance *types.Callback, data *actionData) {
	data.notify.messageError(t.Ref, "not supported on this type")
}

func (t *idlconst) ExecuteDictionary(value *types.Dictionary, data *actionData) {
	data.notify.messageError(t.Ref, "not supported on this type")
}

func (t *idlconst) ExecuteEnum(value *types.Enum, data *actionData) {
	data.notify.messageError(t.Ref, "not supported on this type")
}

func (t *idlconst) ExecuteInterface(value *types.Interface, data *actionData) {
	for _, c := range value.Consts {
		m := c.Name()
		idl := m.Idl
		m.Def = strings.ToUpper(idl[:1]) + idl[1:]
		c.SetName(m)
	}
}

type replace struct {
	abstractAction
	Property string
	From     string
	To       string
}

func (t *replace) OperateOn() scopeMode {
	return scopeType
}

func (t *replace) exec(in string) string {
	return strings.Replace(in, t.From, t.To, -1)
}

func (t *replace) ExecuteCallback(instance *types.Callback, data *actionData) {
	if p, ok := callbackProperties[t.Property]; ok {
		value := p.Get(instance)
		value = t.exec(value)
		if msg := p.Set(instance, value); msg != "" {
			data.notify.messageError(t.Ref, msg)
		}
	} else {
		data.notify.messageError(t.Ref, "%s: unknown property '%s', valid are: %s",
			instance.Basic().Idl, t.Property, strings.Join(callbackPropertyNames, ", "))
	}
}

func (t *replace) ExecuteDictionary(instance *types.Dictionary, data *actionData) {
	if p, ok := dictionaryProperties[t.Property]; ok {
		value := p.Get(instance)
		value = t.exec(value)
		if msg := p.Set(instance, value); msg != "" {
			data.notify.messageError(t.Ref, msg)
		}
	} else {
		data.notify.messageError(t.Ref, "%s: unknown property '%s', valid are: %s",
			instance.Basic().Idl, t.Property, strings.Join(dictionaryPropertyNames, ", "))
	}
}

func (t *replace) ExecuteEnum(instance *types.Enum, data *actionData) {
	if p, ok := enumProperties[t.Property]; ok {
		value := p.Get(instance)
		value = t.exec(value)
		if msg := p.Set(instance, value); msg != "" {
			data.notify.messageError(t.Ref, msg)
		}
	} else {
		data.notify.messageError(t.Ref, "%s: unknown property '%s', valid are: %s",
			instance.Basic().Idl, t.Property, strings.Join(enumPropertyNames, ", "))
	}
}

func (t *replace) ExecuteInterface(instance *types.Interface, data *actionData) {
	if p, ok := interfaceProperties[t.Property]; ok {
		value := p.Get(instance)
		value = t.exec(value)
		if msg := p.Set(instance, value); msg != "" {
			data.notify.messageError(t.Ref, msg)
		}
	} else {
		data.notify.messageError(t.Ref, "%s: unknown property '%s', valid are: %s",
			instance.Basic().Idl, t.Property, strings.Join(interfacePropertyNames, ", "))
	}
}

type commonEventData struct {
	abstractAction
	Method     string
	EventType  string
	EventName  string
	Cancelable bool
	Bubbles    bool
}

func (ce *commonEventData) set(method, eventType string) {
	ce.Method = method
	ce.EventType = eventType
	ce.EventName = strings.ToLower(method)
}

func (ce *commonEventData) addEventToInterface(inf *types.Interface, ev *types.IfVar, data *actionData) {
	// find type
	typ, ok := data.conv.Types[ce.EventType]
	if !ok {
		data.notify.messageError(ce.Ref, "unable to find type '%s'", ce.EventType)
		return
	}

	// create new instance
	ev.Name().Def = "On" + ce.Method
	ev.ShortName = ce.Method
	ev.EventName = ce.EventName
	ev.Type = typ
	inf.Events = append(inf.Events, ev)
}

func (ce *commonEventData) processArgs(args []arg, fail func(format string, args ...interface{})) {
	var err error
	for _, a := range args {
		switch a.Name {
		case "bubbles":
			if ce.Bubbles, err = strconv.ParseBool(a.Value); err != nil {
				fail("bubbles have faulty boolean value: %s", err)
			}
		case "cancelable":
			if ce.Cancelable, err = strconv.ParseBool(a.Value); err != nil {
				fail("cancelable have faulty boolean value: %s", err)
			}
		case "maybe":
		default:
			fail("unknown property '%s'. valid are 'bubbles,cancelable'", a.Name)
		}
	}
}

type event struct {
	commonEventData
}

func (t *event) OperateOn() scopeMode {
	return scopeType
}

func (t *event) ExecuteInterface(instance *types.Interface, data *actionData) {
	searchFor := "on" + strings.ToLower(t.Method)
	for _, attr := range instance.Vars {
		if attr.Name().Idl == searchFor {
			t.eventFuncModify(instance, attr, data)
			key := instance.Basic().Idl + "." + attr.Name().Idl
			data.eventMap[key] = struct{}{}
			return
		}
	}
	data.notify.messageError(t.Ref, "unable to find event '%s' with 'on' attribute '%s'", t.Method, searchFor)
}

func (t *event) eventFuncModify(inf *types.Interface, attr *types.IfVar, data *actionData) {
	// is we a callback attribute?
	_, inner := attr.Type.DefaultParam()
	_, ok := inner.(*types.Callback)
	if !ok {
		data.notify.messageError(t.Ref, "expected a callback function as value")
		return
	}

	t.addEventToInterface(inf, attr.Copy(), data)

	// modify current
	attr.Name().Def = "On" + t.Method
	attr.Readonly = true
}

type addevent struct {
	abstractAction
	commonEventData
}

func (t *addevent) OperateOn() scopeMode {
	return scopeType
}

func (t *addevent) ExecuteInterface(instance *types.Interface, data *actionData) {
	searchFor := "on" + strings.ToLower(t.Method)
	for _, attr := range instance.Vars {
		if attr.Name().Idl == searchFor {
			t.addEventToInterface(instance, attr.Copy(), data)
			return
		}
	}
	data.notify.messageError(t.Ref, "unable to find event '%s' with 'on' attribute '%s'", t.Method, searchFor)
}

type notEvent struct {
	abstractAction
	AttributeName string
}

func (t *notEvent) OperateOn() scopeMode {
	return scopeType
}

func (t *notEvent) ExecuteInterface(instance *types.Interface, data *actionData) {
	// searchFor := strings.ToLower(t.AttributeName)
	searchFor := t.AttributeName
	names := []string{}
	for _, attr := range instance.Vars {
		if attr.Name().Idl == searchFor {
			key := instance.Basic().Idl + "." + attr.Name().Idl
			data.eventMap[key] = struct{}{}
			return
		}
		names = append(names, attr.Name().Idl)
	}
	sort.Strings(names)
	data.notify.messageError(t.Ref, "unable to find attribute '%s'. valid are %s",
		t.AttributeName, strings.Join(names, ", "))
}

type setEventProp struct {
	abstractAction
	Args []arg
}

func (t *setEventProp) OperateOn() scopeMode {
	return scopeType
}

func (t *setEventProp) ExecuteInterface(instance *types.Interface, data *actionData) {
	data.eventAttr = append(data.eventAttr, t.Args...)
}
