// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"container/list"
	"fmt"
)

// peekableLexer wraps a lexer and provides the ability to peek forward without
// losing state.
type peekableLexer struct {
	lex        *lexer     // a reference to the lexer used for tokenization
	readTokens *list.List // tokens already read from the lexer during a lookahead.
}

// peekableLex returns a new peekableLexer for the given lexer.
func peekableLex(lex *lexer) *peekableLexer {
	return &peekableLexer{
		lex:        lex,
		readTokens: list.New(),
	}
}

// nextToken returns the next token found in the lexer.
func (l *peekableLexer) nextToken() lexeme {
	frontElement := l.readTokens.Front()
	if frontElement != nil {
		return l.readTokens.Remove(frontElement).(lexeme)
	}

	return l.lex.nextToken()
}

// peekToken performs lookahead of the given count on the token stream.
func (l *peekableLexer) peekToken(count int) lexeme {
	if count < 1 {
		panic(fmt.Sprintf("Expected count >= 1, received: %v", count))
	}

	// Ensure that the readTokens has at least the requested number of tokens.
	for l.readTokens.Len() < count {
		l.readTokens.PushBack(l.lex.nextToken())
	}

	// Retrieve the count-th token from the list.
	element := l.readTokens.Front()

	for i := 1; i < count; i++ {
		element = element.Next()
	}

	return element.Value.(lexeme)
}
