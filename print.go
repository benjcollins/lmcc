package main

import (
	"fmt"
	"strings"
)

func (expr Expr) prettyPrint(builder *strings.Builder) {
	expr.node.prettyPrint(builder)
}

func (literal IntLiteral) prettyPrint(builder *strings.Builder) {
	fmt.Fprintf(builder, "%d", literal.value)
}

func (literal BoolLiteral) prettyPrint(builder *strings.Builder) {
	if literal.value {
		fmt.Fprintf(builder, "true")
	} else {
		fmt.Fprintf(builder, "false")
	}
}

func (input Input) prettyPrint(builder *strings.Builder) {
	fmt.Fprint(builder, "in")
}

func (ident Ident) prettyPrint(builder *strings.Builder) {
	fmt.Fprintf(builder, "%s", ident.name)
}

func (bin Binary) prettyPrint(builder *strings.Builder) {
	fmt.Fprint(builder, "(")
	bin.left.prettyPrint(builder)
	fmt.Fprintf(builder, " %s ", bin.symbol)
	bin.right.prettyPrint(builder)
	fmt.Fprint(builder, ")")
}

func (unary Unary) prettyPrint(builder *strings.Builder) {
	fmt.Fprint(builder, "(")
	fmt.Fprintf(builder, "%s ", unary.symbol)
	unary.expr.prettyPrint(builder)
	fmt.Fprint(builder, ")")
}

func (stmt Statement) prettyPrint(builder *strings.Builder, indent string) {
	fmt.Fprint(builder, indent)
	stmt.node.prettyPrint(builder, indent)
	fmt.Fprint(builder, "\n")
}

func (decl Declare) prettyPrint(builder *strings.Builder, indent string) {
	fmt.Fprintf(builder, "%s :", decl.name)
	if decl.ty != Undefined {
		fmt.Fprintf(builder, " %s ", decl.ty)
	}
	if decl.expr.node != nil {
		fmt.Fprintf(builder, "= ")
		decl.expr.prettyPrint(builder)
	}
}

func (assign Assign) prettyPrint(builder *strings.Builder, indent string) {
	fmt.Fprintf(builder, "%s = ", assign.name)
	assign.expr.prettyPrint(builder)
}

func (stmt If) prettyPrint(builder *strings.Builder, indent string) {
	fmt.Fprint(builder, "if ")
	stmt.cond.prettyPrint(builder)
	fmt.Fprint(builder, "\n")
	stmt.ifTrue.prettyPrint(builder, indent)
	if stmt.ifFalse.node != nil {
		fmt.Fprint(builder, "else\n")
		stmt.ifFalse.prettyPrint(builder, indent)
	}
}

func (stmt While) prettyPrint(builder *strings.Builder, indent string) {
	fmt.Fprint(builder, "if ")
	stmt.cond.prettyPrint(builder)
	fmt.Fprint(builder, "\n")
	stmt.loop.prettyPrint(builder, indent)
}

func (blockScope BlockScope) prettyPrint(builder *strings.Builder, indent string) {
	fmt.Fprintln(builder, "{")
	for _, stmt := range blockScope.statements {
		stmt.prettyPrint(builder, indent+"    ")
	}
	fmt.Fprint(builder, "}")
}

func (output Output) prettyPrint(builder *strings.Builder, indent string) {
	fmt.Fprint(builder, "out ")
	output.expr.prettyPrint(builder)
}
