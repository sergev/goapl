package apl

import (
	"errors"
	"fmt"
	"strings"
)

// Control signals that unwind the state indicator. They are not APL
// errors and are never reported.
var (
	errEscape = errors.New("escape") // lone →: leave the current suspension
	errUnwind = errors.New("unwind") // a suspended function was abandoned
	errReset  = errors.New("reset")  // )RESET: clear the state indicator
	errQuit   = errors.New("quit")   // end of input or /*
)

// frame is a running user function in the state indicator.
type frame struct {
	f         *Function
	pc        int
	suspended bool
}

// canSuspend reports whether an error in a function suspends it: only
// for APL errors in a session, and not under ⎕EA.
func (in *Interp) canSuspend(err error) bool {
	_, ok := err.(*Error)
	return ok && in.input != nil && in.noSuspend == 0
}

// suspend reports err and runs a session in the context of the function
// on top of the state indicator. It gives the line to resume at, or an
// error to unwind with.
func (in *Interp) suspend(err error) (int, error) {
	in.report(err)
	fr := in.stack[len(in.stack)-1]
	fr.suspended = true
	defer func() { fr.suspended = false }()
	target, err := in.repl(len(in.stack))
	if err == errEscape {
		err = errUnwind
	}
	return target, err
}

// repl reads and runs lines. At level 0 it is the session; above, it
// runs while a function is suspended and returns the target of a branch.
func (in *Interp) repl(level int) (int, error) {
	for {
		if !in.Echo {
			in.prompt()
		}
		if !in.input.Scan() {
			in.endLine()
			return 0, errQuit
		}
		in.interrupted.Store(false) // Ctrl-C at the prompt
		line := in.input.Text()
		if in.Echo {
			in.endLine()
			fmt.Fprintln(in.out, strings.TrimRight(in.promptText()+line, " "))
		}
		if in.editing == nil && strings.HasPrefix(strings.TrimSpace(line), "/*") {
			in.endLine()
			return 0, errQuit
		}
		target, err := in.execLine(line)
		switch {
		case err == errQuit, err == errReset && level > 0:
			return 0, err
		case err == errEscape && level > 0:
			return 0, err
		case err == errReset && in.atTop != nil:
			in.atTop()
			in.atTop = nil
		case err == errEscape, err == errUnwind, err == errReset:
		case err != nil:
			in.report(err)
		case target >= 0 && level > 0:
			return target, nil
		}
	}
}

// lineCounter is ⎕LC: the line numbers of the functions in the state
// indicator, most recent first.
func (in *Interp) lineCounter() *Array {
	r := make([]float64, len(in.stack))
	for i, fr := range in.stack {
		r[len(r)-1-i] = float64(fr.pc)
	}
	return Vector(r)
}

// errorTypes gives ⎕ET for each class.
var errorTypes = map[Class][2]float64{
	Interrupt: {1, 1}, SystemLimit: {1, 4}, NonceError: {1, 6},
	SyntaxError: {2, 1}, DefnError: {2, 2}, ValueError: {3, 1},
	ValenceError: {5, 1}, RankError: {5, 2}, LengthError: {5, 3},
	DomainError: {5, 4}, IndexError: {5, 5}, AxisError: {5, 6},
}

// errorLines are the lines that report an error.
func errorLines(e *Error) []string {
	lines := []string{e.Error()}
	if e.Src == "" {
		return lines
	}
	prefix := ""
	if e.Where != "" {
		prefix = e.Where + " "
	}
	lines = append(lines, prefix+e.Src)
	if e.Pos != NoPos {
		lines = append(lines, strings.Repeat(" ", len([]rune(prefix))+e.Pos)+"^")
	}
	return lines
}

// errorMessage is ⎕EM: the report of the last error.
func (in *Interp) errorMessage() *Array {
	if in.lastErr == nil {
		return &Array{Shape: []int{3, 0}, Elems: []Value{}, Proto: Char(' ')}
	}
	lines := errorLines(in.lastErr)
	return toChars(lines, []int{len(lines)})
}

// errorType is ⎕ET: the type of the last error, 0 0 if none.
func (in *Interp) errorType() *Array {
	if in.lastErr == nil {
		return Vector([]float64{0, 0})
	}
	t := errorTypes[in.lastErr.Class]
	return Vector(t[:])
}

// executeAlternate is A ⎕EA B: executes B; on an error, records it in
// ⎕EM and ⎕ET and executes A instead.
func (in *Interp) executeAlternate(a, b *Array) (*Array, error) {
	in.noSuspend++
	r, err := in.execute(b)
	in.noSuspend--
	if e, ok := err.(*Error); ok {
		in.lastErr = e
		return in.execute(a)
	}
	return r, err
}

// Interrupt asks the interpreter to stop before the next function line,
// as Ctrl-C does. It is safe to call from another goroutine.
func (in *Interp) Interrupt() { in.interrupted.Store(true) }
