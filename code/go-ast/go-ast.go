package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/go-coverage/code"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		// ioutil.ReadAll(os.Stdin)
		fmt.Printf("requires file\n")
		os.Exit(1)
	}
	file := args[1]

	_, ast, _, err := code.ParseFile(file)
	if err != nil {
		fmt.Printf("parsing: %v", err)
		os.Exit(1)
	}

	// get log
	code.Clean(ast, code.CleanOpts{
		Log:       true,
		LogIndent: 4,
	})
}
