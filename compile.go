package main

import (
	"fmt"
	"strings"
)

const (
	Int       Type = iota
	Bool      Type = iota
	Undefined Type = iota
)

type Value struct {
	ty    Type
	acc   bool
	label string
}

type Type int

func (k Type) String() string {
	switch k {
	case Int:
		return "int"
	case Bool:
		return "bool"
	default:
		return "undefined"
	}
}

func (expr Expr) compileValue(asm *Assembly, block **Block, scope *Scope) (Value, error) {
	return expr.node.compileValue(asm, block, scope, expr.pos)
}

func (expr Expr) compileCondition(asm *Assembly, block **Block, ifTrue, ifFalse *Block, scope *Scope) error {
	return expr.node.compileCondition(asm, block, ifTrue, ifFalse, scope, expr.pos)
}

func (literal IntLiteral) compileValue(asm *Assembly, block **Block, scope *Scope, pos Position) (Value, error) {
	return Value{Int, false, asm.getConstant(literal.value)}, nil
}

func (literal BoolLiteral) compileValue(asm *Assembly, block **Block, scope *Scope, pos Position) (Value, error) {
	value := 0
	if literal.value {
		value = 1
	}
	return Value{Bool, false, asm.getConstant(value)}, nil
}

func (input Input) compileValue(asm *Assembly, block **Block, scope *Scope, pos Position) (Value, error) {
	(*block).emitInstruction("INP", "")
	return Value{Int, true, ""}, nil
}

func (ident Ident) compileValue(asm *Assembly, block **Block, scope *Scope, pos Position) (Value, error) {
	label, ty, prs := scope.get(ident.name)
	if !prs {
		return Value{}, fmt.Errorf("undefined variable '%s' at %s", ident.name, pos)
	}
	return Value{ty, false, label}, nil
}

func (bin Binary) compileValue(asm *Assembly, block **Block, scope *Scope, pos Position) (Value, error) {
	switch bin.symbol {
	case "+", "-":
		return compileArithmetic(asm, block, scope, bin.symbol, bin.left, bin.right, pos)
	case "==", "!=", ">", "<", ">=", "<=", "and", "or":
		ifTrue := asm.newUniqueBlock()
		ifFalse := asm.newUniqueBlock()
		exitBlock := asm.newUniqueBlock()

		ifFalse.emitInstruction("LDA", asm.getConstant(0))
		ifTrue.emitInstruction("LDA", asm.getConstant(1))
		ifTrue.emitInstruction("BRA", exitBlock.label)
		ifFalse.emitInstruction("BRA", exitBlock.label)

		if err := compileBinaryCondition(bin, asm, block, ifTrue, ifFalse, scope); err != nil {
			return Value{}, err
		}
		*block = exitBlock
		return Value{Bool, true, ""}, nil
	default:
		panic("undefined symbol")
	}
}

func (unary Unary) compileValue(asm *Assembly, block **Block, scope *Scope, pos Position) (Value, error) {
	switch unary.symbol {
	case "-":
		val, err := compileAndExpect(unary.expr, asm, block, scope, Int)
		if err != nil {
			return Value{}, err
		}
		label := storeToTemp(val, asm, *block)
		defer popTemp(val, asm)
		(*block).emitInstruction("LDA", asm.getConstant(0))
		(*block).emitInstruction("SUB", label)
		return Value{Int, true, ""}, nil
	case "not":
		ifTrue := asm.newUniqueBlock()
		ifFalse := asm.newUniqueBlock()
		exitBlock := asm.newUniqueBlock()

		ifFalse.emitInstruction("LDA", asm.getConstant(0))
		ifTrue.emitInstruction("LDA", asm.getConstant(1))
		ifTrue.emitInstruction("BRA", exitBlock.label)
		ifFalse.emitInstruction("BRA", exitBlock.label)

		if err := compileUnaryCondition(unary, asm, block, ifTrue, ifFalse, scope, pos); err != nil {
			return Value{}, err
		}
		*block = exitBlock
		return Value{Bool, true, ""}, nil
	}
	panic("LOL")
}

func compileArithmetic(asm *Assembly, block **Block, scope *Scope, symbol string, left, right Expr, pos Position) (Value, error) {
	rightVal, err := compileAndExpect(right, asm, block, scope, Int)
	if err != nil {
		return Value{}, err
	}
	rightLabel := storeToTemp(rightVal, asm, *block)
	defer popTemp(rightVal, asm)
	leftVal, err := compileAndExpect(left, asm, block, scope, Int)
	if err != nil {
		return Value{}, err
	}
	loadToAcc(leftVal, *block)
	switch symbol {
	case "+":
		(*block).emitInstruction("ADD", rightLabel)
	case "-":
		(*block).emitInstruction("SUB", rightLabel)
	}
	return Value{Int, true, ""}, nil
}

func (literal IntLiteral) compileCondition(asm *Assembly, block **Block, ifTrue, ifFalse *Block, scope *Scope, pos Position) error {
	return fmt.Errorf("int used as a condition at %s", pos)
}

