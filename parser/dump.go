package parser

import (
	"bytes"
	"io"

	"github.com/gowebapi/webidl-bind/ast"
	"github.com/kr/pretty"
)

func Dump(w io.Writer, n ast.Node) error {
	_, err := pretty.Fprintf(w, "%# v", n)
	return err
}

func DumpString(n ast.Node) string {
	buf := bytes.NewBuffer(nil)
	if err := Dump(buf, n); err != nil {
		panic(err)
	}
	return buf.String()
}
