// Command goapl is an APL interpreter.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/sergev/goapl/apl"
)

func main() {
	tty := false
	if fi, err := os.Stdin.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
		tty = true
	}
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, tty))
}

// run executes script files, or stdin if none; returns the exit status.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer, tty bool) int {
	in := apl.New(stdout)
	if len(args) == 0 {
		if tty {
			in.Prompt = "      "
		}
		in.Run(stdin)
		return 0
	}
	for _, name := range args {
		f, err := os.Open(name)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		in.Run(f)
		f.Close()
	}
	return 0
}
