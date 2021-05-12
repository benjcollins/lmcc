package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	SUM        = iota
	COMPARISON = iota
	LOGIC      = iota
	EXPR
)

type Parser struct {
	pos    Position
	source string
	errors []ParseError
}

type ParseError struct {
	pos Position
	msg string
}

func (err ParseError) String() string {
	return fmt.Sprintf("%s at %s", err.msg, err.pos)
}

type Position struct {
	line, column, index int
}

func Length(start, end Position) int {
	return end.index - start.index
}

func (pos Position) String() string {
	return fmt.Sprintf("(%d, %d)", pos.line, pos.column)
}

func (parser *Parser) peek() rune {
	r, _ := utf8.DecodeRuneInString(parser.source[parser.pos.index:])
	return r
}

func (parser *Parser) next() {
	r, size := utf8.DecodeRuneInString(parser.source[parser.pos.index:])
	if r == '\n' {
		parser.pos.line++
		parser.pos.column = 1
	} else {
		parser.pos.column++
	}
	parser.pos.index += size
}

func (parser *Parser) eof() bool {
	_, s := utf8.DecodeRuneInString(parser.source[parser.pos.index:])
	return s == 0
}

func (parser *Parser) error(msg string) {
	parser.errors = append(parser.errors, ParseError{parser.pos, msg})
}

func (parser *Parser) skipSpaces() {
	for unicode.IsSpace(parser.peek()) {
		parser.next()
	}
}

func (parser *Parser) parseInt() (Expr, bool) {
	if !unicode.IsDigit(parser.peek()) {
		return Expr{}, false
	}
	pos := parser.pos
	for unicode.IsDigit(parser.peek()) {
		parser.next()
	}
	contents := parser.source[pos.index:parser.pos.index]
	value, _ := strconv.Atoi(contents)
	return Expr{pos, Length(pos, parser.pos), IntLiteral{value}}, true
}

func (parser *Parser) parseIdent() (string, bool) {
	if !unicode.IsLetter(parser.peek()) {
		return "", false
	}
	pos := parser.pos
	for unicode.IsLetter(parser.peek()) || unicode.IsDigit(parser.peek()) {
		parser.next()
	}
	name := parser.source[pos.index:parser.pos.index]
	return name, true
}

func (parser *Parser) parseSymbol(symbol string) bool {
	if !strings.HasPrefix(parser.source[parser.pos.index:], symbol) {
		return false
	}
	parser.pos.index += len(symbol)
	return true
}

func (parser *Parser) parseValue() Expr {
	pos := parser.pos
	if literal, ok := parser.parseInt(); ok {
		return literal
	}
	if name, ok := parser.parseIdent(); ok {
		var node ExprNode
		switch name {
		case "true":
			node = BoolLiteral{true}
		case "false":
			node = BoolLiteral{false}
		case "in":
			node = Input{}
		default:
			node = Ident{name}
		}
		return Expr{pos, Length(pos, parser.pos), node}
	}
	parser.error("expected a value")
	return Expr{}
}

func (parser *Parser) parseInfix(left *Expr, symbol string, prec, symbolPrec int) bool {
	if prec > symbolPrec && parser.parseSymbol(symbol) {
		parser.skipSpaces()
		right := parser.parseExpr(symbolPrec)
		*left = Expr{left.pos, Length(left.pos, parser.pos), Binary{symbol, *left, right}}
		return true
	}
	return false
}

func (parser *Parser) parseUnary() Expr {
	pos := parser.pos
	if parser.parseSymbol("-") {
		parser.skipSpaces()
		expr := parser.parseExpr(SUM)
		return Expr{pos, Length(pos, parser.pos), Unary{"-", expr}}
	}
	if parser.parseSymbol("not") {
		parser.skipSpaces()
		expr := parser.parseExpr(LOGIC)
		return Expr{pos, Length(pos, parser.pos), Unary{"not", expr}}
	}
	return parser.parseValue()
}

func (parser *Parser) parseExpr(prec int) Expr {
	left := parser.parseUnary()
	for {
		parser.skipSpaces()

		parsed := parser.parseInfix(&left, "+", prec, SUM) ||
			parser.parseInfix(&left, "-", prec, SUM) ||

			parser.parseInfix(&left, "==", prec, COMPARISON) ||
			parser.parseInfix(&left, "!=", prec, COMPARISON) ||
			parser.parseInfix(&left, "<=", prec, COMPARISON) ||
			parser.parseInfix(&left, ">=", prec, COMPARISON) ||
			parser.parseInfix(&left, "<", prec, COMPARISON) ||
			parser.parseInfix(&left, ">", prec, COMPARISON) ||

			parser.parseInfix(&left, "and", prec, LOGIC) ||
			parser.parseInfix(&left, "or", prec, LOGIC)

		if !parsed {
			return left
		}
	}
}

func (parser *Parser) parseStatement() Statement {
	pos := parser.pos

	if parser.parseSymbol("if") {
		parser.skipSpaces()
		cond := parser.parseExpr(EXPR)
		parser.skipSpaces()
		ifTrue := parser.parseStatement()
		parser.skipSpaces()
		var ifFalse Statement
		if parser.parseSymbol("else") {
			parser.skipSpaces()
			ifFalse = parser.parseStatement()
		}
		return Statement{pos, Length(pos, parser.pos), If{cond, ifTrue, ifFalse}}
	}
	if parser.parseSymbol("while") {
		parser.skipSpaces()
		cond := parser.parseExpr(EXPR)
		parser.skipSpaces()
		loop := parser.parseStatement()
		return Statement{pos, Length(pos, parser.pos), While{cond, loop}}
	}
	if parser.parseSymbol("out") {
		parser.skipSpaces()
		expr := parser.parseExpr(EXPR)
		return Statement{pos, Length(pos, parser.pos), Output{expr}}
	}
	if parser.parseSymbol("{") {
		parser.skipSpaces()
		statements := parser.parseStatements()
		if !parser.parseSymbol("}") {
			parser.error("expected a '}'")
		}
		return Statement{pos, Length(pos, parser.pos), BlockScope{statements}}
	}
	name, ok := parser.parseIdent()
	parser.skipSpaces()
	if ok && parser.parseSymbol(":") {
		parser.skipSpaces()
		ty := Undefined
		if name, ok := parser.parseIdent(); ok {
			switch name {
			case "int":
				ty = Int
			case "bool":
				ty = Bool
			default:
				parser.error("invalid type name")
				return Statement{}
			}
		}
		parser.skipSpaces()
		var expr Expr
		if parser.parseSymbol("=") {
			parser.skipSpaces()
			expr = parser.parseExpr(EXPR)
		}
		return Statement{pos, Length(pos, parser.pos), Declare{name, expr, ty}}
	}
	if ok && parser.parseSymbol("=") {
		parser.skipSpaces()
		expr := parser.parseExpr(EXPR)
		return Statement{pos, Length(pos, parser.pos), Assign{name, expr}}
	}
	parser.error("was expecting a statement")
	return Statement{}
}

func (parser *Parser) parseStatements() []Statement {
	statements := []Statement{}
	parser.skipSpaces()
	for !parser.eof() && !(parser.peek() == '}') && !(parser.peek() == ')') {
		statement := parser.parseStatement()
		if statement.node == nil {
			break
		}
		statements = append(statements, statement)
		parser.skipSpaces()
	}
	return statements
}

func Parse(source string) ([]Statement, []ParseError) {
	parser := Parser{Position{1, 1, 0}, source, []ParseError{}}
	statements := parser.parseStatements()
	return statements, parser.errors
}
