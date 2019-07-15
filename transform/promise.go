package transform

import (
	"fmt"

	"github.com/gowebapi/webidl-bind/types"
)

// promiseCrossReference is used to duplicate an insert into the
// promise map. It need to be cross references between types to work
var promiseCrossReference = map[string]string{
	"DOMString":           "USVString",
	"USVString":           "DOMString",
	"sequence<DOMString>": "sequence<USVString>",
	"sequence<USVString>": "sequence<DOMString>",
}

func (t *Transform) setupPromiseEvaluation(source map[string]types.Type) (map[string]types.Type, map[string]struct{}) {
	dest := make(map[string]types.Type)
	main := trySetupPromise("any", "Promise", dest, source)
	main.SetInUse(true)
	trySetupPromise("void", "PromiseVoid", dest, source)
	t.promiseTemplate = trySetupPromise("*template*", "PromiseTemplate", dest, source)
	// to hide the type from generated output
	trySetupPromise("*template-value-not-used*", "PromiseTemplateValue", dest, source)
	trySetupPromise("*template-fulfilled-do-not-use", "PromiseTemplateOnFulfilled", dest, source)
	trySetupPromise("*template-rejected-do-not-use", "PromiseTemplateOnRejected", dest, source)

	primitive := make(map[string]struct{})
	for k := range dest {
		primitive[k] = struct{}{}
	}
	return dest, primitive
}

func trySetupPromise(key, typeName string, dest map[string]types.Type, source map[string]types.Type) types.Type {
	if typ, found := source[typeName]; found {
		dest[key] = typ
		typ.SetInUse(false)
		return typ
	}
	return nil
}

// evaluatePromise is verifying that all promise usage have a new
// dedicated type
func (t *Transform) evaluatePromise(typ types.Type, promises map[string]types.Type) {
	if !typ.InUse() {
		return
	}
	switch typ := typ.(type) {
	case *types.Callback:
		t.evaluatePromiseCallback(typ, promises)
	case *types.Dictionary:
		t.evaluatePromiseDictionary(typ, promises)
	case *types.Enum:
	case *types.Interface:
		t.evaluatePromiseInterface(typ, promises)
	default:
		panic(fmt.Sprintf("unknown type: %T", typ))
	}
}

func (t *Transform) evaluatePromiseCallback(item *types.Callback, promises map[string]types.Type) {
	item.Return = t.modifyPromise(item.Return, "return", "", promises)
	t.evaluatePromiseParameters(item.Parameters, promises)
}

func (t *Transform) evaluatePromiseDictionary(item *types.Dictionary, promises map[string]types.Type) {
	for _, m := range item.Members {
		m.Type = t.modifyPromise(m.Type, "member", m.Name().Idl, promises)
	}
}

func (t *Transform) evaluatePromiseInterface(item *types.Interface, promises map[string]types.Type) {
	for _, m := range item.Method {
		m.Return = t.modifyPromise(m.Return, "return", "", promises)
		t.evaluatePromiseParameters(m.Params, promises)
	}
	for _, m := range item.StaticMethod {
		m.Return = t.modifyPromise(m.Return, "return", "", promises)
		t.evaluatePromiseParameters(m.Params, promises)
	}
}

func (t *Transform) evaluatePromiseParameters(list []*types.Parameter, promises map[string]types.Type) {
	for _, p := range list {
		p.Type = t.modifyPromise(p.Type, "parameter", p.Name, promises)
	}
}

