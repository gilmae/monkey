package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/repl"
)

const version = "1.0.1"

func main() {
	// Set up flags
	showVersion := flag.Bool("version", false, "Show our version and exit.")
	startRepl := flag.Bool("repl", false, "Start the Monkey REPL.")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Monkey v%s", version)
		os.Exit(1)
	}

	var err error
	var input []byte
	if *startRepl {
		fmt.Printf("Monkey v%s\n", version)
		repl.Start(os.Stdin, os.Stdout)
	} else if len(flag.Args()) > 0 {
		input, err = ioutil.ReadFile(os.Args[1])
	} else {
		input, err = ioutil.ReadAll(os.Stdin)
	}

	if err == nil {
		execute(string(input))
	} else {
		fmt.Printf("Error reading: %s\n", err.Error())
	}
}

func execute(input string) int {
	env := object.NewEnvironment()
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	evaluator.Eval(program, env)
	return 0
}
