package evaluator

import (
	"fmt"
	"monkey/object"
	"strconv"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=%d",
					len(args),
					1)
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"first": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=%d",
					len(args),
					1)
			}

			switch arg := args[0].(type) {
			case *object.String:
				if len(arg.Value) > 0 {
					return &object.String{Value: string(arg.Value[0])}
				}
				return NULL
			case *object.Array:
				if len(arg.Elements) > 0 {
					return arg.Elements[0]
				}
				return NULL

			default:
				return newError("argument to `first` not supported, got %s", args[0].Type())
			}
		},
	},
	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=%d",
					len(args),
					1)
			}

			switch arg := args[0].(type) {
			case *object.String:
				if len(arg.Value) > 0 {
					return &object.String{Value: string(arg.Value[len(arg.Value)-1])}
				}
				return NULL
			case *object.Array:
				if len(arg.Elements) > 0 {
					return arg.Elements[len(arg.Elements)-1]
				}
				return NULL

			default:
				return newError("argument to `last` not supported, got %s", args[0].Type())
			}
		},
	},
	"rest": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=%d",
					len(args),
					1)
			}

			switch arg := args[0].(type) {
			case *object.String:
				if len(arg.Value) > 0 {
					return &object.String{Value: arg.Value[1:len(arg.Value)]}
				}
				return NULL
			case *object.Array:
				if len(arg.Elements) > 0 {
					return &object.Array{Elements: arg.Elements[1:len(arg.Elements)]}
				}
				return NULL

			default:
				return newError("argument to `rest` not supported, got %s", args[0].Type())
			}
		},
	},
	"init": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=%d",
					len(args),
					1)
			}

			switch arg := args[0].(type) {
			case *object.String:
				if len(arg.Value) > 0 {
					return &object.String{Value: arg.Value[0 : len(arg.Value)-1]}
				}
				return NULL
			case *object.Array:
				if len(arg.Elements) > 0 {
					return &object.Array{Elements: arg.Elements[0 : len(arg.Elements)-1]}
				}
				return NULL

			default:
				return newError("argument to `init` not supported, got %s", args[0].Type())
			}
		},
	},
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=%d",
					len(args),
					1)
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)

			length := len(arr.Elements)

			newElements := make([]object.Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]
			return &object.Array{Elements: newElements}
		},
	},
	"puts": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, a := range args {
				fmt.Println(a.Inspect())
			}
			return NULL
		},
	},
	"open": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument to `push` must be STRING, got %s", args[0].Type())
			}

			filename := args[0].(*object.String)
			f := &object.File{Filename: filename.Value}

			f.Open()

			return f
		},
	},
	"read": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if args[0].Type() != object.FILE_OBJ {
				return newError("argument to `push` must be FILE, got %s", args[0].Type())
			}

			f := args[0].(*object.File)
			return f.Read()
		},
	},
	"lines": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if args[0].Type() != object.FILE_OBJ {
				return newError("argument to `push` must be FILE, got %s", args[0].Type())
			}

			f := args[0].(*object.File)
			return f.ReadAll()
		},
	},
	"close": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if args[0].Type() != object.FILE_OBJ {
				return newError("argument to `push` must be FILE, got %s", args[0].Type())
			}

			f := args[0].(*object.File)
			return f.Close()
		},
	},
	"set": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=%d",
					len(args),
					1)
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `set` must be ARRAY, got %s", args[0].Type())
			}

			if args[1].Type() != object.INTEGER_OBJ {
				return newError("index `set` into ARRAY must be INTEGER, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			idx := args[1].(*object.Integer)

			arr.Elements[idx.Value] = args[2]

			return arr
		},
	},
	"int": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument to `push` must be STRING, got %s", args[0].Type())
			}

			s := args[0].(*object.String)
			i, err := strconv.ParseInt(s.Value, 10, 64)
			if err != nil {
				return newError("string is not an int, got %s", s.Value)
			}

			return &object.Integer{Value: i}

		},
	},
}
