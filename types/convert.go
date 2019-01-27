package types

import (
	"errors"
	"fmt"
	"os"

	"github.com/dennwc/webidl/ast"
	"github.com/dennwc/webidl/parser"
)

var (
	StopErr = errors.New("stopping for previous error")
)

type Name struct {
	// Idl name
	Idl string

	// Package name
	Package string

	// Def is short for definition of a type, e.g. Foo
	Def string

	// InOut is method input and output variable type, e.g. *Foo
	InOut string

	// Internal name for used with methods and need to write some code
	Internal string

	// Pointer is true if InOut is a pointer type
	Pointer bool 
}

type Type interface {
	ast.Node
	Name() Name

	GetAllTypeRefs(list []TypeRef) []TypeRef
	// 	Phase1(r *Resources)
	// 	Phase2()

	// CalculateTypeDef()

	// 	// static attributes etc in interfaces
	// 	ExtractSubTypes() []Type
	// 	Phase3()
	// 	WriteTo() error

	NeedRelease() bool
}

type Convert struct {
	Types       map[string]Type
	All         []Type
	Enums       []*Enum
	Callbacks   []*Callback
	Dictionary  []*Dictionary
	partialDict []*Dictionary
	Interface   []*Interface
	partialIf   []*Interface

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
		return StopErr
	}
	return nil
}

// EvaluateInput is doing verification on input IDL according
// to WebIDL specification
func (conv *Convert) EvaluateInput() error {
	if conv.evaluateDictonary(); conv.HaveError {
		return StopErr
	}
	if conv.evaluateTypeRef(); conv.HaveError {
		return StopErr
	}
	return nil
}

func (conv *Convert) evaluateTypeRef() {
	typerefs := make([]TypeRef, 0)
	for _, t := range conv.All {
		typerefs = t.GetAllTypeRefs(typerefs)
	}
	for _, t := range typerefs {
		t.link(conv)
	}
}

func (conv *Convert) evaluateDictonary() {
	for _, pd := range conv.partialDict {
		conv.failing(pd, "partial dictonaries is not supported")
	}
	for _, pd := range conv.partialIf {
		conv.failing(pd, "partial interface is not supported")
	}
	// exapand partial
	// sort members
	// evaluate inherits?

}

// EvaluateOutput is doing evaluation of output that might be needed
//for the output language
func (t *Convert) EvaluateOutput() error {
	if t.HaveError {
		return StopErr
	}
	return nil
}

func (t *Convert) add(v Type) {
	if v == nil {
		return
	}
	name := v.Name().Idl
	if _, f := t.Types[name]; f {
		t.failing(v, "type '%s' already exist", name)
		return
	}
	t.Types[name] = v
	t.All = append(t.All, v)
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
	fmt.Println("evaluate mixim")
	parser.Dump(os.Stdout, value)
	panic("todo")
	// return false
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
	fmt.Println("evaluate includes")
	parser.Dump(os.Stdout, value)
	panic("todo")
}

func (t *extractTypes) Callback(value *ast.Callback) bool {
	// fmt.Println("evaluate callback")
	cb := t.convertCallback(value)
	t.main.Callbacks = append(t.main.Callbacks, cb)
	t.main.add(cb)
	return false
}

func (t *extractTypes) Typedef(value *ast.Typedef) bool {
	fmt.Println("evaluate typedef")
	parser.Dump(os.Stdout, value)
	panic("todo")
	// return false
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
