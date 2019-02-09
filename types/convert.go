package types

import (
	"errors"
	"fmt"
	"os"

	"github.com/gowebapi/webidlparser/ast"
	"github.com/gowebapi/webidlparser/parser"
)

var (
	ErrStop = errors.New("too many errors")
)

type Type interface {
	TypeRef
	ast.Node

	// 	Phase1(r *Resources)
	// 	Phase2()

	// CalculateTypeDef()

	// 	// static attributes etc in interfaces
	// 	ExtractSubTypes() []Type
	// 	Phase3()
	// 	WriteTo() error

	// key to use in Convert.Types
	key() string

	// Used indicate that it's a interface or a type that is used by
	// an interface
	InUse() bool

	SetBasic(basic BasicInfo)
}

type Convert struct {
	Types        map[string]Type
	All          []Type
	Enums        []*Enum
	Callbacks    []*Callback
	Dictionary   []*Dictionary
	partialDict  []*Dictionary
	Interface    []*Interface
	partialIf    []*Interface
	mixin        map[string]*mixin
	partialMixin []*mixin
	includes     []*includes
	Unions       []*UnionType

	HaveError bool
	setup     *Setup
}

type UserMsgFn func(base *ast.Base, format string, args ...interface{})

type Setup struct {
	Package        string
	Warning, Error UserMsgFn
}

func NewConvert() *Convert {
	return &Convert{
		Types: make(map[string]Type),
		Enums: []*Enum{},
		mixin: make(map[string]*mixin),
	}
}

func (t *Convert) Process(file *ast.File, setup *Setup) error {
	t.setup = setup
	// * process the file
	// * create a type map

	// 1 phase:
	// * do basic validation, e.g. enum annotation

	// 2 phase:
	// * nest type expansion

	list := extractTypes{main: t}
	ast.Accept(file, &list)
	if t.HaveError {
		return ErrStop
	}
	return nil
}

// EvaluateInput is doing verification on input IDL according
// to WebIDL specification
func (conv *Convert) EvaluateInput() error {
	if conv.processPartialAndMixin(); conv.HaveError {
		return ErrStop
	}
	if conv.processTypeLinks(); conv.HaveError {
		return ErrStop
	}
	if conv.verifyIndividualTypeCheck(); conv.HaveError {
		return ErrStop
	}
	return nil
}

// processTypeLinks is evaluating all types used
// by interfaces
func (conv *Convert) processTypeLinks() {
	for _, q := range conv.Interface {
		q.link(conv, make(inuseLogic))
	}
}

func (conv *Convert) processPartialAndMixin() {
	for _, pd := range conv.partialDict {
		if candidate, f := conv.Types[pd.key()]; f {
			if parent, ok := candidate.(*Dictionary); ok {
				parent.merge(pd, conv)
			} else {
				conv.failing(pd, "trying to add partial dictionary to a non-dictionary type (%T)", candidate)
			}
		} else {
			conv.failing(pd, "directory '%s' doesn't exist", pd.key())
		}
	}
	for _, pd := range conv.partialIf {
		if candidate, f := conv.Types[pd.key()]; f {
			if parent, ok := candidate.(*Interface); ok {
				parent.merge(pd, conv)
			} else {
				conv.failing(pd, "trying to add partial interface to a non-interface type (%T)", candidate)
			}
		} else {
			conv.failing(pd, "interface '%s' doesn't exist", pd.key())
		}
	}
	for _, pd := range conv.partialMixin {
		if parent, f := conv.mixin[pd.Name]; f {
			parent.merge(pd, conv)
		} else {
			conv.failing(pd, "mixin '%s' doesn't exist", pd.Name)
		}
	}
	for _, inc := range conv.includes {
		target, found := conv.Types[inc.Name]
		if !found {
			conv.failing(inc, "include refernce to '%s' that doesn't exist", inc.Name)
			continue
		}
		src, found := conv.mixin[inc.Source]
		if !found {
			conv.failing(inc, "include is references to '%s' that doesn't exist (or is not an mixin)", inc.Source)
			continue
		}
		if doc, ok := target.(*Interface); ok {
			doc.mergeMixin(src, conv)
		} else {
			conv.failing(inc, "target include existed to be an interface, not %T", target)
		}
	}
	// exapand partial
	// sort members
	// evaluate inherits?

}

