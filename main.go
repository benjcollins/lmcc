package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	debug := flag.Bool("debug", false, "whether to output the AST")
	outputPath := flag.String("output", "output.txt", "where to write the output to")
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println("no source file")
		return
	}
	path := flag.Args()[0]

	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	ast, parseErrors := Parse(string(data))
	if len(parseErrors) > 0 {
		for _, err := range parseErrors {
			fmt.Println(err)
		}
		return
	}

	if *debug {
		builder := strings.Builder{}
		for _, stmt := range ast {
			stmt.prettyPrint(&builder, "")
		}
		fmt.Print(builder.String())
	}

	output, errors := Compile(ast)
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println(err)
		}
		return
	}

	if err := ioutil.WriteFile(*outputPath, []byte(output), 0644); err != nil {
		panic(err)
	}
}
