package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

const StackSize = 2048
const GlobalSize = 65536

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	globals []object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
		globals:      make([]object.Object, GlobalSize),
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
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
		case code.OpTrue:
			err := v.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := v.push(False)
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := v.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpGreaterThan, code.OpGreaterThanOrEqual, code.OpEqual, code.OpNotEqual:
			err := v.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := v.executeBangOperator(op)
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := v.executeMinusOperator(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			v.pop()
		case code.OpJump:
			pos := int(code.ReadUint16(v.instructions[ip+1:]))
			ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(v.instructions[ip+1:]))
			ip += 2
			condition := v.pop()
			if !isTruthy(condition) {
				ip = pos - 1
			}
		case code.OpNull:
			err := v.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(v.instructions[ip+1:])
			ip += 2
			v.globals[globalIndex] = v.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(v.instructions[ip+1:])
			ip += 2
			err := v.push(v.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			numElements := int(code.ReadUint16(v.instructions[ip+1:]))
			ip += 2

			array := v.buildArray(v.sp-numElements, v.sp)
			v.sp = v.sp - numElements

			err := v.push(array)
			if err != nil {
				return err
			}
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

func (v *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = v.stack[i]
	}

	return &object.Array{Elements: elements}
}

func (v *VM) executeBangOperator(op code.Opcode) error {
	operand := v.pop()

	switch operand {
	case True:
		return v.push(False)
	case False:
		return v.push(True)
	case Null:
		return v.push(True)
	default:
		return v.push(False)
	}
}

func (v *VM) executeBinaryOperation(op code.Opcode) error {
	right := v.pop()
	left := v.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return v.executeBinaryIntegerOperation(op, left, right)
	} else if leftType == object.STRING_OBJ && rightType == object.STRING_OBJ {
		return v.executeBinaryStringOperation(op, left, right)
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

func (v *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	var result string
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue

	default:
		return fmt.Errorf("Unknown integer operator: %d", op)
	}
	return v.push(&object.String{Value: result})

}

func (v *VM) executeComparison(op code.Opcode) error {
	right := v.pop()
	left := v.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return v.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return v.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return v.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, leftType, rightType)
	}
}

func (v *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return v.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case code.OpNotEqual:
		return v.push(nativeBoolToBooleanObject(leftValue != rightValue))
	case code.OpGreaterThan:
		return v.push(nativeBoolToBooleanObject(leftValue > rightValue))
	case code.OpGreaterThanOrEqual:
		return v.push(nativeBoolToBooleanObject(leftValue >= rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (v *VM) executeMinusOperator(op code.Opcode) error {
	operand := v.pop()
	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return v.push(&object.Integer{Value: -value})
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
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