// EvaluateOutput is doing evaluation of output that might be needed
//for the output language
func (t *Convert) EvaluateOutput() error {
	if t.HaveError {
		return ErrStop
	}
	return nil
}

func (t *Convert) add(v Type) {
	if v == nil {
		return
	}
	name := v.key()
	t.registerTypeName(v, name)
	t.Types[name] = v
	t.All = append(t.All, v)
}

func (t *Convert) addMixin(m *mixin) {
	if m == nil {
		return
	}
	t.registerTypeName(m, m.Name)
	t.mixin[m.Name] = m
}

func (t *Convert) registerTypeName(ref ast.Node, name string) {
	if _, f := t.Types[name]; f {
		t.failing(ref, "type '%s' already exist", name)
		return
	}
	if _, f := t.mixin[name]; f {
		t.failing(ref, "type '%s' already exist.", name)
		return
	}
}

func (t *Convert) failing(base ast.Node, format string, args ...interface{}) {
	t.setup.Error(base.NodeBase(), format, args...)
	t.HaveError = true
}

func (t *Convert) warning(base ast.Node, format string, args ...interface{}) {
	t.setup.Warning(base.NodeBase(), format, args...)
}

func (t *Convert) assertTrue(test bool, node ast.Node, format string, args ...interface{}) {
	if !test {
		t.failing(node, format, args...)
	}
}

func (t *Convert) warningTrue(test bool, node ast.Node, format string, args ...interface{}) {
	if !test {
		t.warning(node, format, args...)
	}
}

type extractTypes struct {
	ast.EmptyVisitor
	main *Convert
}

func (t *extractTypes) Enum(value *ast.Enum) bool {
	// fmt.Println("evaluate enum ")
	next := t.convertEnum(value)
	t.main.Enums = append(t.main.Enums, next)
	t.main.add(next)
	return false
}

func (t *extractTypes) Interface(value *ast.Interface) bool {
	// fmt.Println("evaluate interface")
	next, partial := t.convertInterface(value)
	if partial {
		t.main.partialIf = append(t.main.partialIf, next)
	} else {
		t.main.Interface = append(t.main.Interface, next)
		t.main.add(next)
	}
	return false
}

func (t *extractTypes) Mixin(value *ast.Mixin) bool {
	// fmt.Println("evaluate mixim")
	next, partial := t.convertMixin(value)
	if partial {
		t.main.partialMixin = append(t.main.partialMixin, next)
	} else {
		t.main.addMixin(next)
	}
	return false
}

func (t *extractTypes) Dictionary(value *ast.Dictionary) bool {
	// fmt.Println("evaluate dirctionary")
	next, partial := t.convertDictionary(value)
	if partial {
		t.main.partialDict = append(t.main.partialDict, next)
	} else {
		t.main.Dictionary = append(t.main.Dictionary, next)
		t.main.add(next)
	}
	return false
}

func (t *extractTypes) Implementation(value *ast.Implementation) {
	fmt.Println("evaluate implementation")
	parser.Dump(os.Stdout, value)
	panic("todo")
}

func (t *extractTypes) Includes(value *ast.Includes) {
	// fmt.Println("evaluate includes")
	next := t.convertIncludes(value)
	t.main.includes = append(t.main.includes, next)
}

func (t *extractTypes) Callback(value *ast.Callback) bool {
	// fmt.Println("evaluate callback")
	cb := t.convertCallback(value)
	t.main.Callbacks = append(t.main.Callbacks, cb)
	t.main.add(cb)
	return false
}

func (t *extractTypes) Typedef(value *ast.Typedef) bool {
	// fmt.Println("evaluate typedef")
	next := t.convertTypeDef(value)
	t.main.add(next)
	return false
}

func (t *extractTypes) failing(node ast.Node, format string, args ...interface{}) {
	t.main.failing(node, format, args...)
}

func (t *extractTypes) warning(node ast.Node, format string, args ...interface{}) {
	t.main.warning(node, format, args...)
}

func (t *extractTypes) assertTrue(test bool, node ast.Node, format string, args ...interface{}) {
	t.main.assertTrue(test, node, format, args...)
}

func (t *extractTypes) warningTrue(test bool, node ast.Node, format string, args ...interface{}) {
	t.main.warningTrue(test, node, format, args...)
}
