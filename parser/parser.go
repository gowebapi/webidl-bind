// Copyright 2015 The Serulian Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// parser package defines the parser and lexer for translating a *supported subset* of
// WebIDL (http://www.w3.org/TR/WebIDL/) into an AST.
package parser

import (
	"fmt"
	"slices"

	"github.com/gowebapi/webidl-bind/ast"
)

// tryConsumeIdentifier attempts to consume an expected identifier.
func (p *sourceParser) tryConsumeIdentifier() (string, bool) {
	if !p.isToken(tokenTypeIdentifier) {
		return "", false
	}

	value := p.currentToken.value
	p.consumeToken()
	return value, true
}

// consumeIdentifier consumes an expected identifier token or adds an error node.
func (p *sourceParser) consumeIdentifier() string {
	if identifier, ok := p.tryConsumeIdentifier(); ok {
		return identifier
	}

	p.emitError("Expected identifier, found token %v", p.currentToken)
	return ""
}

func (p *sourceParser) consumeLiteral() ast.Literal {
	base := &ast.Base{}
	finish := p.node(base)
	l, ok := p.consume(tokenTypeIdentifier, tokenTypeString, tokenTypeNumber, tokenTypeLeftBracket)
	if !ok {
		p.emitError("Expected literal, found token %v", p.currentToken)
		finish()
		return &ast.BasicLiteral{Base: *base}
	}
	switch l.kind {
	case tokenTypeIdentifier, tokenTypeString, tokenTypeNumber:
		finish()
		return &ast.BasicLiteral{Base: *base, Value: l.value}
	case tokenTypeLeftBracket:
		n := &ast.SequenceLiteral{}
		for !p.isToken(tokenTypeRightBracket) {
			if len(n.Elems) != 0 {
				p.consume(tokenTypeComma)
			}
			if p.isToken(tokenTypeRightBracket) {
				break
			}
			n.Elems = append(n.Elems, p.consumeLiteral())
		}
		// , (optional)
		p.tryConsume(tokenTypeComma)
		p.consume(tokenTypeRightBracket)
		finish()
		n.Base = *base
		return n
	}
	panic("unreachable")
}

// tryParserFn is a function that attempts to build an AST node.
type tryParserFn func() (ast.Node, bool)

// lookaheadParserFn is a function that performs lookahead.
type lookaheadParserFn func(currentToken lexeme) bool

// rightNodeConstructor is a function which takes in a left expr node and the
// token consumed for a left-recursive operator, and returns a newly constructed
// operator expression if a right expression could be found.
type rightNodeConstructor func(ast.Node, lexeme) (ast.Node, bool)

// commentedLexeme is a lexeme with comments attached.
type commentedLexeme struct {
	lexeme
	comments []string
}

// sourceParser holds the state of the parser.
type sourceParser struct {
	startIndex    bytePosition    // The start index for position decoration on nodes.
	lex           *peekableLexer  // a reference to the lexer used for tokenization
	nodes         nodeStack       // the stack of the current nodes
	currentToken  commentedLexeme // the current token
	previousToken commentedLexeme // the previous token
	config        parserConfig    // Configuration for customizing the parser
}

// parserConfig holds configuration for customizing the parser
type parserConfig struct {
	ignoredTokenTypes map[tokenType]struct{} // the token types ignored by the parser
}

// buildParser returns a new sourceParser instance.
func buildParser(lexer *lexer, config parserConfig, startIndex bytePosition) *sourceParser {
	l := peekableLex(lexer)
	newLexeme := func() commentedLexeme {
		return commentedLexeme{lexeme: lexeme{tokenTypeEOF, 0, 0, ""}}
	}
	return &sourceParser{
		startIndex:    startIndex,
		lex:           l,
		currentToken:  newLexeme(),
		previousToken: newLexeme(),
		config:        config,
	}
}

// createErrorNode creates a new error node and returns it.
func (p *sourceParser) createErrorNode(format string, args ...any) *ast.ErrorNode {
	n := &ast.ErrorNode{Message: fmt.Sprintf(format, args...)}
	p.decorateStartRuneAndComments(n, p.currentToken)
	p.decorateEndRune(n, p.previousToken)
	return n
}

