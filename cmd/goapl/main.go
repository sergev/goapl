// Command goapl is an APL interpreter.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/sergev/goapl/apl"
)

func main() {
	tty := isTerminal(os.Stdin)
	color := isTerminal(os.Stdout) && os.Getenv("NO_COLOR") == ""
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, tty, color))
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// run executes script files, or stdin if none; returns the exit status.
// tty enables the prompt, color the ANSI colors.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer, tty, color bool) int {
	in := apl.New(stdout)
	in.Color = color
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
