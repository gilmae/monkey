package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (v *VM) Run() error {
	for ip := 0; ip < len(v.instructions); ip++ {
		op := code.Opcode(v.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(v.instructions[ip+1:])
			ip += 2

			err := v.push(v.constants[constIndex])
			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := v.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			v.pop()
		}
	}
	return nil
}

func (v *VM) LastPoppedStackElem() object.Object {
	return v.stack[v.sp]
}

func (v *VM) StackTop() object.Object {
	if v.sp == 0 {
		return nil
	}
	return v.stack[v.sp-1]
}

func (v *VM) executeBinaryOperation(op code.Opcode) error {
	right := v.pop()
	left := v.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return v.executeBinaryIntegerOperation(op, left, right)
	}
	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (v *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("Unknown integer operator: %d", op)
	}
	return v.push(&object.Integer{Value: result})

}

func (v *VM) pop() object.Object {
	o := v.stack[v.sp-1]
	v.sp--
	return o
}

func (v *VM) push(obj object.Object) error {
	if v.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	v.stack[v.sp] = obj
	v.sp++
	return nil
}