// node creates a new node of the given type, decorates it with the current token's
// position as its start position, and pushes it onto the nodes stack.
func (p *sourceParser) node(node ast.Node) func() {
	p.decorateStartRuneAndComments(node, p.currentToken)
	p.nodes.push(node)
	return func() {
		// finishNode pops the current node from the top of the stack and decorates it with
		// the current token's end position as its end position.
		if p.currentNode() == nil {
			panic(fmt.Sprintf("No current node on stack. Token: %s", p.currentToken.value))
		}

		p.decorateEndRune(p.currentNode(), p.previousToken)
		p.nodes.pop()
	}
}

// decorateStartRuneAndComments decorates the given node with the location of the given token as its
// starting rune, as well as any comments attached to the token.
func (p *sourceParser) decorateStartRuneAndComments(node ast.Node, token commentedLexeme) {
	b := node.NodeBase()
	b.Start = int(token.position) + int(p.startIndex)
	b.Line = int(token.line)
	p.decorateComments(node, token.comments)
}

// decorateComments decorates the given node with the specified comments.
func (p *sourceParser) decorateComments(node ast.Node, comments []string) {
	b := node.NodeBase()
	b.Comments = append(b.Comments, comments...)
}

// decorateEndRune decorates the given node with the location of the given token as its
// ending rune.
func (p *sourceParser) decorateEndRune(node ast.Node, token commentedLexeme) {
	node.NodeBase().End = int(token.position) + len(token.value) - 1 + int(p.startIndex)
}

// currentNode returns the node at the top of the stack.
func (p *sourceParser) currentNode() ast.Node {
	return p.nodes.topValue()
}

// consumeToken advances the lexer forward, returning the next token.
func (p *sourceParser) consumeToken() commentedLexeme {
	var comments = make([]string, 0)

	for {
		token := p.lex.nextToken()

		if token.kind == tokenTypeComment {
			comments = append(comments, token.value)
		}

		if _, ok := p.config.ignoredTokenTypes[token.kind]; !ok {
			p.previousToken = p.currentToken
			p.currentToken = commentedLexeme{token, comments}
			return p.currentToken
		}
	}
}

// isToken returns true if the current token matches one of the types given.
func (p *sourceParser) isToken(types ...tokenType) bool {
	return slices.Contains(types, p.currentToken.kind)
}

// nextToken returns the next token found, without advancing the parser. Used for
// lookahead.
func (p *sourceParser) nextToken() lexeme {
	for i := range 1000 {
		token := p.lex.peekToken(i + 1)
		if _, ok := p.config.ignoredTokenTypes[token.kind]; !ok {
			return token
		}
	}
	panic("stale")
}

// isNextToken returns true if the *next* token matches one of the types given.
func (p *sourceParser) isNextToken(types ...tokenType) bool {
	token := p.nextToken()

	return slices.Contains(types, token.kind)
}

func (p *sourceParser) isIdentifier(name string) bool {
	return p.isToken(tokenTypeIdentifier) && p.currentToken.value == name
}

// isNextIdentifier returns true if the next token is a keyword matching that given.
func (p *sourceParser) isNextIdentifier(keyword string) bool {
	token := p.nextToken()
	return token.kind == tokenTypeIdentifier && token.value == keyword
}

// emitError creates a new error node and attachs it as a child of the current
// node.
func (p *sourceParser) emitError(format string, args ...any) {
	errorNode := p.createErrorNode(format, args...)
	b := p.currentNode().NodeBase()
	b.Errors = append(b.Errors, errorNode)
}

// consumeKeyword consumes an expected keyword token or adds an error node.
func (p *sourceParser) consumeKeyword(keyword string) bool {
	if !p.tryConsumeKeyword(keyword) {
		p.emitError("Expected keyword %s, found token %v", keyword, p.currentToken)
		return false
	}
	return true
}