func (t *Transform) modifyPromise(value types.TypeRef, from, name string, promises map[string]types.Type) types.TypeRef {
	in, found := value.(*types.ParametrizedType)
	if !found {
		return value
	}
	r := convertRef(in.Ref)
	if in.ParamName == "FrozenArray" {
		// TODO support FrozenArray
		// t.messageError(r, "%s:%s:FrozenArray type need to be transformed", from, name)
		return value
	}
	if in.ParamName != "Promise" {
		panic(in.ParamName)
	}
	src := in.Elems[0]
	info, inner := src.DefaultParam()
	prefix, typeName := "Promise", inner.Basic().Def
	internal, isprim := isPromiseBuiltInType(inner)
	key := info.Idl
	if info.Nullable {
		key += "?"
		prefix += "Nil"
	}
	pkg := src.Basic().Package

	if seq, ok := inner.(*types.SequenceType); ok {
		prefix += "Sequence"
		internal, isprim = isPromiseBuiltInType(seq.Elem)
		typeName = seq.Elem.Basic().Def
		pkg = seq.Elem.Basic().Package

	} else if _, ok := inner.(*types.ParametrizedType); ok {
		// fmt.Println("SKIPPING PT: ", value.Basic())
		// return value
	}

	if old, exist := promises[key]; exist {
		old.SetInUse(true)
		return old
	}
	if internal {
		t.messageError(r, "%s:%s: unable to transform primitive/void/any promise type. please define it in IDL (key: '%s')", from, name, key)
		return value
	} else if isprim != "" {
		// pkg = "github.com/gowebapi/webapi/javascript"
		pkg = t.promiseTemplate.Basic().Package
		typeName = isprim
	}

	// creating a new type
	next := t.createNewInterfacePromiseType(pkg, prefix, typeName, value, promises, r)
	fulfilled := t.createNewPromiseCallbackType(pkg, prefix, typeName+"OnFulfilled", promises, r, "*template-fulfilled-do-not-use")
	reject := t.createNewPromiseCallbackType(pkg, prefix, typeName+"OnRejected", promises, r, "*template-rejected-do-not-use")

	if next == nil || fulfilled == nil || reject == nil {
		// on error
		return value
	}
	typeConv := func(in types.TypeRef) types.TypeRef {
		idl := in.Basic().Idl
		if idl == "PromiseTemplateValue" {
			return inner
		} else if idl == "PromiseTemplate" {
			return next
		} else if idl == "PromiseTemplateOnFulfilled" {
			return fulfilled
		} else if idl == "PromiseTemplateOnRejected" {
			return reject
		}
		return in
	}
	next.ChangeType(typeConv)
	fulfilled.ChangeType(typeConv)
	reject.ChangeType(typeConv)

	promises[key] = next
	promises[key+"+OnFulfilled"] = fulfilled
	promises[key+"+OnRejected"] = reject

	insertPromiseType(key, next, fulfilled, reject, promises)
	if other, found := promiseCrossReference[key]; found {
		insertPromiseType(other, next, fulfilled, reject, promises)
	}
	return next
}

func insertPromiseType(key string, next, fulfilled, reject types.Type, promises map[string]types.Type) {
	promises[key] = next
	promises[key+"+OnFulfilled"] = fulfilled
	promises[key+"+OnRejected"] = reject

}

func (t *Transform) createNewInterfacePromiseType(pkg, prefix, typeName string, value types.TypeRef, promises map[string]types.Type, r ref) *types.Interface {
	tmpl := t.findInterfaceTemplate("*template*", promises, r)
	if tmpl == nil {
		return nil
	}
	bvalue := value.Basic()
	basic := types.BasicInfo{
		Idl:      bvalue.Idl,
		Package:  pkg,
		Def:      prefix + typeName,
		Internal: "internal" + prefix + typeName,
		Template: "",
	}
	next := tmpl.TemplateCopy(basic)
	next.SetInUse(true)
	return next
}

func (t *Transform) findInterfaceTemplate(name string, promises map[string]types.Type, r ref) *types.Interface {
	tmplType, exist := promises[name]
	if !exist {
		t.messageError(r, "unable to find PromiseTemplate to use as Promise")
		return nil
	}
	tmpl, ok := tmplType.(*types.Interface)
	if !ok {
		t.messageError(r, "PromiseTemplate must be an interface type")
		return nil
	}
	return tmpl
}

func (t *Transform) createNewPromiseCallbackType(pkg, prefix, typeName string, promises map[string]types.Type, r ref, key string) *types.Callback {
	tmpl := t.findCallbackTemplate(key, promises, r)
	if tmpl == nil {
		return nil
	}
	basic := types.BasicInfo{
		Idl:      tmpl.Basic().Idl,
		Package:  pkg,
		Def:      prefix + typeName,
		Internal: "internal" + prefix + typeName,
		Template: "",
	}
	next := tmpl.TemplateCopy(basic)
	next.SetInUse(true)
	return next
}

func (t *Transform) findCallbackTemplate(name string, promises map[string]types.Type, r ref) *types.Callback {
	tmplType, exist := promises[name]
	if !exist {
		t.messageError(r, "unable to find PromiseTemplate to use as Promise")
		return nil
	}
	tmpl, ok := tmplType.(*types.Callback)
	if !ok {
		t.messageError(r, "PromiseTemplate must be an interface type")
		return nil
	}
	return tmpl
}

func isPromiseBuiltInType(value types.TypeRef) (bool, string) {
	isvoid := types.IsVoid(value)
	_, isany := value.(*types.AnyType)
	name := ""
	if prim, isprim := value.(*types.PrimitiveType); isprim {
		name = prim.JsMethod
	}
	return isvoid || isany, name
}
