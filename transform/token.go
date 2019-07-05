package transform

import (
	"unicode"
)

type lexCommandItem struct {
	name  string
	state stateFn
}

var lexCommandList []lexCommandItem

func init() {
	lexCommandList = []lexCommandItem{
		lexCommandItem{"on", lexCommandOn},
		lexCommandItem{"patch", lexCommandPatch},
		lexCommandItem{"changetype", lexCommandChangeType},
		lexCommandItem{"replace", lexCommandReplace},
	}
}

func lexLineStart(l *lexer) stateFn {
	l.ignore()
main:
	for {
		// word evaluation
		switch {
		case l.evalWord("##", true):
			l.emit(itemTypeHeader)
			ignoreWhitespaces(l)
			tryConsumeIdent(l)
			return requireRemaningToBeEmpty(l, lexLineStart)
		case l.evalWord("#", true):
			l.emit(itemFileHeader)
			return lexValueOrString
		}

		// rune evaluation
		ch := l.next()
		// fmt.Printf("LINE START: '%c'\n", ch)
		switch {
		case ch == eof:
			break main
		case isNewLine(ch):
			l.emit(itemNewLine)
			continue main
		case unicode.IsSpace(ch):
			return lexCommentLine
		case ch == '.':
			l.emit(itemSpecial)
			return lexPropertyStart
		case ch == '@':
			return lexCommandStart
		case isIdentFirst(ch):
			return lexRenameStmt
		default:
			return l.errorf("unknown line start")
		}
	}
	l.emit(itemNewLine)
	l.emit(itemEOF)
	return nil
}

func lexCommandStart(l *lexer) stateFn {
	// e.g. @on "HTML." : .package = github.com/gowebapi/webapi/html
	l.ignore()

	for _, item := range lexCommandList {
		if l.acceptWord(item.name) {
			l.emit(itemCommand)
			next := item.state
			next = requireWhitespace(l, next)
			return next
		}
	}
	return l.errorf("unknown command or invalid syntax")
}

func lexCommandOn(l *lexer) stateFn {
	// e.g. @on "HTML." : .package = github.com/gowebapi/webapi/html
	var match bool
	match = match || l.acceptWord("interface")
	match = match || l.acceptWord("enum")
	match = match || l.acceptWord("callback")
	match = match || l.acceptWord("dictionary")
	if match {
		l.emit(itemKeyword)
		if !l.acceptWith(isWhitespace) {
			return l.errorf("invalid command or syntax")
		}
		l.ignore()
	} else {
		ignoreWhitespaces(l)
	}

	// consume regular expression
	if state := tryConsumeString(l); state != nil {
		return state
	}

	// white spaces before ':'
	ignoreWhitespaces(l)

	// expecting ':'
	if !l.accept(":") {
		return l.errorf("expected ':' after regular expression string")
	}
	l.emit(itemSpecial)
	ignoreWhitespaces(l)

	// evalaute command that should be executed
	ch := l.next()
	switch {
	case ch == '.':
		l.emit(itemSpecial)
		return lexPropertyStart
	case ch == '@':
		l.ignore()
		return lexCommandStart
	case isIdentFirst(ch):
		return lexRenameStmt
	}
	return l.errorf("unknown or unsupported command after ':'")
}

func lexCommandChangeType(l *lexer) stateFn {
	ch := l.next()
	if !isIdentFirst(ch) {
		return l.errorf("expected identifier as first argument for @changetype")
	}
	l.acceptWith(isReferenceName)
	l.emit(itemIdent)

	// space in between
	next := emitNewLineGotoLineStart
	next = requireWhitespace(l, next)

	if l.acceptWord("rawjs") {
		l.emit(itemKeyword)
	} else {
		return l.errorf("missing or invalid type argument. valid are 'rawjs'")
	}
	ignoreWhitespaces(l)
	return emitNewLineGotoLineStart
}

func lexCommandPatch(l *lexer) stateFn {
	if l.acceptWord("idlconst") {
		l.emit(itemKeyword)
	}
	return emitNewLineGotoLineStart
}

func lexCommandReplace(l *lexer) stateFn {
	ch := l.next()
	if ch == '.' {
		l.emit(itemSpecial)
	} else {
		l.backup()
	}
	tryConsumeIdent(l)
	next := emitNewLineGotoLineStart
	next = requireWhitespace(l, next)
	tryConsumeString(l)
	next = requireWhitespace(l, next)
	tryConsumeString(l)
	l.acceptWith(isWhitespace)
	return next
}