func (literal BoolLiteral) compileCondition(asm *Assembly, block **Block, ifTrue, ifFalse *Block, scope *Scope, pos Position) error {
	if literal.value {
		(*block).emitInstruction("BRA", ifTrue.label)
	} else {
		(*block).emitInstruction("BRA", ifFalse.label)
	}
	return nil
}

func (input Input) compileCondition(asm *Assembly, block **Block, ifTrue, ifFalse *Block, scope *Scope, pos Position) error {
	return fmt.Errorf("cannot use input as condition at %s", pos)
}

func (bin Binary) compileCondition(asm *Assembly, block **Block, ifTrue, ifFalse *Block, scope *Scope, pos Position) error {
	return compileBinaryCondition(bin, asm, block, ifTrue, ifFalse, scope)
}

func (unary Unary) compileCondition(asm *Assembly, block **Block, ifTrue, ifFalse *Block, scope *Scope, pos Position) error {
	return compileUnaryCondition(unary, asm, block, ifTrue, ifFalse, scope, pos)
}

func compileUnaryCondition(unary Unary, asm *Assembly, block **Block, ifTrue, ifFalse *Block, scope *Scope, pos Position) error {
	switch unary.symbol {
	case "not":
		return unary.expr.compileCondition(asm, block, ifFalse, ifTrue, scope)
	case "-":
		return fmt.Errorf("cannot use '-' operator in condition at %s", pos)
	}
	panic("invalid symbol")
}

func compileBinaryCondition(bin Binary, asm *Assembly, block **Block, ifTrue, ifFalse *Block, scope *Scope) error {
	switch bin.symbol {
	case ">=":
		return compilePostiveBranch(bin.left, bin.right, asm, block, ifTrue, ifFalse, scope)
	case "<":
		return compilePostiveBranch(bin.left, bin.right, asm, block, ifFalse, ifTrue, scope)
	case "<=":
		return compilePostiveBranch(bin.right, bin.left, asm, block, ifTrue, ifFalse, scope)
	case ">":
		return compilePostiveBranch(bin.right, bin.left, asm, block, ifFalse, ifTrue, scope)
	case "==":
		return compileZeroBranch(bin.left, bin.right, asm, block, ifTrue, ifFalse, scope)
	case "!=":
		return compileZeroBranch(bin.left, bin.right, asm, block, ifFalse, ifTrue, scope)
	case "and":
		nextCondition := asm.newUniqueBlock()
		if err := bin.left.compileCondition(asm, block, nextCondition, ifFalse, scope); err != nil {
			return err
		}
		if err := bin.right.compileCondition(asm, &nextCondition, ifTrue, ifFalse, scope); err != nil {
			return err
		}
	case "or":
		nextCondition := asm.newUniqueBlock()
		if err := bin.left.compileCondition(asm, block, ifTrue, nextCondition, scope); err != nil {
			return err
		}
		if err := bin.right.compileCondition(asm, &nextCondition, ifTrue, ifFalse, scope); err != nil {
			return err
		}
	default:
		panic("undefined symbol")
	}
	return nil
}

func compileCompare(left, right Expr, asm *Assembly, block **Block, scope *Scope) error {
	rightVal, err := compileAndExpect(right, asm, block, scope, Int)
	if err != nil {
		return err
	}
	rightLabel := storeToTemp(rightVal, asm, *block)
	defer popTemp(rightVal, asm)
	leftVal, err := compileAndExpect(left, asm, block, scope, Int)
	if err != nil {
		return err
	}
	loadToAcc(leftVal, *block)
	(*block).emitInstruction("SUB", rightLabel)
	return nil
}

func compileZeroBranch(left, right Expr, asm *Assembly, block **Block, ifTrue, ifFalse *Block, scope *Scope) error {
	if err := compileCompare(left, right, asm, block, scope); err != nil {
		return err
	}
	(*block).emitInstruction("BRP", ifTrue.label)
	(*block).emitInstruction("BRA", ifFalse.label)
	return nil
}

func compilePostiveBranch(left, right Expr, asm *Assembly, block **Block, ifTrue, ifFalse *Block, scope *Scope) error {
	if err := compileCompare(left, right, asm, block, scope); err != nil {
		return err
	}
	(*block).emitInstruction("BRP", ifTrue.label)
	(*block).emitInstruction("BRA", ifFalse.label)
	return nil
}

func loadToAcc(val Value, block *Block) {
	if !val.acc {
		block.emitInstruction("LDA", val.label)
	}
}

func storeToTemp(val Value, asm *Assembly, block *Block) string {
	if val.acc {
		label := asm.pushTemp()
		block.emitInstruction("STA", label)
		return label
	}
	return val.label
}

func popTemp(val Value, asm *Assembly) {
	if !val.acc {
		asm.popTemp()
	}
}

func compileAndExpect(expr Expr, asm *Assembly, block **Block, scope *Scope, ty Type) (Value, error) {
	val, err := expr.compileValue(asm, block, scope)
	if err != nil {
		return Value{}, err
	}
	if val.ty != ty {
		return Value{}, fmt.Errorf("expected a %s instead got %s at %s", ty, val.ty, expr.pos)
	}
	return val, nil
}

