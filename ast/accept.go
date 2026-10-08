package ast

import "fmt"

type Visitor interface {
	Base(base *Base)

	ErrorNode(value *ErrorNode)
	File(value *File) bool
	Interface(value *Interface) bool
	Mixin(value *Mixin) bool
	Dictionary(value *Dictionary) bool
	Annotation(value *Annotation) bool
	Parameter(value *Parameter) bool
	Implementation(value *Implementation)
	Includes(value *Includes)
	Member(value *Member) bool
	CustomOp(value *CustomOp)
	TypeName(value *TypeName)
	Pattern(value *Pattern)
	Callback(value *Callback) bool
	Enum(value *Enum) bool
	Typedef(value *Typedef) bool
	AnyType(value *AnyType)
	SequenceType(value *SequenceType) bool
	RecordType(value *RecordType) bool
	ParametrizedType(value *ParametrizedType) bool
	UnionType(value *UnionType) bool
	NullableType(value *NullableType) bool
	BasicLiteral(value *BasicLiteral)
	SequenceLiteral(value *SequenceLiteral) bool
}

func Accept(node Node, v Visitor) {
	if node == nil {
		return
	}
	v.Base(node.NodeBase())
	switch n := node.(type) {
	case *ErrorNode:
		v.ErrorNode(n)
	case *File:
		if v.File(n) {
			for _, c := range n.Declarations {
				Accept(c, v)
			}
		}
	case *Interface:
		if !v.Interface(n) {
			break
		}
		for _, a := range n.Annotations {
			Accept(a, v)
		}
		for _, m := range n.Members {
			AcceptInterfaceMember(m, v)
		}
		for _, c := range n.CustomOps {
			Accept(c, v)
		}
		for _, p := range n.Patterns {
			Accept(p, v)
		}
	case *Mixin:
		if !v.Mixin(n) {
			break
		}
		for _, a := range n.Annotations {
			Accept(a, v)
		}
		for _, m := range n.Members {
			AcceptMixinMember(m, v)
		}
		for _, c := range n.CustomOps {
			Accept(c, v)
		}
		for _, p := range n.Patterns {
			Accept(p, v)
		}
	case *Dictionary:
		if !v.Dictionary(n) {
			break
		}
		for _, a := range n.Annotations {
			Accept(a, v)
		}
		for _, m := range n.Members {
			AcceptInterfaceMember(m, v)
		}
	case *Annotation:
		if !v.Annotation(n) {
			break
		}
		for _, p := range n.Parameters {
			Accept(p, v)
		}
	case *Parameter:
		if !v.Parameter(n) {
			break
		}
		Accept(n.Type, v)
		AcceptLiteral(n.Init, v)
		for _, a := range n.Annotations {
			v.Annotation(a)
		}
	case *Implementation:
		v.Implementation(n)
	case *Includes:
		v.Includes(n)
	case *CustomOp:
		v.CustomOp(n)
	case *Pattern:
		v.Pattern(n)
	case *Callback:
		if !v.Callback(n) {
			break
		}
		Accept(n.Return, v)
		for _, p := range n.Parameters {
			Accept(p, v)
		}
	case *Enum:
		if !v.Enum(n) {
			break
		}
		for _, a := range n.Annotations {
			Accept(a, v)
		}
		for _, q := range n.Values {
			AcceptLiteral(q, v)
		}
	case *Typedef:
		if !v.Typedef(n) {
			break
		}
		Accept(n.Type, v)
		for _, a := range n.Annotations {
			v.Annotation(a)
		}
	case *TypeName:
		v.TypeName(n)
	case *AnyType:
		v.AnyType(n)
	case *SequenceType:
		if !v.SequenceType(n) {
			break
		}
		Accept(n.Elem, v)
	case *RecordType:
		if !v.RecordType(n) {
			break
		}
		Accept(n.Key, v)
		Accept(n.Elem, v)
	case *ParametrizedType:
		if !v.ParametrizedType(n) {
			break
		}
		for _, e := range n.Elems {
			Accept(e, v)
		}
	case *UnionType:
		if !v.UnionType(n) {
			break
		}
		for _, t := range n.Types {
			Accept(t, v)
		}
	case *NullableType:
		if !v.NullableType(n) {
			break
		}
		Accept(n.Type, v)
	default:
		unknownTypeError(node)
	}
}

func AcceptInterfaceMember(m InterfaceMember, v Visitor) {
	if m == nil {
		return
	}
	switch m := m.(type) {
	case *Member:
		acceptMember(m, v)
	default:
		unknownTypeError(m)
	}
}

func AcceptMixinMember(m MixinMember, v Visitor) {
	if m == nil {
		return
	}
	switch m := m.(type) {
	case *Member:
		acceptMember(m, v)
	default:
		unknownTypeError(m)
	}
}

