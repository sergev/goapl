package apl

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Interp holds the workspace and I/O of one interpreter session.
type Interp struct {
	vars   map[string]*Array
	out    io.Writer
	Prompt string // printed before each input line
}

// New returns an interpreter writing to out.
func New(out io.Writer) *Interp {
	return &Interp{vars: map[string]*Array{}, out: out}
}

// Exec runs one line, printing the result unless it is an assignment.
func (in *Interp) Exec(line string) error {
	n, err := Parse(line)
	if err != nil || n == nil {
		return err
	}
	v, err := in.eval(n)
	if err != nil {
		return err
	}
	switch n.(type) {
	case *assignNode, *indexAssignNode:
		return nil
	}
	return in.print(v)
}

func (in *Interp) print(v *Array) error {
	if v.Rank() > MaxRank {
		return newError(60)
	}
	_, err := fmt.Fprintln(in.out, Format(v))
	return err
}

// Run executes lines from r until EOF or a line starting with "/*".
func (in *Interp) Run(r io.Reader) {
	sc := bufio.NewScanner(r)
	for {
		fmt.Fprint(in.out, in.Prompt)
		if !sc.Scan() {
			return
		}
		line := sc.Text()
		if strings.HasPrefix(line, "/*") {
			return
		}
		if err := in.Exec(line); err != nil {
			in.report(line, err)
		}
	}
}

// report prints an error message, then the line with a caret, like SError.
func (in *Interp) report(line string, err error) {
	fmt.Fprintln(in.out, err)
	if e, ok := err.(*Error); ok && e.Pos != NoPos {
		fmt.Fprintln(in.out, line)
		fmt.Fprintln(in.out, strings.Repeat(" ", e.Pos)+"↑")
	}
}
