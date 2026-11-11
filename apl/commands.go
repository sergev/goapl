package apl

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// command runs a system command such as )SI. Problems are reported as
// messages, not as APL errors.
func (in *Interp) command(line string) error {
	f := strings.Fields(line)
	args := f[1:]
	switch strings.ToUpper(f[0]) {
	case ")OFF":
		in.off = true
		return errQuit
	case ")SI":
		for i := len(in.stack) - 1; i >= 0; i-- {
			fr := in.stack[i]
			s := fmt.Sprintf("%s[%d]", fr.f.Name, fr.pc)
			if fr.suspended {
				s += " *"
			}
			in.say(s)
		}
	case ")RESET":
		return in.reset(nil)
	case ")FNS":
		in.listNames(args, func(n string) bool { return in.fns[n] != nil })
	case ")VARS":
		in.listNames(args, func(n string) bool { return in.vars[n] != nil })
	case ")ERASE":
		var missing []string
		for _, n := range args {
			if in.fns[n] == nil && in.vars[n] == nil {
				missing = append(missing, n)
			}
			in.erase(n)
		}
		if len(missing) > 0 {
			in.say("NOT FOUND: " + strings.Join(missing, " "))
		}
	case ")CLEAR":
		return in.reset(func() {
			in.replaceWith(New(nil))
			in.say("CLEAR WS")
		})
	case ")WSID":
		switch len(args) {
		case 0:
			in.say(in.wsName())
		case 1:
			in.say("WAS " + in.wsName())
			in.wsid = args[0]
		default:
			in.say("INCORRECT COMMAND")
		}
	case ")SAVE":
		in.save(args)
	case ")LOAD":
		if len(args) != 1 {
			in.say("INCORRECT COMMAND")
			return nil
		}
		ws, err := in.loadFile(args[0])
		if err != nil {
			in.say("WS NOT FOUND")
			return nil
		}
		return in.reset(func() {
			in.replaceWith(ws)
			in.wsid = args[0]
			in.say("LOADED " + args[0])
			in.RunLatent()
		})
	case ")COPY":
		if len(args) == 0 {
			in.say("INCORRECT COMMAND")
			return nil
		}
		ws, err := in.loadFile(args[0])
		if err != nil {
			in.say("WS NOT FOUND")
			return nil
		}
		in.copyFrom(ws, args[1:])
	default:
		in.say("INCORRECT COMMAND")
	}
	return nil
}

// say prints a line of command output.
func (in *Interp) say(s string) {
	in.endLine()
	fmt.Fprintln(in.out, s)
}

// reset clears the state indicator and then runs action, if any; with
// suspended functions, the action waits until they are unwound.
func (in *Interp) reset(action func()) error {
	if len(in.stack) == 0 {
		if action != nil {
			action()
		}
		return nil
	}
	in.atTop = action
	return errReset
}

// replaceWith makes the workspace that of ws, keeping the session's I/O.
func (in *Interp) replaceWith(ws *Interp) {
	in.vars, in.fns, in.env, in.lx = ws.vars, ws.fns, ws.env, ws.lx
	in.wsid, in.lastErr, in.editing = ws.wsid, nil, nil
	in.fnGen++
}

func (in *Interp) wsName() string {
	if in.wsid == "" {
		return "CLEAR WS"
	}
	return in.wsid
}

// listNames prints the sorted names that pass keep, from the one given
// in args on, folded at ⎕PW.
func (in *Interp) listNames(args []string, keep func(string) bool) {
	var names []string
	for n := range in.vars {
		names = append(names, n)
	}
	for n := range in.fns {
		names = append(names, n)
	}
	slices.Sort(names)
	names = slices.Compact(names)
	line := ""
	for _, n := range names {
		if !keep(n) || len(args) > 0 && n < args[0] {
			continue
		}
		if line != "" && len([]rune(line))+1+len([]rune(n)) > in.pw {
			in.say(line)
			line = ""
		}
		if line != "" {
			line += " "
		}
		line += n
	}
	if line != "" {
		in.say(line)
	}
}

// wsFile is the file of workspace name: name.apl unless it has an
// extension.
func wsFile(name string) string {
	if filepath.Ext(name) == "" {
		return name + ".apl"
	}
	return name
}

// save is )SAVE [name].
func (in *Interp) save(args []string) {
	name := in.wsid
	switch {
	case len(args) > 1:
		in.say("INCORRECT COMMAND")
		return
	case len(args) == 1:
		name = args[0]
	case name == "":
		in.say("NOT SAVED, THIS WS IS CLEAR WS")
		return
	}
	if len(in.stack) > 0 {
		in.say("NOT SAVED, WS HAS SUSPENDED FUNCTIONS")
		return
	}
	if err := os.WriteFile(wsFile(name), []byte(in.source()), 0o644); err != nil {
		in.say("NOT SAVED: " + err.Error())
		return
	}
	if in.wsid == "" {
		in.wsid = name
	}
	in.say("SAVED " + name)
}

