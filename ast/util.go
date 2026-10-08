package ast

type getAllErrors struct {
	EmptyVisitor
	out []*ErrorNode
}

func GetAllErrorNodes(start Node) []*ErrorNode {
	data := getAllErrors{}
	if start != nil {
		Accept(start, &data)
	}
	return data.out
}

func (t *getAllErrors) Base(value *Base) {
	if value.Errors != nil {
		t.out = append(t.out, value.Errors...)
	}
}
