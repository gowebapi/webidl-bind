package types

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/gowebapi/webidlparser/ast"
	"github.com/gowebapi/webidlparser/parser"
)

var (
	ErrStop = errors.New("too many errors")
)

// handle this as global variables is not ideal. However,
// the usage of Basic type inside all templates make it
// a lot of work. The issue will highlight it self if
// we are using multple go routines.

// TransformBasic is used to make modification to BasicInfo
// structure when it's returned from all types. Main usage
// is to change package name references when query for a type.
var TransformBasic = func(t TypeRef, basic BasicInfo) BasicInfo {
	return basic
}

type Type interface {
	TypeRef
	GetRef

	// key to use in Convert.Types
	key() string

	// Used indicate that it's a interface or a type that is used by
	// an interface
	InUse() bool

	SetBasic(basic BasicInfo)

	TypeID() TypeID
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

type UserMsgFn func(ref GetRef, format string, args ...interface{})

type Setup struct {
	Package        string
	Filename       string
	Warning, Error UserMsgFn
}

type TypeID int

const (
	TypeInterface TypeID = iota
	TypeEnum
	TypeCallback
	TypeDictionary
	TypeTypeDef
)

func (id TypeID) IsPublic() bool {
	return id != TypeTypeDef
}

func NewConvert() *Convert {
	return &Convert{
		Types: make(map[string]Type),
		Enums: []*Enum{},
		mixin: make(map[string]*mixin),
	}
}

// Load is reading a file from disc and Process it
func (t *Convert) Load(setup *Setup) error {
	content, err := ioutil.ReadFile(setup.Filename)
	if err != nil {
		return err
	}
	return t.Parse(content, setup)
}

// ParseTest is parsing a text and Process it
func (t *Convert) Parse(content []byte, setup *Setup) error {
	t.setup = setup
	file := parser.Parse(string(content))
	trouble := ast.GetAllErrorNodes(file)
	if len(trouble) > 0 {
		sort.SliceStable(trouble, func(i, j int) bool { return trouble[i].Line < trouble[j].Line })
		for _, e := range trouble {
			ref := Ref{Filename: setup.Filename, Line: e.Line}
			t.failing(&ref, e.Message)
		}
		return ErrStop
	}
	list := extractTypes{main: t}
	ast.Accept(file, &list)
	if t.HaveError {
		return ErrStop
	}
	return nil
}

// Evaluate is doing verification on input IDL according
// to WebIDL specification. It also expand types, like removing
// typedef etc.
func (conv *Convert) Evaluate() error {
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
			conv.failing(pd, "dictionary '%s' doesn't exist", pd.key())
		}
	}
	for _, pd := range conv.partialIf {
		if candidate, f := conv.Types[pd.key()]; f {
			if parent, ok := candidate.(*Interface); ok {
				parent.merge(pd, conv)
			} else {
				conv.failing(pd, "trying to add partial interface to a non-interface type (%T)", candidate)
			}
		} else if mixin, f := conv.mixin[pd.key()]; f {
			mixin.mergeIf(pd, conv)
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

func (t *Convert) registerTypeName(ref GetRef, name string) {
	if other, f := t.Types[name]; f {
		t.failing(ref, "type '%s' already exist at %s", name, other.SourceReference())
		return
	}
	if other, f := t.mixin[name]; f {
		t.failing(ref, "type '%s' already exist at %s", name, other.SourceReference())
		return
	}
}

// Sort is sorting the Enum, Callbacks, Dictionary and Interface.
func (t *Convert) Sort() {
	sort.Slice(t.Enums, func(i, j int) bool {
		return t.Enums[i].lessThan(t.Enums[j])
	})
	sort.Slice(t.Callbacks, func(i, j int) bool {
		return t.Callbacks[i].lessThan(t.Callbacks[j])
	})
	sort.Slice(t.Dictionary, func(i, j int) bool {
		return t.Dictionary[i].lessThan(t.Dictionary[j])
	})
	sort.Slice(t.Interface, func(i, j int) bool {
		return t.Interface[i].lessThan(t.Interface[j])
	})
	sort.Slice(t.Unions, func(i, j int) bool {
		return t.Unions[i].lessThan(t.Unions[j])
	})
}

func (t *Convert) failing(ref GetRef, format string, args ...interface{}) {
	t.setup.Error(ref, format, args...)
	t.HaveError = true
}

func (t *Convert) warning(ref GetRef, format string, args ...interface{}) {
	t.setup.Warning(ref, format, args...)
}

func (t *Convert) assertTrue(test bool, ref GetRef, format string, args ...interface{}) {
	if !test {
		t.failing(ref, format, args...)
	}
}

func (t *Convert) warningTrue(test bool, ref GetRef, format string, args ...interface{}) {
	if !test {
		t.warning(ref, format, args...)
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

func (t *extractTypes) failing(ref GetRef, format string, args ...interface{}) {
	t.main.failing(ref, format, args...)
}

func (t *extractTypes) warning(ref GetRef, format string, args ...interface{}) {
	t.main.warning(ref, format, args...)
}

func (t *extractTypes) assertTrue(test bool, ref GetRef, format string, args ...interface{}) {
	t.main.assertTrue(test, ref, format, args...)
}

func (t *extractTypes) warningTrue(test bool, ref GetRef, format string, args ...interface{}) {
	t.main.warningTrue(test, ref, format, args...)
}
