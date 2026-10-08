// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Based on design first introduced in: http://blog.golang.org/two-go-talks-lexical-scanning-in-go-and
// Portions copied and modified from: https://github.com/golang/go/blob/master/src/text/template/parse/lex.go

//go:generate stringer -type=tokenType -trimprefix=tokenType

package parser

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	return buildlex(input, performLexSource, isWhitespaceToken)
}

// tokenType identifies the type of lexer lexemes.
type tokenType int

const (
	tokenTypeError tokenType = iota // error occurred; value is text of error
	tokenTypeEOF
	tokenTypeWhitespace
	tokenTypeComment

	tokenTypeIdentifier // helloworld, interface
	tokenTypeString     // "hello"
	tokenTypeNumber     // 123

	tokenTypeLeftBrace    // {
	tokenTypeRightBrace   // }
	tokenTypeLeftParen    // (
	tokenTypeRightParen   // )
	tokenTypeLeftBracket  // [
	tokenTypeRightBracket // ]
	tokenTypeLeftTri      // [
	tokenTypeRightTri     // ]

	tokenTypeEquals       // =
	tokenTypeSemicolon    // ;
	tokenTypeComma        // ,
	tokenTypeQuestionMark // ?
	tokenTypeColon        // :
	tokenTypeVariadic     // ...
)

func isWhitespaceToken(kind tokenType) bool {
	return kind == tokenTypeWhitespace
}

// performLexSource scans until EOFRUNE
func performLexSource(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case r == EOFRUNE:
			break Loop

		case r == '{':
			l.emit(tokenTypeLeftBrace)

		case r == '}':
			l.emit(tokenTypeRightBrace)

		case r == '(':
			l.emit(tokenTypeLeftParen)

		case r == ')':
			l.emit(tokenTypeRightParen)

		case r == '[':
			l.emit(tokenTypeLeftBracket)

		case r == ']':
			l.emit(tokenTypeRightBracket)

		case r == '<':
			l.emit(tokenTypeLeftTri)

		case r == '>':
			l.emit(tokenTypeRightTri)

		case r == ';':
			l.emit(tokenTypeSemicolon)

		case r == ',':
			l.emit(tokenTypeComma)

		case r == '.':
			if l.acceptString("..") {
				l.emit(tokenTypeVariadic)
			} else {
				return l.errorf("unrecognized character at this location: %#U", r)
			}

		case r == '=':
			l.emit(tokenTypeEquals)

		case r == '?':
			l.emit(tokenTypeQuestionMark)

		case r == ':':
			l.emit(tokenTypeColon)

		case isSpace(r) || isNewline(r):
			l.emit(tokenTypeWhitespace)

		case r == '"':
			l.backup()
			return lexStringLiteral

		case isNumber(r):
			l.backup()
			return lexNumber

		case isAlphaNumeric(r):
			l.backup()
			return lexIdentifierOrKeyword

		case r == '/':
			return lexComment

		default:
			return l.errorf("unrecognized character at this location: %#U", r)
		}
	}

	l.emit(tokenTypeEOF)
	return nil
}

func lexComment(l *lexer) stateFn {
	switch l.peek() {
	case '/':
		l.accept("/")
		return buildLexUntil(tokenTypeComment, func(r rune) (bool, error) {
			result := r == EOFRUNE || isNewline(r)
			return !result, nil
		})
	case '*':
		l.accept("*")
		for {
			if l.acceptString("*/") {
				break
			}
			l.next()
		}
		l.emit(tokenTypeComment)
	}
	return lexSource
}

// lexIdentifierOrKeyword searches for a keyword or literal identifier.
func lexIdentifierOrKeyword(l *lexer) stateFn {
	for {
		if !isAlphaNumeric(l.peek()) {
			break
		}

		l.next()
	}
	l.emit(tokenTypeIdentifier)
	return lexSource
}

func lexStringLiteral(l *lexer) stateFn {
	l.accept(`"`)
	esc := false
	for {
		c := l.peek()
		if c == '"' && !esc {
			l.next()
			break
		}
		esc = c == '\\' && !esc
		l.next()
	}
	l.emit(tokenTypeString)
	return lexSource
}
