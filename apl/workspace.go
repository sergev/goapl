package apl

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Interp holds the workspace and I/O of one interpreter session.
type Interp struct {
	vars     map[string]*Array
	fns      map[string]*Function
	fnGen    int       // bumped when fns change; invalidates parse caches
	defining *Function // function being defined, or nil
	depth    int       // user function call depth
	out      io.Writer
	input    *bufio.Scanner // source of ⎕ input
	Prompt   string         // printed before each input line
}

// New returns an interpreter writing to out.
func New(out io.Writer) *Interp {
	return &Interp{
		vars: map[string]*Array{},
		fns:  map[string]*Function{},
		out:  out,
	}
}

// Exec processes one input line: a statement, a ∇ header,
// or a line of the function being defined.
func (in *Interp) Exec(line string) error {
	if in.defining != nil {
		return withSrc(in.defineLine(line), line)
	}
	toks, err := Lex(line)
	if err != nil {
		return withSrc(err, line)
	}
	if len(toks) > 0 && isSpecial(toks[0], "∇") {
		return withSrc(in.defineHeader(line, toks), line)
	}
	if len(toks) >= 2 && toks[0].Kind == Name && isSpecial(toks[1], ":") {
		toks = toks[2:] // label is ignored outside a function
	}
	n, err := in.parse(toks, len([]rune(line)))
	if err == nil {
		_, err = in.run(n) // a branch at top level is ignored
	}
	return withSrc(err, line)
}

// withSrc records the source line of an error, unless already recorded.
func withSrc(err error, src string) error {
	if e, ok := err.(*Error); ok && e.Src == "" {
		e.Src = src
	}
	return err
}

func (in *Interp) print(v *Array) error {
	if v.Rank() > MaxRank {
		return newError(60)
	}
	_, err := fmt.Fprintln(in.out, Format(v))
	return err
}

// Run executes lines from r until EOF or a line starting with "/*".
// The same reader supplies ⎕ input.
func (in *Interp) Run(r io.Reader) {
	in.input = bufio.NewScanner(r)
	for {
		in.prompt()
		if !in.input.Scan() {
			return
		}
		line := in.input.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "/*") {
			return
		}
		if err := in.Exec(line); err != nil {
			in.report(err)
		}
	}
}

func (in *Interp) prompt() {
	if in.Prompt == "" {
		return
	}
	if in.defining != nil {
		fmt.Fprintf(in.out, "[%d] ", len(in.defining.Stmts)+1)
		return
	}
	fmt.Fprint(in.out, in.Prompt)
}

// quadInput reads and evaluates a line for ⎕, prompting again on error.
func (in *Interp) quadInput() (*Array, error) {
	for {
		fmt.Fprintln(in.out, "⎕:")
		if in.input == nil || !in.input.Scan() {
			return nil, newError(18)
		}
		line := in.input.Text()
		toks, err := Lex(line)
		if err == nil && len(toks) == 0 {
			continue
		}
		var n node
		if err == nil {
			n, err = in.parse(toks, len([]rune(line)))
		}
		var v *Array
		if err == nil {
			v, err = in.eval(n)
		}
		if err != nil {
			in.report(withSrc(err, line))
			continue
		}
		return v, nil
	}
}

// report prints an error message, then the line with a caret, like SError.
func (in *Interp) report(err error) {
	fmt.Fprintln(in.out, err)
	e, ok := err.(*Error)
	if !ok || e.Src == "" {
		return
	}
	prefix := ""
	if e.Where != "" {
		prefix = e.Where + " "
	}
	fmt.Fprintln(in.out, prefix+e.Src)
	if e.Pos != NoPos {
		fmt.Fprintln(in.out, strings.Repeat(" ", len([]rune(prefix))+e.Pos)+"↑")
	}
}
