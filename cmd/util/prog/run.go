package prog

import (
	"flag"
	"fmt"
	"os"
)

type RunOptions struct {
	Usage func(usage func()) func()

	AfterFlagParse func()

	Commands map[string]func(commd string, args []string, extraArgs []string)
	Default  func(commd string, args []string, extraArgs []string) // default command
}

func Run(prog interface{}, opts *RunOptions) {
	Bind(prog)
	if opts == nil {
		opts = &RunOptions{}
	}

	arg0 := os.Args[0]
	args := os.Args[1:]
	commd := ""
	if len(args) > 0 {
		commd = args[0]
		args = args[1:]
	}

	// other args
	var extraArgs []string
	n := len(args)
	for i := 0; i < n; i++ {
		if args[i] == "--" {
			// modify extraArgs first
			if i < n-1 {
				extraArgs = args[i+1:]
			}
			args = args[:i]
			break
		}
	}

	os.Args = append([]string{arg0}, args...)
	flag.Parse()
	args = flag.Args()

	// set usage
	if opts.Usage != nil {
		flag.Usage = opts.Usage(flag.Usage)
	}

	if opts.AfterFlagParse != nil {
		opts.AfterFlagParse()
	}

	handler := opts.Commands[commd]
	if handler == nil {
		handler = opts.Default
	}
	if handler == nil {
		panic(fmt.Errorf("unrecognized command: %s", commd))
	}
	handler(commd, args, extraArgs)
}
