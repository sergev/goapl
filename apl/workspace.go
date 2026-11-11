package apl

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
)

// Interp holds the workspace and I/O of one interpreter session.
type Interp struct {
	env         // system variables
	vars        map[string]*Array
	fns         map[string]*Function
	fnGen       int                  // bumped when fns change; invalidates parse caches
	editing     *editor              // function being defined or edited, or nil
	opFrames    []map[string]*fnNode // operands of the running defined operators
	stack       []*frame             // state indicator: running functions
	noSuspend   int                  // >0 while errors must not suspend (⎕EA)
	lastErr     *Error               // last error, for ⎕EM and ⎕ET
	interrupted atomic.Bool          // set by Interrupt (Ctrl-C)
	lx          *Array               // ⎕LX latent expression
	wsid        string               // workspace name, "" for a clear workspace
	atTop       func()               // to run when the state indicator is cleared
	off         bool                 // )OFF was given
	depth       int                  // user function call depth
	out         io.Writer
	input       *bufio.Scanner // source of ⎕ input
	Prompt      string         // printed before each input line
	Echo        bool           // copy input lines to output, for transcripts
	Color       bool           // use ANSI colors in output
	pending     bool           // ⍞ output left a line unfinished
	partial     string         // the unfinished line
	// ReadLine, if set, reads input lines in place of the reader given
	// to Run, showing the prompt; an error ends the input.
	ReadLine func(prompt string) (string, error)
}

// New returns an interpreter writing to out.
func New(out io.Writer) *Interp {
	return &Interp{
		lx:   Chars(""),
		env:  defaultEnv(),
		vars: map[string]*Array{},
		fns:  map[string]*Function{},
		out:  out,
	}
}

// Exec processes one input line: a statement, a ∇ header,
// or a line of the function being defined.
func (in *Interp) Exec(line string) error {
	_, err := in.execLine(line)
	if err == errEscape || err == errUnwind || err == errReset {
		err = nil
	}
	return err
}

// execLine is Exec giving also the target of a branch, or -1.
func (in *Interp) execLine(line string) (int, error) {
	if in.editing != nil {
		return -1, withSrc(in.editLine(line), line)
	}
	if t := strings.TrimSpace(line); strings.HasPrefix(t, ")") {
		return -1, withSrc(in.command(t), line)
	}
	toks, err := Lex(line)
	if err != nil {
		return -1, withSrc(err, line)
	}
	if len(toks) > 0 && isSpecial(toks[0], "∇") {
		return -1, withSrc(in.openEditor(line), line)
	}
	if len(toks) >= 2 && toks[0].Kind == Name && isSpecial(toks[1], ":") {
		toks = toks[2:] // label is ignored outside a function
	}
	nodes, err := in.parseLine(toks, len([]rune(line)), nil)
	target := -1
	if err == nil {
		target, err = in.runLine(nodes)
	}
	return target, withSrc(err, line)
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
		in.pending, in.partial = false, ""
	}
}

// readLine reads an input line after prompt, shown in style: with
// ReadLine if set, else from the input of Run, showing the prompt (or
// echoing it with the line). It reports false at end of input.
func (in *Interp) readLine(prompt, style string) (string, bool) {
	in.endLine()
	shown := prompt
	if style != "" {
		shown = in.paint(style, prompt)
	}
	if in.ReadLine != nil {
		line, err := in.ReadLine(shown)
		return line, err == nil
	}
	if !in.Echo {
		fmt.Fprint(in.out, shown)
	}
	if in.input == nil || !in.input.Scan() {
		return "", false
	}
	line := in.input.Text()
	if in.Echo {
		fmt.Fprintln(in.out, strings.TrimRight(prompt+line, " "))
	}
	return line, true
}

// quoteOutput is ⍞←X: prints X without ending the line.
func (in *Interp) quoteOutput(v *Array) error {
	s := format(v, in.pp)
	if s == "" {
		return nil
	}
	_, err := fmt.Fprint(in.out, in.paint(styleResult, s))
	in.pending = true
	if i := strings.LastIndexByte(s, '\n'); i >= 0 {
		in.partial = s[i+1:]
	} else {
		in.partial += s
	}
	return err
}

// quoteInput is ⍞: reads a line as a character vector. Any ⍞ output
// on the same line serves as the prompt.
func (in *Interp) quoteInput() (*Array, error) {
	if in.ReadLine != nil {
		line, err := in.ReadLine(in.partial) // redraws the prompt
		in.pending, in.partial = false, ""
		if err != nil {
			return nil, newError(Interrupt)
		}
		return Chars(line), nil
	}
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
	switch name {
	case "⎕LC":
		return in.lineCounter(), true
	case "⎕EM":
		return in.errorMessage(), true
	case "⎕ET":
		return in.errorType(), true
	case "⎕LX":
		return in.lx, true
	}
	if isSysName(name) {
		return in.env.get(name), true
	}
	v, ok := in.vars[name]
	return v, ok
}

// setVar assigns a variable or system variable.
func (in *Interp) setVar(name string, v *Array) error {
	if readOnly[name] {
		return newError(SyntaxError)
	}
	if name == "⎕LX" {
		if v.Rank() > 1 || v.Size() > 0 && !v.IsCharacter() {
			return newError(DomainError)
		}
		in.lx = &Array{Shape: []int{v.Size()}, Elems: v.elems(), Proto: Char(' ')}
		return nil
	}
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
	in.repl(0)
	in.input = nil
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
		line, ok := in.readLine(in.Prompt, "")
		if !ok {
			return nil, newError(Interrupt)
		}
		line = Transliterate(line)
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
	if e, ok := err.(*Error); ok {
		in.lastErr = e
	}
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