// tryConsumeKeyword attempts to consume an expected keyword token.
func (p *sourceParser) tryConsumeKeyword(keyword string) bool {
	if !p.isIdentifier(keyword) {
		return false
	}

	p.consumeToken()
	return true
}

// consume performs consumption of the next token if it matches any of the given
// types and returns it. If no matching type is found, adds an error node.
func (p *sourceParser) consume(types ...tokenType) (lexeme, bool) {
	token, ok := p.tryConsume(types...)
	if !ok {
		p.emitError("Expected one of: %v, found: %v", types, p.currentToken)
	}
	return token, ok
}

// tryConsume performs consumption of the next token if it matches any of the given
// types and returns it.
func (p *sourceParser) tryConsume(types ...tokenType) (lexeme, bool) {
	token, found := p.tryConsumeWithComments(types...)
	return token.lexeme, found
}

// tryConsume performs consumption of the next token if it matches any of the given
// types and returns it.
func (p *sourceParser) tryConsumeWithComments(types ...tokenType) (commentedLexeme, bool) {
	if p.isToken(types...) {
		token := p.currentToken
		p.consumeToken()
		return token, true
	}

	return commentedLexeme{lexeme{tokenTypeError, -1, -1, ""}, make([]string, 0)}, false
}

// consumeUntil consumes all tokens until one of the given token types is found.
func (p *sourceParser) consumeUntil(types ...tokenType) lexeme {
	for {
		found, ok := p.tryConsume(types...)
		if ok {
			return found
		}

		p.consumeToken()
	}
}

// oneOf runs each of the sub parser functions, in order, until one returns true. Otherwise
// returns nil and false.
func (p *sourceParser) oneOf(subParsers ...tryParserFn) (ast.Node, bool) {
	for _, subParser := range subParsers {
		node, ok := subParser()
		if ok {
			return node, ok
		}
	}
	return nil, false
}

// performLeftRecursiveParsing performs left-recursive parsing of a set of operators. This method
// first performs the parsing via the subTryExprFn and then checks for one of the left-recursive
// operator token types found. If none found, the left expression is returned. Otherwise, the
// rightNodeBuilder is called to attempt to construct an operator expression. This method also
// properly handles decoration of the nodes with their proper start and end run locations.
func (p *sourceParser) performLeftRecursiveParsing(subTryExprFn tryParserFn, rightNodeBuilder rightNodeConstructor, rightTokenTester lookaheadParserFn, operatorTokens ...tokenType) (ast.Node, bool) {
	var currentLeftToken commentedLexeme
	currentLeftToken = p.currentToken

	// Consume the left side of the expression.
	leftNode, ok := subTryExprFn()
	if !ok {
		return nil, false
	}

	// Check for an operator token. If none found, then we've found just the left side of the
	// expression and so we return that node.
	if !p.isToken(operatorTokens...) {
		return leftNode, true
	}

	// Keep consuming pairs of operators and child expressions until such
	// time as no more can be consumed. We use this loop+custom build rather than recursion
	// because these operators are *left* recursive, not right.
	var currentLeftNode ast.Node
	currentLeftNode = leftNode

	for {
		// Check for an operator.
		if !p.isToken(operatorTokens...) {
			break
		}

		// If a lookahead function is defined, check the lookahead for the matched token.
		if rightTokenTester != nil && !rightTokenTester(p.currentToken.lexeme) {
			break
		}

		// Consume the operator.
		operatorToken, ok := p.tryConsumeWithComments(operatorTokens...)
		if !ok {
			break
		}

		// Consume the right hand expression and build an expression node (if applicable).
		exprNode, ok := rightNodeBuilder(currentLeftNode, operatorToken.lexeme)
		if !ok {
			p.emitError("Expected right hand expression, found: %v", p.currentToken)
			return currentLeftNode, true
		}

		p.decorateStartRuneAndComments(exprNode, currentLeftToken)
		p.decorateEndRune(exprNode, p.previousToken)

		currentLeftNode = exprNode
		currentLeftToken = operatorToken
	}

	return currentLeftNode, true
}
