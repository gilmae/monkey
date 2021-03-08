package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
		{OpGetLocal, []int{255}, []byte{byte(OpGetLocal), 255}},
		{OpClosure, []int{65535, 254}, []byte{byte(OpClosure), 255, 255, 254}},
	}

	for _, tt := range tests {
		instruction := Make(tt.op, tt.operands...)
		if len(instruction) != len(tt.expected) {
			t.Errorf("instruction has wrong length, want %d, got %d",
				len(tt.expected),
				len(instruction))
		}

		for i, b := range tt.expected {
			if instruction[i] != tt.expected[i] {
				t.Errorf("wrong byte at pos %d, wanted %d, got %d",
					i,
					b,
					instruction[i])
			}
		}
	}
}

func TestInstructionStrings(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
		Make(OpPop),
		Make(OpSub),
		Make(OpMul),
		Make(OpDiv),
		Make(OpTrue),
		Make(OpFalse),
		Make(OpEqual),
		Make(OpNotEqual),
		Make(OpGreaterThan),
		Make(OpGreaterThanOrEqual),
		Make(OpMinus),
		Make(OpBang),
		Make(OpJumpNotTruthy, 25),
		Make(OpJump, 25),
		Make(OpNull),
		Make(OpGetGlobal, 1),
		Make(OpSetGlobal, 1),
		Make(OpArray, 10),
		Make(OpHash, 10),
		Make(OpIndex),
		Make(OpCall, 2),
		Make(OpReturnValue),
		Make(OpReturn),
		Make(OpGetLocal, 1),
		Make(OpSetLocal, 1),
		Make(OpClosure, 65535, 255),
	}

	expected := `0000 OpAdd
0001 OpConstant 2
0004 OpConstant 65535
0007 OpPop
0008 OpSub
0009 OpMul
0010 OpDiv
0011 OpTrue
0012 OpFalse
0013 OpEqual
0014 OpNotEqual
0015 OpGreaterThan
0016 OpGreaterThanOrEqual
0017 OpMinus
0018 OpBang
0019 OpJumpNotTruthy 25
0022 OpJump 25
0025 OpNull
0026 OpGetGlobal 1
0029 OpSetGlobal 1
0032 OpArray 10
0035 OpHash 10
0038 OpIndex
0039 OpCall 2
0041 OpReturnValue
0042 OpReturn
0043 OpGetLocal 1
0045 OpSetLocal 1
0047 OpClosure 65535 255
`

	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted.\n\twant:%q\n\tgot: %q", expected, concatted.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
		{OpGetLocal, []int{255}, 1},
		{OpClosure, []int{65535, 255}, 3},
	}

	for _, tt := range tests {
		instruction := Make(tt.op, tt.operands...)
		def, err := Lookup(byte(tt.op))
		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}
		operandsRead, n := ReadOperands(def, instruction[1:])
		if n != tt.bytesRead {
			t.Fatalf("n wrong, want %d, got %d", tt.bytesRead, n)
		}

		for i, want := range tt.operands {
			if operandsRead[i] != want {
				t.Errorf("operand at position %d wrong, want %d, got %d", i, want, operandsRead[i])
			}
		}
	}
}
