// Command goapl is an APL interpreter.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sergev/goapl/apl"
)

func main() {
	if len(os.Args) > 1 {
		for _, name := range os.Args[1:] {
			f, err := os.Open(name)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			run(f, os.Stdout)
			f.Close()
		}
		return
	}
	run(os.Stdin, os.Stdout)
}

// run reads lines until EOF or "/*".
// For now it only prints tokens; evaluation comes next.
func run(r io.Reader, w io.Writer) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "/*") {
			return
		}
		toks, err := apl.Lex(line)
		if err != nil {
			reportLexError(w, line, err)
			continue
		}
		for i, t := range toks {
			if i > 0 {
				fmt.Fprint(w, " ")
			}
			fmt.Fprint(w, t)
		}
		fmt.Fprintln(w)
	}
}

// reportLexError prints message, line and caret, like the prototype's SError.
func reportLexError(w io.Writer, line string, err error) {
	fmt.Fprintln(w, err)
	if le, ok := err.(*apl.LexError); ok {
		fmt.Fprintln(w, line)
		fmt.Fprintln(w, strings.Repeat(" ", le.Pos)+"↑")
	}
}
