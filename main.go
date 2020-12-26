package main

import (
	"fmt"
	"monkey/repl"
	"os"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello, %s. This is the Moneky programming language\n", user.Username)
	fmt.Printf("What would you like to say?\n")
	repl.Start(os.Stdin, os.Stdout)
}