func lexPropertyStart(l *lexer) stateFn {
	tryConsumeIdent(l)
	ignoreWhitespaces(l)
	if l.acceptWord("=") {
		l.emit(itemSpecial)
		return lexValueOrString
	}
	return l.errorf("expected to find '=' on property line")
}

func lexRenameStmt(l *lexer) stateFn {
	l.acceptWith(isReferenceName)
	l.emit(itemIdent)
	ignoreWhitespaces(l)
	if l.acceptWord("=") {
		l.emit(itemSpecial)
		return lexValueOrString
	}
	return l.errorf("expected to find '=' on rename line")
}

// read reamning of the line as a value or a string
func lexValueOrString(l *lexer) stateFn {
	ignoreWhitespaces(l)
	ch := l.next()
	if isNewLine(ch) {
		l.emit(itemNewLine)
		return lexLineStart
	}
	if ch == '"' {
		// a string line
		l.ignore()
		for {
			ch = l.next()
			switch {
			case isNewLine(ch):
				return l.errorf("unexpected end of line, missing '\"'")
			case ch == '"':
				l.backup()
				l.emit(itemString)
				l.next()
				l.ignore()
				return requireRemaningToBeEmpty(l, lexLineStart)
			}
		}
	} else {
		// a value line
		for {
			ch = l.next()
			if isNewLine(ch) {
				l.backup()
				l.emit(itemValue)
				return emitNewLineGotoLineStart(l)
			}
		}
	}
}

func lexCommentLine(l *lexer) stateFn {
	ignoreWhitespaces(l)
	for {
		ch := l.next()
		if isNewLine(ch) {
			l.backup()
			l.emit(itemComment)
			l.next()
			l.emit(itemNewLine)
			return lexLineStart
		}
	}
}

func emitNewLineGotoLineStart(l *lexer) stateFn {
	ch := l.next()
	if !isNewLine(ch) {
		panic("faulty")
	}
	l.emit(itemNewLine)
	return lexLineStart
}

func tryConsumeIdent(l *lexer) {
	ch := l.next()
	if !isIdentFirst(ch) {
		l.backup()
		return
	}
	l.acceptWith(isIdentAny)
	l.emit(itemIdent)
}

func tryConsumeString(l *lexer) stateFn {
	ch := l.next()
	if ch != '"' {
		return l.errorf("expected a string inside \"...\"")
	}
	l.ignore()
	escape := false
main:
	for {
		ch = l.next()
		if isNewLine(ch) {
			return l.errorf("unexpected end of string, missing '\"'")
		}
		if !escape {
			switch ch {
			case '"':
				break main
			case '\\':
				escape = true
			}
		} else {
			escape = false
		}
	}
	l.backup()
	l.emit(itemString)
	l.next()
	l.ignore()
	return nil
}

/*
func tryConsumeNumber(l *lexer) {
	// Optional leading sign.
	l.accept("+-")
	// Is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	l.emit(itemNumber)
}
*/

func ignoreWhitespaces(l *lexer) {
	l.acceptWith(isWhitespace)
	l.ignore()
}

func requireWhitespace(l *lexer, onSucess stateFn) stateFn {
	if !l.acceptWith(isWhitespace) {
		return l.errorf("invalid command or syntax, expecting white space")
	}
	l.ignore()
	return onSucess
}

func requireRemaningToBeEmpty(l *lexer, next stateFn) stateFn {
	l.acceptRun(" \t")
	l.ignore()
	ch := l.next()
	if isNewLine(ch) {
		l.emit(itemNewLine)
		return next
	}
	return l.errorf("unexpected character '%c'", ch)
}

func isNewLine(ch rune) bool {
	return ch == '\n' || ch == '\r' || ch == eof
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

func isNumberStart(ch rune) bool {
	return isDigit(ch) || ch == '-' || ch == '+'
}

func isIdentFirst(ch rune) bool {
	return isLetter(ch) || ch == '_'
}

func isIdentAny(ch rune) bool {
	return isIdentFirst(ch) || isDigit(ch)
}

func isReferenceName(ch rune) bool {
	return isIdentAny(ch) || ch == '.'
}
