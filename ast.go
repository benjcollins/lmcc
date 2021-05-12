package main

import "strings"

type Expr struct {
	pos    Position
	length int
	node   ExprNode
}

type ExprNode interface {
	compileValue(*Assembly, **Block, *Scope, Position) (Value, error)
	compileCondition(*Assembly, **Block, *Block, *Block, *Scope, Position) error
	prettyPrint(*strings.Builder)
}

type IntLiteral struct {
	value int
}

type BoolLiteral struct {
	value bool
}

type Input struct{}

type Ident struct {
	name string
}

type Binary struct {
	symbol      string
	left, right Expr
}

type Unary struct {
	symbol string
	expr   Expr
}

type Statement struct {
	pos    Position
	length int
	node   StatementNode
}

type StatementNode interface {
	compile(*Assembly, **Block, *Scope, Position, []error) []error
	prettyPrint(*strings.Builder, string)
}

func (statement Statement) compile(asm *Assembly, block **Block, scope *Scope, errors []error) []error {
	return statement.node.compile(asm, block, scope, statement.pos, errors)
}

type Declare struct {
	name string
	expr Expr
	ty   Type
}

type Assign struct {
	name string
	expr Expr
}

type BlockScope struct {
	statements []Statement
}

type If struct {
	cond    Expr
	ifTrue  Statement
	ifFalse Statement
}

type While struct {
	cond Expr
	loop Statement
}

type Output struct {
	expr Expr
}
