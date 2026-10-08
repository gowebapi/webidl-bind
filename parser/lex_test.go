// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Portions copied and modified from: https://github.com/golang/go/blob/master/src/text/template/parse/lex_test.go

package parser

import (
	"testing"
)

type lexerTest struct {
	name   string
	input  string
	tokens []lexeme
}

var (
	tEOF        = lexeme{tokenTypeEOF, 0, 0, ""}
	tWhitespace = lexeme{tokenTypeWhitespace, 0, 0, " "}
)

var lexerTests = []lexerTest{
	// Simple tests.
	{"empty", "", []lexeme{tEOF}},

	{"single whitespace", " ", []lexeme{tWhitespace, tEOF}},
	{"single tab", "\t", []lexeme{{tokenTypeWhitespace, 0, 0, "\t"}, tEOF}},
	{"multiple whitespace", "   ", []lexeme{tWhitespace, tWhitespace, tWhitespace, tEOF}},

	{"newline r", "\r", []lexeme{{tokenTypeWhitespace, 0, 0, "\r"}, tEOF}},
	{"newline n", "\n", []lexeme{{tokenTypeWhitespace, 0, 0, "\n"}, tEOF}},
	{"newline rn", "\r\n", []lexeme{{tokenTypeWhitespace, 0, 0, "\r"}, {tokenTypeWhitespace, 0, 0, "\n"}, tEOF}},

	{"comment", "// a comment", []lexeme{{tokenTypeComment, 0, 0, "// a comment"}, tEOF}},
	{"multiline comment", "/* a comment */foo", []lexeme{
		{tokenTypeComment, 0, 0, "/* a comment */"}, {tokenTypeIdentifier, 0, 0, "foo"}, tEOF,
	}},
	{"multiline comment 2", "/* a\ncomment */foo", []lexeme{
		{tokenTypeComment, 0, 0, "/* a\ncomment */"}, {tokenTypeIdentifier, 0, 0, "foo"}, tEOF,
	}},

	{"left brace", "{", []lexeme{{tokenTypeLeftBrace, 0, 0, "{"}, tEOF}},
	{"right brace", "}", []lexeme{{tokenTypeRightBrace, 0, 0, "}"}, tEOF}},

	{"left bracket", "[", []lexeme{{tokenTypeLeftBracket, 0, 0, "["}, tEOF}},
	{"right bracket", "]", []lexeme{{tokenTypeRightBracket, 0, 0, "]"}, tEOF}},

	{"left paren", "(", []lexeme{{tokenTypeLeftParen, 0, 0, "("}, tEOF}},
	{"right paren", ")", []lexeme{{tokenTypeRightParen, 0, 0, ")"}, tEOF}},

	{"semicolon", ";", []lexeme{{tokenTypeSemicolon, 0, 0, ";"}, tEOF}},
	{"comma", ",", []lexeme{{tokenTypeComma, 0, 0, ","}, tEOF}},
	{"variadic", "...", []lexeme{{tokenTypeVariadic, 0, 0, "..."}, tEOF}},

	{"keyword", "interface", []lexeme{{tokenTypeIdentifier, 0, 0, "interface"}, tEOF}},
	{"identifier", "interace", []lexeme{{tokenTypeIdentifier, 0, 0, "interace"}, tEOF}},
	{"string", `"val"`, []lexeme{{tokenTypeString, 0, 0, `"val"`}, tEOF}},
	{"string esc", `"va\"l"`, []lexeme{{tokenTypeString, 0, 0, `"va\"l"`}, tEOF}},
	{"string noesc", `"val\\"`, []lexeme{{tokenTypeString, 0, 0, `"val\\"`}, tEOF}},
	{"number", `0.0`, []lexeme{{tokenTypeNumber, 0, 0, `0.0`}, tEOF}},
}

func TestLexer(t *testing.T) {
	for _, test := range lexerTests {
		t.Run(test.name, func(t *testing.T) {
			tokens := collect(&test)
			if !equal(tokens, test.tokens, false) {
				t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", test.name, tokens, test.tokens)
			}
		})
	}
}

// collect gathers the emitted tokens into a slice.
func collect(t *lexerTest) (tokens []lexeme) {
	l := lex(t.input)
	for {
		token := l.nextToken()
		tokens = append(tokens, token)
		if token.kind == tokenTypeEOF || token.kind == tokenTypeError {
			break
		}
	}
	return
}

// equal checks that the two sets of tokens are structurally equal
func equal(i1, i2 []lexeme, checkPos bool) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].kind != i2[k].kind {
			return false
		}
		if i1[k].value != i2[k].value {
			return false
		}
		if checkPos && i1[k].position != i2[k].position {
			return false
		}
	}
	return true
}
