package transform

import (
	"sort"

	"github.com/gowebapi/webidlgenerator/types"
)

var globalProperties = map[string]bool{
	"package": true,
}
var globalPropertyNames = []string{}

var callbackProperties = map[string]func(cb *types.Callback, value string){
	"name":    callbackName,
	"package": callbackPackage,
}
var callbackPropertyNames = []string{}

func callbackName(cb *types.Callback, value string) {
	b := cb.Basic()
	b.Def = value
	cb.SetBasic(b)
}

func callbackPackage(cb *types.Callback, value string) {
	b := cb.Basic()
	b.Package = value
	cb.SetBasic(b)
}

var dictionaryProperties = map[string]func(cb *types.Dictionary, value string){
	"name":    dictionaryName,
	"package": dictionaryPackage,
}
var dictionaryPropertyNames = []string{}

func dictionaryName(cb *types.Dictionary, value string) {
	b := cb.Basic()
	b.Def = value
	cb.SetBasic(b)
}

func dictionaryPackage(cb *types.Dictionary, value string) {
	b := cb.Basic()
	b.Package = value
	cb.SetBasic(b)
}

var enumProperties = map[string]func(cb *types.Enum, value string){
	"name":    enumName,
	"package": enumPackage,
	"prefix":  enumPrefix,
	"suffix":  enumSuffix,
}
var enumPropertyNames = []string{}

func enumName(cb *types.Enum, value string) {
	b := cb.Basic()
	b.Def = value
	cb.SetBasic(b)
}

func enumPackage(cb *types.Enum, value string) {
	b := cb.Basic()
	b.Package = value
	cb.SetBasic(b)
}

func enumPrefix(cb *types.Enum, value string) {
	cb.Prefix = value
}

func enumSuffix(cb *types.Enum, value string) {
	cb.Suffix = value
}

var interfaceProperties = map[string]func(inf *types.Interface, value string){
	"constPrefix":     interfaceConstPrefix,
	"constSuffix":     interfaceConstSuffix,
	"constructorName": interfaceConstructorName,
	"name":            interfaceName,
	"package":         interfacePackage,
}
var interfacePropertyNames = []string{}

func interfaceConstructorName(inf *types.Interface, value string) {
	if inf.Constructor == nil {
		// TODO add failure here
		return
	}
	name := inf.Constructor.Name()
	name.Def = value
}

func interfaceConstPrefix(inf *types.Interface, value string) {
	inf.ConstPrefix = value
}

func interfaceConstSuffix(inf *types.Interface, value string) {
	inf.ConstSuffix = value
}

func interfaceName(inf *types.Interface, value string) {
	b := inf.Basic()
	b.Def = value
	inf.SetBasic(b)
}

func interfacePackage(inf *types.Interface, value string) {
	b := inf.Basic()
	b.Package = value
	inf.SetBasic(b)
}

func init() {
	for k := range callbackProperties {
		callbackPropertyNames = append(callbackPropertyNames, k)
	}
	sort.Strings(callbackPropertyNames)
	for k := range dictionaryProperties {
		dictionaryPropertyNames = append(dictionaryPropertyNames, k)
	}
	sort.Strings(dictionaryPropertyNames)
	for k := range enumProperties {
		enumPropertyNames = append(enumPropertyNames, k)
	}
	sort.Strings(enumPropertyNames)
	for k := range globalProperties {
		globalPropertyNames = append(globalPropertyNames, k)
	}
	sort.Strings(globalPropertyNames)
	for k := range interfaceProperties {
		callbackPropertyNames = append(interfacePropertyNames, k)
	}
	sort.Strings(interfacePropertyNames)
}
