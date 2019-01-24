package types

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dennwc/webidl/ast"
	"github.com/dennwc/webidl/parser"
)

var (
	StopErr                  = errors.New("stopping for previous error")
	UnsupportedAnnotationErr = errors.New("unsupported annotations")
	UnsupportedLiteralErr    = errors.New("unsupported literal type")
)

type Name struct {
	Package            string
	Idl, Public, Local string
}

func fromIdlName(pkg string, name string) Name {
	if strings.HasPrefix(name, "_") && len(name) > 1 {
		name = name[1:]
	}
	return Name{
		Package: pkg,
		Idl:     name,
		Public:  toCamelCase(name, true),
		Local:   toCamelCase(name, false),
	}
}

type Type interface {
	Base() *ast.Base
	Name() Name

	GetAllTypeRefs(list []TypeRef) []TypeRef
	// 	Phase1(r *Resources)
	// 	Phase2()

	// CalculateTypeDef()

	// 	// static attributes etc in interfaces
	// 	ExtractSubTypes() []Type
	// 	Phase3()
	// 	WriteTo() error
}

type Convert struct {
	Types     map[string]Type
	All       []Type
	Enums     []*Enum
	Callbacks []*Callback

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
	typerefs := make([]TypeRef, 0)
	for _, t := range conv.All {
		typerefs = t.GetAllTypeRefs(typerefs)
	}
	for _, t := range typerefs {
		t.link(conv)
	}
	if conv.HaveError {
		return StopErr
	}
	return nil
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
	name := v.Name().Idl
	if _, f := t.Types[name]; f {
		t.failing(v.Base(), "type '%s' already exist", name, 1)
		return
	}
	t.Types[name] = v
	t.All = append(t.All, v)
}

func (t *Convert) failing(base *ast.Base, format string, args ...interface{}) {
	t.setup.Error(base, format, args)
	t.HaveError = true
}

func (t *Convert) warning(base *ast.Base, format string, args ...interface{}) {
	t.setup.Warning(base, format, args)
}

type extractTypes struct {
	ast.EmptyVisitor
	main *Convert
}

func (t *extractTypes) Enum(value *ast.Enum) bool {
	fmt.Println("evaluate enum ")
	next, err := ConvertEnum(value, t.main.setup)
	if err != nil {
		t.main.failing(value.NodeBase(), "enum trouble: %s", next)
		return false
	}
	t.main.Enums = append(t.main.Enums, next)
	t.main.add(next)
	return false
}

func (t *extractTypes) Interface(value *ast.Interface) bool {
	fmt.Println("evaluate interface")
	return false
}

func (t *extractTypes) Mixin(value *ast.Mixin) bool {
	fmt.Println("evaluate mixim")
	panic("todo")
	// return false
}

func (t *extractTypes) Dictionary(value *ast.Dictionary) bool {
	fmt.Println("evaluate dirctionary")
	return false
}

func (t *extractTypes) Implementation(value *ast.Implementation) {
	fmt.Println("evaluate implementation")
	panic("todo")
}

func (t *extractTypes) Includes(value *ast.Includes) {
	fmt.Println("evaluate includes")
	panic("todo")
}

func (t *extractTypes) Callback(value *ast.Callback) bool {
	fmt.Println("evaluate callback")
	cb, err := ConvertCallback(value, t.main.setup)
	if err != nil {
		t.main.failing(value.NodeBase(), "convert error: %s", err)
		return false
	}
	t.main.Callbacks = append(t.main.Callbacks, cb)
	t.main.add(cb)
	parser.Dump(os.Stdout, value)
	return false
}

func (t *extractTypes) Typedef(value *ast.Typedef) bool {
	fmt.Println("evaluate typedef")
	return false
}
