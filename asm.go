package main

import (
	"fmt"
	"io"
)

type Assembly struct {
	blocks       []*Block
	constants    map[int]bool
	maxTemp      int
	currentTemp  int
	currentBlock int
}

type Block struct {
	label string
	insts []Instruction
}

type Instruction struct {
	opcode  string
	operand string
}

func (asm *Assembly) newBlock(label string) *Block {
	block := &Block{label, []Instruction{}}
	asm.blocks = append(asm.blocks, block)
	return block
}

func (asm *Assembly) newUniqueBlock() *Block {
	block := asm.newBlock(fmt.Sprintf("b%d", asm.currentBlock))
	asm.currentBlock++
	return block
}

func (asm *Assembly) createVariable(label string, value int) {
	block := asm.newBlock(label)
	block.emitInstruction("DAT", fmt.Sprint(value))
}

func (asm *Assembly) getConstant(value int) string {
	if _, contains := asm.constants[value]; !contains {
		asm.createVariable(fmt.Sprintf("c%d", value), value)
		asm.constants[value] = true
	}
	return "c" + fmt.Sprint(value)
}

func (asm *Assembly) pushTemp() string {
	label := "temp" + fmt.Sprint(asm.currentTemp)
	if asm.currentTemp == asm.maxTemp {
		asm.maxTemp++
		asm.createVariable(label, 0)
	}
	asm.currentTemp++
	return label
}

func (asm *Assembly) popTemp() {
	asm.currentTemp--
}

func (block *Block) emitInstruction(opcode string, operand string) {
	block.insts = append(block.insts, Instruction{opcode, operand})
}

func (asm *Assembly) assemble(w io.Writer) {
	for _, block := range asm.blocks {
		block.assemble(w)
	}
}

func (block Block) assemble(w io.Writer) {
	fmt.Fprintf(w, "%s", block.label)
	for _, inst := range block.insts {
		fmt.Fprintf(w, "\t%s %s\n", inst.opcode, inst.operand)
	}
}

func InitAssembly() Assembly {
	return Assembly{
		constants: make(map[int]bool),
	}
}
