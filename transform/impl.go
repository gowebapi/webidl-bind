package transform

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gowebapi/webidl-bind/types"
)

var globalProperties = map[string]bool{
	"package": true,
}
var globalPropertyNames = []string{}

var fileProperties = map[string]func(spec *SpecStatus, value string) string{
	"comment": fileComment,
	"title":   fileTitle,
	"url":     fileUrl,
}
var filePropertyNames = []string{}

func fileComment(spec *SpecStatus, value string) string {
	if spec.Comment != "" {
		return fmt.Sprintf("title is already defined '%s'", spec.Comment)
	}
	spec.Comment = value
	return ""
}

func fileTitle(spec *SpecStatus, value string) string {
	if spec.Title != "" {
		return fmt.Sprintf("title is already defined '%s'", spec.Title)
	}
	spec.Title = value
	return ""
}

func fileUrl(spec *SpecStatus, value string) string {
	if spec.Url != "" {
		return fmt.Sprintf("url  is already defined '%s'", spec.Url)
	}
	if strings.HasPrefix(value, "<") && strings.HasSuffix(value, ">") {
		value = value[1 : len(value)-1]
	}
	spec.Url = value
	return ""
}

var callbackProperties = map[string]func(cb *types.Callback, value string) string{
	"name":    callbackName,
	"package": callbackPackage,
}
var callbackPropertyNames = []string{}

func callbackName(cb *types.Callback, value string) string {
	b := cb.Basic()
	b.Def = value
	cb.SetBasic(b)
	return ""
}

func callbackPackage(cb *types.Callback, value string) string {
	msg := verifyPackageName(value)
	b := cb.Basic()
	b.Package = value
	cb.SetBasic(b)
	return msg
}

var dictionaryProperties = map[string]func(cb *types.Dictionary, value string) string{
	"name":    dictionaryName,
	"package": dictionaryPackage,
}
var dictionaryPropertyNames = []string{}

func dictionaryName(cb *types.Dictionary, value string) string {
	b := cb.Basic()
	b.Def = value
	cb.SetBasic(b)
	return ""
}

func dictionaryPackage(cb *types.Dictionary, value string) string {
	msg := verifyPackageName(value)
	b := cb.Basic()
	b.Package = value
	cb.SetBasic(b)
	return msg
}

var enumProperties = map[string]func(cb *types.Enum, value string) string{
	"name":    enumName,
	"package": enumPackage,
	"prefix":  enumPrefix,
	"suffix":  enumSuffix,
}
var enumPropertyNames = []string{}

func enumName(cb *types.Enum, value string) string {
	b := cb.Basic()
	b.Def = value
	cb.SetBasic(b)
	return ""
}

func enumPackage(cb *types.Enum, value string) string {
	msg := verifyPackageName(value)
	b := cb.Basic()
	b.Package = value
	cb.SetBasic(b)
	return msg
}

func enumPrefix(cb *types.Enum, value string) string {
	cb.Prefix = value
	return ""
}

func enumSuffix(cb *types.Enum, value string) string {
	cb.Suffix = value
	return ""
}

var interfaceProperties = map[string]func(inf *types.Interface, value string) string{
	"constPrefix":     interfaceConstPrefix,
	"constSuffix":     interfaceConstSuffix,
	"constructorName": interfaceConstructorName,
	"name":            interfaceName,
	"package":         interfacePackage,
}
var interfacePropertyNames = []string{}

func interfaceConstructorName(inf *types.Interface, value string) string {
	if inf.Constructor == nil {
		return "interface doesn't have any constructor"
	}
	name := inf.Constructor.Name()
	name.Def = value
	return ""
}

func interfaceConstPrefix(inf *types.Interface, value string) string {
	inf.ConstPrefix = value
	return ""
}

func interfaceConstSuffix(inf *types.Interface, value string) string {
	inf.ConstSuffix = value
	return ""
}

func interfaceName(inf *types.Interface, value string) string {
	b := inf.Basic()
	b.Def = value
	inf.SetBasic(b)
	return ""
}

func interfacePackage(inf *types.Interface, value string) string {
	msg := verifyPackageName(value)
	b := inf.Basic()
	b.Package = value
	inf.SetBasic(b)
	return msg
}

func verifyPackageName(value string) string {
	if strings.HasSuffix(value, "/") {
		return "invalid package name"
	}
	return ""
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
	for k := range fileProperties {
		filePropertyNames = append(filePropertyNames, k)
	}
	sort.Strings(filePropertyNames)
}