func (ident Ident) compileCondition(asm *Assembly, block **Block, ifTrue, ifFalse *Block, scope *Scope, pos Position) error {
	label, ty, prs := scope.get(ident.name)
	if !prs {
		return fmt.Errorf("undefined variable '%s' at %s", ident.name, pos)
	}
	if ty == Int {
		return fmt.Errorf("variable '%s' at %s has type %s but is being used in condition so should be bool", ident.name, pos, ty)
	}
	(*block).emitInstruction("LDA", label)
	(*block).emitInstruction("BRZ", ifFalse.label)
	(*block).emitInstruction("BRA", ifTrue.label)
	return nil
}

func (assign Assign) compile(asm *Assembly, block **Block, scope *Scope, pos Position, errors []error) []error {
	label, ty, prs := scope.get(assign.name)
	if !prs {
		return append(errors, fmt.Errorf("cannot assign to undefined variable '%s' at %s", assign.name, pos))
	}
	value, err := compileAndExpect(assign.expr, asm, block, scope, ty)
	if err != nil {
		return append(errors, err)
	}
	loadToAcc(value, *block)
	(*block).emitInstruction("STA", label)
	return errors
}

func (decl Declare) compile(asm *Assembly, block **Block, scope *Scope, pos Position, errors []error) []error {
	value := Value{1, true, ""}
	ty := Undefined
	if decl.expr.node != nil {
		var err error
		value, err = decl.expr.compileValue(asm, block, scope)
		if err != nil {
			return append(errors, err)
		}
		if decl.ty != value.ty && decl.ty != Undefined {
			err := fmt.Errorf("expression at %s has type %s but declaration at %s has type %s", decl.expr.pos, value.ty, pos, decl.ty)
			return append(errors, err)
		}
		ty = value.ty
	} else {
		ty = decl.ty
	}

	label := scope.declare(decl.name, ty)
	asm.createVariable(label, 0)

	if decl.expr.node != nil {
		loadToAcc(value, *block)
		(*block).emitInstruction("STA", label)
	}
	return errors
}

func (statement If) compile(asm *Assembly, block **Block, scope *Scope, pos Position, errors []error) []error {
	ifTrue := asm.newUniqueBlock()
	ifFalse := asm.newUniqueBlock()
	if err := statement.cond.compileCondition(asm, block, ifTrue, ifFalse, scope); err != nil {
		return append(errors, err)
	}
	errors = statement.ifTrue.compile(asm, &ifTrue, scope, errors)
	if len(errors) > 0 {
		return errors
	}
	if statement.ifFalse.node == nil {
		ifTrue.emitInstruction("BRA", ifFalse.label)
		*block = ifFalse
		return nil
	}
	exitBlock := asm.newUniqueBlock()
	errors = statement.ifFalse.compile(asm, &ifFalse, scope, errors)
	if len(errors) > 0 {
		return errors
	}
	ifTrue.emitInstruction("BRA", exitBlock.label)
	ifFalse.emitInstruction("BRA", exitBlock.label)
	*block = exitBlock
	return nil
}

func (statement While) compile(asm *Assembly, block **Block, scope *Scope, pos Position, errors []error) []error {
	condBlock := asm.newUniqueBlock()
	loopBlock := asm.newUniqueBlock()
	exitBlock := asm.newUniqueBlock()

	(*block).emitInstruction("BRA", condBlock.label)
	if err := statement.cond.compileCondition(asm, &condBlock, loopBlock, exitBlock, scope); err != nil {
		return append(errors, err)
	}

	errors = statement.loop.compile(asm, &loopBlock, scope, errors)
	loopBlock.emitInstruction("BRA", condBlock.label)

	*block = exitBlock
	return errors
}

func (blockScope BlockScope) compile(asm *Assembly, block **Block, scope *Scope, pos Position, errors []error) []error {
	scope.pushScope()
	errors = compileStatements(blockScope.statements, asm, block, scope, errors)
	scope.popScope()
	return errors
}

func (output Output) compile(asm *Assembly, block **Block, scope *Scope, pos Position, errors []error) []error {
	val, err := compileAndExpect(output.expr, asm, block, scope, Int)
	if err != nil {
		return append(errors, err)
	}
	loadToAcc(val, *block)
	(*block).emitInstruction("OUT", "")
	return errors
}

func compileStatements(statements []Statement, asm *Assembly, block **Block, scope *Scope, errors []error) []error {
	for _, statement := range statements {
		errors = statement.compile(asm, block, scope, errors)
	}
	return errors
}

func Compile(statements []Statement) (string, []error) {
	asm := InitAssembly()
	block := asm.newBlock("start")
	scope := InitScope()
	errors := []error{}

	errors = compileStatements(statements, &asm, &block, &scope, errors)
	block.emitInstruction("HLT", "")

	if len(errors) > 0 {
		return "", errors
	}

	builder := strings.Builder{}
	asm.assemble(&builder)
	return builder.String(), []error{}
}