// source is APL text that recreates the workspace.
func (in *Interp) source() string {
	var b strings.Builder
	b.WriteString("⍝ goapl workspace\n")
	for _, n := range []string{"⎕IO", "⎕CT", "⎕PP", "⎕PW", "⎕RL", "⎕LX"} {
		v, _ := in.getVar(n)
		fmt.Fprintf(&b, "%s←%s\n", n, literal(v))
	}
	var names []string
	for n := range in.vars {
		names = append(names, n)
	}
	slices.Sort(names)
	for _, n := range names {
		fmt.Fprintf(&b, "%s←%s\n", n, literal(in.vars[n]))
	}
	names = names[:0]
	for n := range in.fns {
		names = append(names, n)
	}
	slices.Sort(names)
	for _, n := range names {
		lines := in.fns[n].Lines()
		fmt.Fprintf(&b, "∇%s\n", lines[0])
		for _, l := range lines[1:] {
			b.WriteString(l + "\n")
		}
		b.WriteString("∇\n")
	}
	return b.String()
}

// loadFile runs a workspace file in a new workspace.
func (in *Interp) loadFile(name string) (*Interp, error) {
	f, err := os.Open(wsFile(name))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ws := New(nopWriter{})
	ws.Run(f)
	ws.input = nil
	return ws, nil
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

// copyFrom copies the named objects (all if none) of ws.
func (in *Interp) copyFrom(ws *Interp, names []string) {
	if len(names) == 0 {
		for n := range ws.vars {
			names = append(names, n)
		}
		for n := range ws.fns {
			names = append(names, n)
		}
	}
	var missing []string
	for _, n := range names {
		switch {
		case ws.fns[n] != nil:
			delete(in.vars, n)
			in.fns[n] = ws.fns[n]
			in.fnGen++
		case ws.vars[n] != nil:
			if in.fns[n] != nil {
				delete(in.fns, n)
				in.fnGen++
			}
			in.vars[n] = ws.vars[n]
		default:
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		in.say("NOT FOUND: " + strings.Join(missing, " "))
	}
}

// RunLatent executes ⎕LX, if set, as a session line.
func (in *Interp) RunLatent() {
	if in.lx.Size() == 0 {
		return
	}
	var b strings.Builder
	for _, v := range in.lx.Elems {
		b.WriteRune(rune(v.(Char)))
	}
	if err := in.Exec(b.String()); err != nil {
		in.report(err)
	}
}

// Off reports whether )OFF ended the session.
func (in *Interp) Off() bool { return in.off }

// literal is an APL expression whose value is v.
func literal(v *Array) string {
	switch {
	case v.IsScalar():
		return elemLiteral(v.At(0), false)
	case v.Rank() == 1:
		return vectorLiteral(v)
	}
	shape := make([]string, v.Rank())
	for i, n := range v.Shape {
		shape[i] = strconv.Itoa(n)
	}
	rv := &Array{Shape: []int{v.Size()}, Data: v.Data, Elems: v.Elems, Proto: v.Proto}
	return strings.Join(shape, " ") + "⍴" + paren(vectorLiteral(rv))
}

func paren(s string) string { return "(" + s + ")" }

// vectorLiteral is an expression for a vector.
func vectorLiteral(v *Array) string {
	n := v.Size()
	switch {
	case n == 0:
		switch p := v.Fill().(type) {
		case Char:
			return "''"
		case *Array:
			return "0⍴⊂" + paren(literal(p))
		}
		return "⍳0"
	case n == 1:
		return "," + elemLiteral(v.At(0), false)
	case v.IsCharacter():
		var b strings.Builder
		for _, c := range v.Elems {
			b.WriteRune(rune(c.(Char)))
		}
		return quote(b.String())
	}
	items := make([]string, n)
	for i := range n {
		items[i] = elemLiteral(v.At(i), true)
	}
	return strings.Join(items, " ")
}

// elemLiteral is an element as a scalar, or as an item of a strand.
func elemLiteral(e Value, inStrand bool) string {
	switch e := e.(type) {
	case float64:
		return numLiteral(e)
	case Char:
		return quote(string(rune(e)))
	case *Array:
		if inStrand {
			return paren(literal(e))
		}
		return "⊂" + paren(literal(e))
	}
	return "?"
}

// numLiteral writes a number exactly: ¯ for minus, E for exponent.
func numLiteral(v float64) string {
	s := strconv.FormatFloat(v, 'g', -1, 64)
	s = strings.Replace(s, "e+", "E", 1)
	s = strings.Replace(s, "e", "E", 1)
	return strings.ReplaceAll(s, "-", "¯")
}

func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
