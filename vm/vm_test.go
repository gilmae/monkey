package vm

import (
	"fmt"
	"monkey/ast"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func TestBooleanExpression(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1<2", true},
		{"1>2", false},
		{"1<1", false},
		{"1>1", false},

		{"1==1", true},
		{"1!=1", false},
		{"1==2", false},
		{"1!=2", true},

		{"true==true", true},
		{"false==false", true},
		{"true==false", false},
		{"true!=false", true},
		{"false!=true", true},

		{"(1<2) == true", true},
		{"(1<2) == false", false},
		{"(1>2) == true", false},
		{"(1>2) == false", true},

		{"1<=2", true},
		{"1<=1", true},
		{"2>=1", true},
		{"2>=2", true},

		{"2>=3", false},
		{"1<=0", false},
	}

	runVmTests(t, tests)
}
func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1+2", 3},
		{"1-2", -1},
		{"2*2", 4},
		{"6/2", 3},
		{"50/2*2+10-5", 55},
	}

	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		comp := compiler.New()
		err := comp.Compile(program)

		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()

		if err != nil {
			t.Fatalf("vm error: %s", err)
		}
		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)

	if !ok {
		return fmt.Errorf("object is not Boolean, got %t (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value, expected %t, got %t", expected, result.Value)
	}

	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)

	if !ok {
		return fmt.Errorf("object is not Integer, got %t (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value, expected %d, got %d", expected, result.Value)
	}

	return nil
}