func acceptMember(m *Member, v Visitor) {
	v.Base(m.NodeBase())
	if !v.Member(m) {
		return
	}
	Accept(m.Type, v)
	AcceptLiteral(m.Init, v)
	for _, p := range m.Parameters {
		Accept(p, v)
	}
	for _, a := range m.Annotations {
		Accept(a, v)
	}
}

func unknownTypeError(value any) {
	msg := fmt.Sprintf("unknown type %T", value)
	panic(msg)
}

func AcceptLiteral(in Literal, v Visitor) {
	if in == nil {
		return
	}
	switch n := in.(type) {
	case *BasicLiteral:
		v.Base(&n.Base)
		v.BasicLiteral(n)
	case *SequenceLiteral:
		v.Base(&n.Base)
		if !v.SequenceLiteral(n) {
			break
		}
		for _, e := range n.Elems {
			AcceptLiteral(e, v)
		}
	default:
		unknownTypeError(in)
	}
}

// EmptyVisitor implement a default Visitor that is doing nothing
type EmptyVisitor struct {
	UseFlags             bool
	ScanBase             bool
	ScanErrorNode        bool
	ScanFile             bool
	ScanInterface        bool
	ScanMixin            bool
	ScanDictionary       bool
	ScanAnnotation       bool
	ScanParameter        bool
	ScanImplementation   bool
	ScanIncludes         bool
	ScanMember           bool
	ScanCustomOp         bool
	ScanTypeName         bool
	ScanPattern          bool
	ScanCallback         bool
	ScanEnum             bool
	ScanTypedef          bool
	ScanAnyType          bool
	ScanSequenceType     bool
	ScanRecordType       bool
	ScanParametrizedType bool
	ScanUnionType        bool
	ScanNullableType     bool
	ScanBasicLiteral     bool
	ScanSequenceLiteral  bool
}

func (t *EmptyVisitor) Base(base *Base) {
	// return !t.UseFlags || (t.UseFlags && t.ScanBase)
}

func (t *EmptyVisitor) ErrorNode(value *ErrorNode) {
	// return !t.UseFlags || (t.UseFlags && t.ScanErrorNode)
}

func (t *EmptyVisitor) File(value *File) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanFile)
}

func (t *EmptyVisitor) Interface(value *Interface) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanInterface)
}

func (t *EmptyVisitor) Mixin(value *Mixin) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanMixin)
}

func (t *EmptyVisitor) Dictionary(value *Dictionary) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanDictionary)
}

func (t *EmptyVisitor) Annotation(value *Annotation) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanAnnotation)
}

func (t *EmptyVisitor) Parameter(value *Parameter) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanParameter)
}

func (t *EmptyVisitor) Implementation(value *Implementation) {
	// return !t.UseFlags || (t.UseFlags && t.ScanImplementation)
}

func (t *EmptyVisitor) Includes(value *Includes) {
	// return !t.UseFlags || (t.UseFlags && t.ScanIncludes)
}

func (t *EmptyVisitor) Member(value *Member) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanMember)
}

func (t *EmptyVisitor) CustomOp(value *CustomOp) {
	// return !t.UseFlags || (t.UseFlags && t.ScanCustomOp)
}

func (t *EmptyVisitor) TypeName(value *TypeName) {
	// return !t.UseFlags || (t.UseFlags && t.ScanTypeName)
}

func (t *EmptyVisitor) Pattern(value *Pattern) {
	// return !t.UseFlags || (t.UseFlags && t.ScanPattern)
}

func (t *EmptyVisitor) Callback(value *Callback) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanCallback)
}

func (t *EmptyVisitor) Enum(value *Enum) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanEnum)
}

func (t *EmptyVisitor) Typedef(value *Typedef) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanTypedef)
}

func (t *EmptyVisitor) AnyType(value *AnyType) {
	// return !t.UseFlags || (t.UseFlags && t.ScanAnyType)
}

func (t *EmptyVisitor) SequenceType(value *SequenceType) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanSequenceType)
}

func (t *EmptyVisitor) RecordType(value *RecordType) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanRecordType)
}

func (t *EmptyVisitor) ParametrizedType(value *ParametrizedType) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanParametrizedType)
}

func (t *EmptyVisitor) UnionType(value *UnionType) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanUnionType)
}

func (t *EmptyVisitor) NullableType(value *NullableType) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanNullableType)
}

func (t *EmptyVisitor) BasicLiteral(value *BasicLiteral) {
	// return !t.UseFlags || (t.UseFlags && t.ScanBasicLiteral)
}

func (t *EmptyVisitor) SequenceLiteral(value *SequenceLiteral) bool {
	return !t.UseFlags || (t.UseFlags && t.ScanSequenceLiteral)
}
