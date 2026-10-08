package ast

type Node interface {
	NodeBase() *Base
}

type Base struct {
	Start    int // rune
	End      int // rune
	Line     int // line number
	Comments []string
	Errors   []*ErrorNode
}

func (b *Base) NodeBase() *Base {
	return b
}

// error occurred; value is text of error
type ErrorNode struct {
	Base
	Message string
}

type Decl interface {
	Node
	isDecl()
}

// The file root node
type File struct {
	Base
	Declarations []Decl
}

// interface Foo { ... }
type Interface struct {
	Base
	Partial     bool
	Callback    bool
	Name        string
	Inherits    string
	Annotations []*Annotation
	Members     []InterfaceMember
	CustomOps   []*CustomOp
	Patterns    []*Pattern
}

func (*Interface) isDecl() {}

type InterfaceMember interface {
	isInterfaceMember()
}

// interface mixin Foo { ... }
type Mixin struct {
	Base
	Name        string
	Inherits    string
	Partial     bool
	Annotations []*Annotation
	Members     []MixinMember
	CustomOps   []*CustomOp
	Patterns    []*Pattern
}

func (*Mixin) isDecl() {}

type MixinMember interface {
	isMixinMember()
}

type Dictionary struct {
	Base
	Name        string
	Inherits    string
	Partial     bool
	Annotations []*Annotation
	Members     []*Member
}

func (*Dictionary) isDecl() {}

// [Constructor], []
type Annotation struct {
	Base
	Name       string
	Value      string       // [A=B]
	Parameters []*Parameter // [A(X x, Y y)]
	Values     []string     // [A=(a,b,c)]
}

// optional any SomeArg
type Parameter struct {
	Base
	Type        Type
	Optional    bool
	Variadic    bool
	Name        string
	Init        Literal
	Annotations []*Annotation
}

// Window implements ECMA262Globals
type Implementation struct {
	Base
	Name   string
	Source string
}

func (*Implementation) isDecl() {}

// Document includes DocumentOrShadowRoot
type Includes struct {
	Base
	Name   string
	Source string
}

func (*Includes) isDecl() {}

// readonly attribute something
type Member struct {
	Base
	Name           string
	Type           Type
	Init           Literal
	Attribute      bool
	Static         bool
	Const          bool
	Readonly       bool
	Required       bool
	Specialization string
	Parameters     []*Parameter
	Annotations    []*Annotation
}

func (*Member) isInterfaceMember() {}
func (*Member) isMixinMember()     {}

type CustomOp struct {
	Base
	Name string
}

type TypeName struct {
	Base
	Name string
}

func (*TypeName) isType() {}

type Pattern struct {
	Base
	Type     PatternType
	Key      Type
	Elem     Type
	ReadOnly bool
}

type PatternType int

const (
	Iterable PatternType = iota + 1
	AsyncIterable
	Maplike
	Setlike
)

type Callback struct {
	Base
	Name       string
	Return     Type
	Parameters []*Parameter
}

func (*Callback) isDecl() {}

type Enum struct {
	Base
	Annotations []*Annotation
	Name        string
	Values      []Literal
}

func (*Enum) isDecl() {}

type Typedef struct {
	Base
	Annotations []*Annotation
	Name        string
	Type        Type
}

func (*Typedef) isDecl() {}

type Type interface {
	Node
	isType()
}

type AnyType struct {
	Base
}

func (*AnyType) isType() {}

type SequenceType struct {
	Base
	Elem Type
}

func (*SequenceType) isType() {}

type RecordType struct {
	Base
	Key  Type
	Elem Type
}

func (*RecordType) isType() {}

type ParametrizedType struct {
	Base
	Name  string
	Elems []Type
}

func (*ParametrizedType) isType() {}

type UnionType struct {
	Base
	Types []Type
}

func (*UnionType) isType() {}

type NullableType struct {
	Base
	Type Type
}

func (*NullableType) isType() {}

type Literal interface {
	isLiteral()
}

type BasicLiteral struct {
	Base
	Value string
}

func (*BasicLiteral) isLiteral() {}

type SequenceLiteral struct {
	Base
	Elems []Literal
}

func (*SequenceLiteral) isLiteral() {}
