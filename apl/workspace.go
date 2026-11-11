package apl

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Interp holds the workspace and I/O of one interpreter session.
type Interp struct {
	env      // system variables
	vars     map[string]*Array
	fns      map[string]*Function
	fnGen    int                  // bumped when fns change; invalidates parse caches
	editing  *editor              // function being defined or edited, or nil
	opFrames []map[string]*fnNode // operands of the running defined operators
	depth    int                  // user function call depth
	out      io.Writer
	input    *bufio.Scanner // source of ⎕ input
	Prompt   string         // printed before each input line
	Echo     bool           // copy input lines to output, for transcripts
	Color    bool           // use ANSI colors in output
	pending  bool           // ⍞ output left a line unfinished
}

// New returns an interpreter writing to out.
func New(out io.Writer) *Interp {
	return &Interp{
		env:  defaultEnv(),
		vars: map[string]*Array{},
		fns:  map[string]*Function{},
		out:  out,
	}
}

// Exec processes one input line: a statement, a ∇ header,
// or a line of the function being defined.
func (in *Interp) Exec(line string) error {
	if in.editing != nil {
		return withSrc(in.editLine(line), line)
	}
	toks, err := Lex(line)
	if err != nil {
		return withSrc(err, line)
	}
	if len(toks) > 0 && isSpecial(toks[0], "∇") {
		return withSrc(in.openEditor(line), line)
	}
	if len(toks) >= 2 && toks[0].Kind == Name && isSpecial(toks[1], ":") {
		toks = toks[2:] // label is ignored outside a function
	}
	nodes, err := in.parseLine(toks, len([]rune(line)), nil)
	if err == nil {
		_, err = in.runLine(nodes) // a branch at top level is ignored
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
	in.endLine()
	_, err := fmt.Fprintln(in.out, in.paint(styleResult, wrap(format(v, in.pp), in.pw)))
	return err
}

// endLine finishes a line left open by ⍞ output.
func (in *Interp) endLine() {
	if in.pending {
		fmt.Fprintln(in.out)
		in.pending = false
	}
}

// quoteOutput is ⍞←X: prints X without ending the line.
func (in *Interp) quoteOutput(v *Array) error {
	s := format(v, in.pp)
	if s == "" {
		return nil
	}
	in.endLine()
	_, err := fmt.Fprint(in.out, in.paint(styleResult, s))
	in.pending = true
	return err
}

// quoteInput is ⍞: reads a line as a character vector. Any ⍞ output
// on the same line serves as the prompt.
func (in *Interp) quoteInput() (*Array, error) {
	if in.input == nil || !in.input.Scan() {
		in.endLine()
		return nil, newError(Interrupt)
	}
	line := in.input.Text()
	switch {
	case in.Echo:
		fmt.Fprintln(in.out, line)
		in.pending = false
	case in.Prompt != "":
		in.pending = false // the terminal echoed the line
	}
	return Chars(line), nil
}

// getVar returns the value of a variable or system variable.
func (in *Interp) getVar(name string) (*Array, bool) {
	if isSysName(name) {
		return in.env.get(name), true
	}
	v, ok := in.vars[name]
	return v, ok
}

// setVar assigns a variable or system variable.
func (in *Interp) setVar(name string, v *Array) error {
	if isSysName(name) {
		return in.env.set(name, v)
	}
	in.vars[name] = v
	return nil
}

// Run executes lines from r until EOF or a line starting with "/*".
// The same reader supplies ⎕ input.
func (in *Interp) Run(r io.Reader) {
	in.input = bufio.NewScanner(r)
	for {
		if !in.Echo {
			in.prompt()
		}
		if !in.input.Scan() {
			in.endLine()
			return
		}
		line := in.input.Text()
		if in.Echo {
			in.endLine()
			fmt.Fprintln(in.out, strings.TrimRight(in.promptText()+line, " "))
		}
		if strings.HasPrefix(strings.TrimSpace(line), "/*") {
			in.endLine()
			return
		}
		if err := in.Exec(line); err != nil {
			in.report(err)
		}
	}
}

func (in *Interp) prompt() {
	in.endLine()
	p := in.promptText()
	if in.editing != nil {
		p = in.paint(stylePrompt, p)
	}
	fmt.Fprint(in.out, p)
}

// promptText is the prompt, or "[n] " while defining a function.
func (in *Interp) promptText() string {
	if in.Prompt != "" && in.editing != nil {
		return "[" + lineNum(in.editing.cur) + "] "
	}
	return in.Prompt
}

// quadInput reads and evaluates a line for ⎕, prompting again on error.
func (in *Interp) quadInput() (*Array, error) {
	for {
		in.endLine()
		fmt.Fprintln(in.out, in.paint(stylePrompt, "⎕:"))
		if in.input == nil || !in.input.Scan() {
			return nil, newError(Interrupt)
		}
		line := in.input.Text()
		if in.Echo {
			fmt.Fprintln(in.out, strings.TrimRight(in.Prompt+line, " "))
		}
		toks, err := Lex(line)
		if err == nil && len(toks) == 0 {
			continue
		}
		var n node
		if err == nil {
			n, err = in.parse(toks, len([]rune(line)), nil)
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
	in.endLine()
	fmt.Fprintln(in.out, in.paint(styleError, err.Error()))
	e, ok := err.(*Error)
	if !ok || e.Src == "" {
		return
	}
	prefix := ""
	if e.Where != "" {
		prefix = e.Where + " "
	}
	fmt.Fprintln(in.out, in.paint(styleWhere, prefix)+e.Src)
	if e.Pos != NoPos {
		fmt.Fprintln(in.out, strings.Repeat(" ", len([]rune(prefix))+e.Pos)+in.paint(styleCaret, "^"))
	}
}
