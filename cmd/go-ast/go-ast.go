package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/go-coverage/cmd/util/prog"
	"github.com/xhd2015/go-coverage/code"
)

type Prog struct {
	// A string `prog:"project-dir '' project directory, default current dir"`
}

var progArgs Prog

var commands = map[string]func(comm string, args []string, extraArgs []string){
	"help": help,
}

func main() {
	prog.Run(&progArgs, &prog.RunOptions{
		Usage:    usage,
		Commands: commands,
		Default:  defaultCommand,
	})
}

func help(commd string, args []string, extraArgs []string) {
	flag.Usage()
	os.Exit(0)
}

func dump(commd string, args []string, extraArgs []string) {
	if len(args) == 0 {
		fmt.Errorf("requires FILE")
		os.Exit(1)
	}
	file := args[0]

	_, ast, _, err := code.ParseFile(file)
	if err != nil {
		fmt.Printf("parsing: %v", err)
		os.Exit(1)
	}
}

func defaultCommand(commd string, args []string, extraArgs []string) {
	help(commd, args, extraArgs)
}

func usage(defaultUsage func()) func() {
	return func() {
		fmt.Sprintf(strings.Join([]string{
			"supported commands: dump\n",
			"    dump FILE\n",
			"        dump ast of given FILE\n",
			"    help\n",
			"        show help message\n",
		}, "\n"))
		defaultUsage()
		fmt.Sprintf(strings.Join([]string{
			"examples:\n",
			"    # dump:\n",
			"    $  go-ast dump a.go\n",
			"\n",
		}, "\n"))
	}
}
