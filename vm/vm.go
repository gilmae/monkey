package vm

import (
	"fmt"

	"github.com/gilmae/monkey/code"
	"github.com/gilmae/monkey/compiler"
	"github.com/gilmae/monkey/object"
)

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

const StackSize = 2048
const GlobalSize = 65536
const MaxFrames = 1024

type VM struct {
	constants []object.Object

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	globals []object.Object

	frames      []*Frame
	framesIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: make([]object.Object, GlobalSize),

		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (v *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for v.currentFrame().ip < len(v.currentFrame().Instructions())-1 {
		v.currentFrame().ip++

		ip = v.currentFrame().ip
		ins = v.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			v.currentFrame().ip += 2

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
			pos := int(code.ReadUint16(ins[ip+1:]))
			v.currentFrame().ip = pos - 1

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			v.currentFrame().ip += 2
			condition := v.pop()
			if !isTruthy(condition) {
				v.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			err := v.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			v.currentFrame().ip += 2
			v.globals[globalIndex] = v.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			v.currentFrame().ip += 2
			err := v.push(v.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			v.currentFrame().ip += 2

			array := v.buildArray(v.sp-numElements, v.sp)
			v.sp = v.sp - numElements

			err := v.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			v.currentFrame().ip += 2

			hash, err := v.buildHash(v.sp-numElements, v.sp)
			if err != nil {
				return err
			}

			v.sp = v.sp - numElements

			err = v.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := v.pop()
			left := v.pop()

			err := v.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case code.OpCall:
			numArgs := code.ReadUint8((ins[ip+1:]))
			v.currentFrame().ip += 1

			err := v.executeCall(int(numArgs))

			if err != nil {
				return err
			}
		case code.OpReturnValue:
			returnValue := v.pop()
			frame := v.popFrame()
			v.sp = frame.basePointer - 1
			err := v.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := v.popFrame()
			v.sp = frame.basePointer - 1
			err := v.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			v.currentFrame().ip += 1

			frame := v.currentFrame()

			v.stack[frame.basePointer+int(localIndex)] = v.pop()

		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			v.currentFrame().ip += 1

			frame := v.currentFrame()

			err := v.push(v.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(ins[ip+1:])
			v.currentFrame().ip += 1
			definition := object.Builtins[builtinIndex]

			err := v.push(definition.Builtin)
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

func (v *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := v.stack[i]
		value := v.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

func (v *VM) callFunction(fn *object.CompiledFunction, numArgs int) error {
	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", fn.NumParameters, numArgs)
	}

	frame := NewFrame(fn, v.sp-numArgs)
	v.pushFrame(frame)
	v.sp = frame.basePointer + fn.NumLocals

	return nil
}

func (v *VM) callBuiltin(fn *object.Builtin, numArgs int) error {
	args := v.stack[v.sp-numArgs : v.sp]
	result := fn.Fn(args...)

	v.sp = v.sp - numArgs - 1

	if result != nil {
		v.push(result)
	} else {
		v.push(Null)
	}

	return nil
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

func (v *VM) executeCall(numArgs int) error {
	callee := v.stack[v.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.CompiledFunction:
		return v.callFunction(callee, numArgs)
	case *object.Builtin:
		return v.callBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
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

func (v *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return v.executeArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return v.executeHashIndexExpression(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s[%s]", left.Type(), index.Type())
	}
}

func (v *VM) executeArrayIndexExpression(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements)) - 1

	if i < 0 || i > max {
		return v.push(Null)
	}

	return v.push(arrayObject.Elements[i])
}

func (v *VM) executeHashIndexExpression(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return v.push(Null)
	}

	return v.push(pair.Value)

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

func (v *VM) currentFrame() *Frame {
	return v.frames[v.framesIndex-1]
}

func (v *VM) pushFrame(f *Frame) {
	v.frames[v.framesIndex] = f
	v.framesIndex++
}

func (v *VM) popFrame() *Frame {
	v.framesIndex--
	return v.frames[v.framesIndex]
}
