package object

import (
	"fmt"
	"os"
	"strconv"
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
		},
	},
	{
		"puts",
		&Builtin{
			Fn: func(args ...Object) Object {
				for _, a := range args {
					fmt.Println(a.Inspect())
				}
				return nil
			},
		},
	},
	{
		"first",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=%d",
						len(args),
						1)
				}

				switch arg := args[0].(type) {
				case *String:
					if len(arg.Value) > 0 {
						return &String{Value: string(arg.Value[0])}
					}
					return nil
				case *Array:
					if len(arg.Elements) > 0 {
						return arg.Elements[0]
					}
					return nil

				default:
					return newError("argument to `first` not supported, got %s", args[0].Type())
				}
			},
		},
	},
	{
		"last",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=%d",
						len(args),
						1)
				}

				switch arg := args[0].(type) {
				case *String:
					if len(arg.Value) > 0 {
						return &String{Value: string(arg.Value[len(arg.Value)-1])}
					}
					return nil
				case *Array:
					if len(arg.Elements) > 0 {
						return arg.Elements[len(arg.Elements)-1]
					}
					return nil

				default:
					return newError("argument to `last` not supported, got %s", args[0].Type())
				}
			},
		},
	},
	{
		"rest",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=%d",
						len(args),
						1)
				}

				switch arg := args[0].(type) {
				case *String:
					if len(arg.Value) > 0 {
						return &String{Value: arg.Value[1:len(arg.Value)]}
					}
					return nil
				case *Array:
					if len(arg.Elements) > 0 {
						return &Array{Elements: arg.Elements[1:len(arg.Elements)]}
					}
					return nil

				default:
					return newError("argument to `rest` not supported, got %s", args[0].Type())
				}
			},
		},
	},
	{
		"init",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=%d",
						len(args),
						1)
				}

				switch arg := args[0].(type) {
				case *String:
					if len(arg.Value) > 0 {
						return &String{Value: arg.Value[0 : len(arg.Value)-1]}
					}
					return nil
				case *Array:
					if len(arg.Elements) > 0 {
						return &Array{Elements: arg.Elements[0 : len(arg.Elements)-1]}
					}
					return nil

				default:
					return newError("argument to `init` not supported, got %s", args[0].Type())
				}
			},
		},
	},
	{
		"push",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=%d",
						len(args),
						1)
				}

				if args[0].Type() != ARRAY_OBJ {
					return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
				}

				arr := args[0].(*Array)

				length := len(arr.Elements)

				newElements := make([]Object, length+1, length+1)
				copy(newElements, arr.Elements)
				newElements[length] = args[1]
				return &Array{Elements: newElements}
			},
		},
	},
	{
		"open",
		&Builtin{
			Fn: func(args ...Object) Object {
				if args[0].Type() != STRING_OBJ {
					return newError("argument to `open` must be STRING, got %s", args[0].Type())
				}

				filename := args[0].(*String)
				f := &File{Filename: filename.Value}

				f.Open()

				return f
			},
		},
	},
	{
		"read",
		&Builtin{
			Fn: func(args ...Object) Object {
				if args[0].Type() != FILE_OBJ {
					return newError("argument to `read` must be FILE, got %s", args[0].Type())
				}

				f := args[0].(*File)
				return f.Read()
			},
		},
	},
	{
		"lines",
		&Builtin{
			Fn: func(args ...Object) Object {
				if args[0].Type() != FILE_OBJ {
					return newError("argument to `push` must be FILE, got %s", args[0].Type())
				}

				f := args[0].(*File)
				return f.ReadAll()
			},
		},
	},
	{
		"close",
		&Builtin{
			Fn: func(args ...Object) Object {
				if args[0].Type() != FILE_OBJ {
					return newError("argument to `push` must be FILE, got %s", args[0].Type())
				}

				f := args[0].(*File)
				return f.Close()
			},
		},
	},
	{
		"set",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 3 {
					return newError("wrong number of arguments. got=%d, want=%d",
						len(args),
						1)
				}

				if args[0].Type() != ARRAY_OBJ {
					return newError("argument to `set` must be ARRAY, got %s", args[0].Type())
				}

				if args[1].Type() != INTEGER_OBJ {
					return newError("index `set` into ARRAY must be INTEGER, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				idx := args[1].(*Integer)

				arr.Elements[idx.Value] = args[2]

				return arr
			},
		},
	},
	{
		"int",
		&Builtin{
			Fn: func(args ...Object) Object {
				if args[0].Type() != STRING_OBJ {
					return newError("argument to `push` must be STRING, got %s", args[0].Type())
				}

				s := args[0].(*String)
				i, err := strconv.ParseInt(s.Value, 10, 64)
				if err != nil {
					return newError("string is not an int, got %s", s.Value)
				}

				return &Integer{Value: i}

			},
		},
	},
	{
		"exit",
		&Builtin{
			Fn: func(args ...Object) Object {
				os.Exit(0)
				return nil
			},
		},
	},
}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}
