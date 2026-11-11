// Command goapl is an APL interpreter.
package main

import (
	"fmt"
	"os"

	"github.com/sergev/goapl/apl"
)

func main() {
	in := apl.New(os.Stdout)
	if len(os.Args) > 1 {
		for _, name := range os.Args[1:] {
			f, err := os.Open(name)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			in.Run(f)
			f.Close()
		}
		return
	}
	if fi, err := os.Stdin.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
		in.Prompt = "      "
	}
	in.Run(os.Stdin)
}
