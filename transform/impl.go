package transform

import (
	"sort"

	"github.com/gowebapi/webidlgenerator/types"
)

var callbackProperties = map[string]func(cb *types.Callback, value string){
	"name": callbackName,
}
var callbackPropertyNames = []string{}

func callbackName(cb *types.Callback, value string) {
	b := cb.Basic()
	b.Def = value
	cb.SetBasic(b)
}

var interfaceProperties = map[string]func(inf *types.Interface, value string){
	"name":        interfaceName,
	"constPrefix": interfaceConstPrefix,
	"constSuffix": interfaceConstSuffix,
}
var interfacePropertyNames = []string{}

func interfaceName(inf *types.Interface, value string) {
	b := inf.Basic()
	b.Def = value
	inf.SetBasic(b)
}

func interfaceConstPrefix(inf *types.Interface, value string) {
	inf.ConstPrefix = value
}

func interfaceConstSuffix(inf *types.Interface, value string) {
	inf.ConstSuffix = value
}

func init() {
	for k := range callbackProperties {
		callbackPropertyNames = append(callbackPropertyNames, k)
	}
	sort.Strings(callbackPropertyNames)
	for k := range interfaceProperties {
		callbackPropertyNames = append(interfacePropertyNames, k)
	}
	sort.Strings(interfacePropertyNames)
}
