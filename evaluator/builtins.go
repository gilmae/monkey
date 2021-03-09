package evaluator

import (
	"github.com/gilmae/monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len":   object.GetBuiltinByName("len"),
	"first": object.GetBuiltinByName("first"),
	"last":  object.GetBuiltinByName("last"),
	"rest":  object.GetBuiltinByName("rest"),
	"init":  object.GetBuiltinByName("init"),
	"push":  object.GetBuiltinByName("push"),
	"puts":  object.GetBuiltinByName("puts"),
	"open":  object.GetBuiltinByName("open"),
	"read":  object.GetBuiltinByName("read"),
	"lines": object.GetBuiltinByName("lines"),
	"close": object.GetBuiltinByName("close"),
	"set":   object.GetBuiltinByName("set"),
	"int":   object.GetBuiltinByName("int"),
	"exit":  object.GetBuiltinByName("exit"),
}
