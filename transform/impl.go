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

type fileProperty interface {
	Set(spec *SpecStatus, value string) string
}

var fileProperties = map[string]fileProperty{
	"comment": &fileComment{},
	"title":   &fileTitle{},
	"url":     &fileUrl{},
}
var filePropertyNames = []string{}

type fileComment struct{}

func (fc *fileComment) Set(spec *SpecStatus, value string) string {
	if spec.Comment != "" {
		return fmt.Sprintf("title is already defined '%s'", spec.Comment)
	}
	spec.Comment = value
	return ""
}

type fileTitle struct{}

func (t *fileTitle) Set(spec *SpecStatus, value string) string {
	if spec.Title != "" {
		return fmt.Sprintf("title is already defined '%s'", spec.Title)
	}
	spec.Title = value
	return ""
}

type fileUrl struct{}

func (t *fileUrl) Set(spec *SpecStatus, value string) string {
	if spec.Url != "" {
		return fmt.Sprintf("url  is already defined '%s'", spec.Url)
	}
	if strings.HasPrefix(value, "<") && strings.HasSuffix(value, ">") {
		value = value[1 : len(value)-1]
	}
	spec.Url = value
	return ""
}

type callbackProperty interface {
	Get(cb *types.Callback) string
	Set(cb *types.Callback, value string) string
}

var callbackProperties = map[string]callbackProperty{
	"name":    &callbackName{},
	"package": &callbackPackage{},
}
var callbackPropertyNames = []string{}

type callbackName struct{}

func (t *callbackName) Get(cb *types.Callback) string {
	return cb.Basic().Def
}

func (t *callbackName) Set(cb *types.Callback, value string) string {
	b := cb.Basic()
	b.Def = value
	cb.SetBasic(b)
	return ""
}

type callbackPackage struct{}

func (t *callbackPackage) Get(cb *types.Callback) string {
	return cb.Basic().Package
}

func (t *callbackPackage) Set(cb *types.Callback, value string) string {
	msg := verifyPackageName(value)
	b := cb.Basic()
	b.Package = value
	cb.SetBasic(b)
	return msg
}

type dictionaryProperty interface {
	Get(cb *types.Dictionary) string
	Set(cb *types.Dictionary, value string) string
}

var dictionaryProperties = map[string]dictionaryProperty{
	"name":    &dictionaryName{},
	"package": &dictionaryPackage{},
}
var dictionaryPropertyNames = []string{}

type dictionaryName struct{}

func (t *dictionaryName) Get(cb *types.Dictionary) string {
	return cb.Basic().Def
}

func (t *dictionaryName) Set(cb *types.Dictionary, value string) string {
	b := cb.Basic()
	b.Def = value
	cb.SetBasic(b)
	return ""
}

type dictionaryPackage struct{}

func (t *dictionaryPackage) Get(cb *types.Dictionary) string {
	return cb.Basic().Package
}

func (t *dictionaryPackage) Set(cb *types.Dictionary, value string) string {
	msg := verifyPackageName(value)
	b := cb.Basic()
	b.Package = value
	cb.SetBasic(b)
	return msg
}

type enumProperty interface {
	Get(cb *types.Enum) string
	Set(cb *types.Enum, value string) string
}

var enumProperties = map[string]enumProperty{
	"name":    &enumName{},
	"package": &enumPackage{},
	"prefix":  &enumPrefix{},
	"suffix":  &enumSuffix{},
}
var enumPropertyNames = []string{}

type enumName struct{}

func (t *enumName) Get(cb *types.Enum) string {
	return cb.Basic().Def
}

func (t *enumName) Set(cb *types.Enum, value string) string {
	b := cb.Basic()
	b.Def = value
	cb.SetBasic(b)
	return ""
}

type enumPackage struct{}

func (t *enumPackage) Get(cb *types.Enum) string {
	return cb.Basic().Package
}

func (t *enumPackage) Set(cb *types.Enum, value string) string {
	msg := verifyPackageName(value)
	b := cb.Basic()
	b.Package = value
	cb.SetBasic(b)
	return msg
}

type enumPrefix struct{}

func (t *enumPrefix) Get(cb *types.Enum) string {
	return cb.Prefix
}

func (t *enumPrefix) Set(cb *types.Enum, value string) string {
	cb.Prefix = value
	return ""
}

type enumSuffix struct{}

func (t *enumSuffix) Get(cb *types.Enum) string {
	return cb.Suffix
}

func (t *enumSuffix) Set(cb *types.Enum, value string) string {
	cb.Suffix = value
	return ""
}

type interfaceProperty interface {
	Get(cb *types.Interface) string
	Set(inf *types.Interface, value string) string
}

var interfaceProperties = map[string]interfaceProperty{
	"constPrefix":     &interfaceConstPrefix{},
	"constSuffix":     &interfaceConstSuffix{},
	"constructorName": &interfaceConstructorName{},
	"name":            &interfaceName{},
	"package":         &interfacePackage{},
}
var interfacePropertyNames = []string{}

type interfaceConstructorName struct{}

func (t *interfaceConstructorName) Get(inf *types.Interface) string {
	return inf.Constructor.Name().Def
}

func (t *interfaceConstructorName) Set(inf *types.Interface, value string) string {
	if inf.Constructor == nil {
		return "interface doesn't have any constructor"
	}
	name := inf.Constructor.Name()
	name.Def = value
	return ""
}

type interfaceConstPrefix struct{}

func (t *interfaceConstPrefix) Get(inf *types.Interface) string {
	return inf.ConstPrefix
}

func (t *interfaceConstPrefix) Set(inf *types.Interface, value string) string {
	inf.ConstPrefix = value
	return ""
}

type interfaceConstSuffix struct{}

func (t *interfaceConstSuffix) Get(inf *types.Interface) string {
	return inf.ConstSuffix
}

func (t *interfaceConstSuffix) Set(inf *types.Interface, value string) string {
	inf.ConstSuffix = value
	return ""
}

type interfaceName struct{}

func (t *interfaceName) Get(inf *types.Interface) string {
	return inf.Basic().Def
}

func (t *interfaceName) Set(inf *types.Interface, value string) string {
	b := inf.Basic()
	b.Def = value
	inf.SetBasic(b)
	return ""
}

type interfacePackage struct{}

func (t *interfacePackage) Get(inf *types.Interface) string {
	return inf.Basic().Package
}

func (t *interfacePackage) Set(inf *types.Interface, value string) string {
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
