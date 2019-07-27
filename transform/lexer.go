package transform

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

var eof = rune(0)

type itemType int

// replace this types to create a new lexer
const (
	itemError   itemType = iota
	itemEOF              // 1: End of file
	itemNewLine          // 2: New Line
	itemSpecial          // 3: special char
	itemIdent            // 4: Identifier
	itemComment          // 5: coment line
	itemFileHeader
	itemTypeHeader
	itemString
	itemValue
	itemCommand
	itemWord
	itemKeyword
)

type item struct {
	typ  itemType
	val  string
	line int
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// stateFn represents the state of the scanner
// as a function that returns the next state.
type stateFn func(*lexer) stateFn

type acceptFn func(ch rune) bool
type acceptIdxFn func(ch rune, idx int) bool

// lexer holds the state of the scanner.
type lexer struct {
	name  string    // used only for error reports.
	input string    // the string being scanned.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
	line  int       // current line number
	wasNL bool      // if last rune was a new line (used by backup())
	items chan item // channel of scanned items.
	state stateFn
	fail  bool // indicate that an error have been sent into the items channel
}

func newLex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		line:  1,
		state: lexLineStart,
		items: make(chan item, 10),
	}
	return l
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			l.state = l.state(l)
		}
	}
	// panic("not reached")
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	line := l.line
	if l.wasNL {
		// on new line we get faulty line number as we increase the
		// line number in next() before doing evaluation of rune
		line--
	}
	l.items <- item{t, l.input[l.start:l.pos], line}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	if l.wasNL {
		l.line--
		l.wasNL = false
	}
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() rune {
	ch := l.next()
	l.backup()
	return ch
}

// next returns the next rune in the input.
func (l *lexer) next() (ch rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	l.wasNL = false
	ch, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	if ch == '\n' {
		l.line++
		l.wasNL = true
	}
	return ch
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
		l.line,
	}
	l.fail = true
	return nil
}

// accept consumes the next rune
// if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// acceptWith is calling the accept function for every
// rune and will continue as long as the function return true
func (l *lexer) acceptWith(fn acceptFn) bool {
	return l.acceptWithIdx(func(ch rune, idx int) bool {
		return fn(ch)
	})
}

// acceptWith is calling the accept function for every
// rune and will continue as long as the function return true
func (l *lexer) acceptWithIdx(fn acceptIdxFn) bool {
	idx := 0
	for fn(l.next(), idx) {
		idx++
	}
	l.backup()
	return idx > 0
}

// evalWord is evaluting if comming next() will result
// in a given word. If accept is true, next() will be called
// and consume that word.
// return true if the word was found, otherwise false
func (l *lexer) evalWord(expected string, accept bool) bool {
	current := l.input[l.pos:]
	if !strings.HasPrefix(current, expected) {
		return false
	}
	if accept {
		for _, e := range expected {
			a := l.next()
			if a != e {
				panic("how can HasPrefix be true but not when iterating over it?")
			}
		}
	}
	return true
}

func (l *lexer) acceptWord(expected string) bool {
	return l.evalWord(expected, true)
}
