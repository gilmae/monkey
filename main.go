package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gilmae/monkey/compiler"
	"github.com/gilmae/monkey/lexer"
	"github.com/gilmae/monkey/parser"
	"github.com/gilmae/monkey/repl"
	"github.com/gilmae/monkey/vm"
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
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	// evaluated := evaluator.Eval(program, env)

	// if evaluated != nil && evaluated.Type() != object.NULL_OBJ {
	// 	fmt.Printf("%s\n", evaluated.Inspect())
	// }
	comp := compiler.New()
	err := comp.Compile(program)

	if err != nil {
		fmt.Printf("Compile error:\n%s\n", err)
		return 1
	}

	machine := vm.New(comp.Bytecode())
	err = machine.Run()
	if err != nil {
		fmt.Printf("Executing bytecode failed:\n%s\n", err)
		return 1
	}

	// stackTop := machine.LastPoppedStackElem()
	// fmt.Printf("%s\n", stackTop.Inspect())

	return 0
}
