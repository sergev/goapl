package apl

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

// edLine is a line in the ∇ editor; the header is line 0.
type edLine struct {
	num  float64
	text string
}

// editor holds a function being defined or edited.
type editor struct {
	lines []edLine // sorted by number
	cur   float64  // number of the next line
}

// lineNum formats a line number: 3 or 1.5.
func lineNum(n float64) string { return strconv.FormatFloat(n, 'f', -1, 64) }

// openEditor handles "∇header", "∇F", "∇F[cmd] text" and a trailing ∇.
func (in *Interp) openEditor(line string) error {
	rs := []rune(line)
	start := strings.IndexRune(line, '∇')
	s := string([]rune(line[start:])[1:])
	closing := false
	if t := strings.TrimRight(s, " "); strings.HasSuffix(t, "∇") && len(rs) > 1 {
		closing = true
		s = strings.TrimSuffix(t, "∇")
	}
	header, cmd := s, ""
	if i := strings.Index(s, "["); i >= 0 {
		header, cmd = s[:i], s[i:]
	}
	pos := len([]rune(line[:start])) + 1 + len([]rune(header)) - len([]rune(strings.TrimLeft(header, " ")))
	header = strings.TrimSpace(header)
	toks, err := Lex(header)
	if err != nil || len(toks) == 0 {
		return errorAt(DefnError, pos)
	}
	ed := &editor{}
	if old := in.fns[header]; old != nil && len(toks) == 1 {
		for i, l := range old.Lines() {
			ed.lines = append(ed.lines, edLine{float64(i), l})
		}
		ed.cur = float64(len(old.Stmts) + 1)
	} else {
		f, err := parseHeader(header)
		if err != nil {
			if e, ok := err.(*Error); ok && e.Pos != NoPos {
				e.Pos += pos
			}
			return err
		}
		if in.fns[f.Name] != nil || in.vars[f.Name] != nil || cmd != "" {
			return errorAt(DefnError, pos)
		}
		ed.lines = []edLine{{0, header}}
		ed.cur = 1
	}
	in.editing = ed
	if cmd != "" {
		if err := in.editInput(cmd); err != nil {
			in.editing = nil
			return err
		}
	}
	if closing {
		return in.closeEditor()
	}
	return nil
}

// editLine processes a line typed in the editor.
func (in *Interp) editLine(line string) error {
	t := strings.TrimRight(line, " ")
	closing := strings.HasSuffix(t, "∇")
	if closing {
		t = strings.TrimSuffix(t, "∇")
	}
	if err := in.editInput(t); err != nil {
		return err
	}
	if closing {
		return in.closeEditor()
	}
	return nil
}

// editInput handles "[cmd] text" or "text": cmd is n (go to line n),
// ⎕ (show all), ⎕n (show from n), n⎕ (show line n) or ∆n (delete n).
func (in *Interp) editInput(text string) error {
	ed := in.editing
	t := strings.TrimLeft(text, " ")
	if strings.HasPrefix(t, "[") {
		end := strings.Index(t, "]")
		if end < 0 {
			return errorAt(DefnError, 0)
		}
		cmd := strings.TrimSpace(t[1:end])
		t = strings.TrimLeft(t[end+1:], " ")
		num := func(s string) (float64, error) {
			n, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil || n < 0 {
				return 0, errorAt(DefnError, 1)
			}
			return n, nil
		}
		switch {
		case cmd == "⎕":
			in.showLines(0, math.Inf(1))
		case strings.HasPrefix(cmd, "⎕"):
			n, err := num(strings.TrimPrefix(cmd, "⎕"))
			if err != nil {
				return err
			}
			in.showLines(n, math.Inf(1))
		case strings.HasSuffix(cmd, "⎕"):
			n, err := num(strings.TrimSuffix(cmd, "⎕"))
			if err != nil {
				return err
			}
			in.showLines(n, n)
		case strings.HasPrefix(cmd, "∆"):
			n, err := num(strings.TrimPrefix(cmd, "∆"))
			if err != nil || n == 0 {
				return errorAt(DefnError, 1)
			}
			ed.lines = slices.DeleteFunc(ed.lines, func(l edLine) bool { return l.num == n })
			ed.cur = n
		default:
			n, err := num(cmd)
			if err != nil {
				return err
			}
			ed.cur = n
		}
	}
	if strings.TrimSpace(t) == "" {
		return nil
	}
	if ed.cur == 0 {
		if _, err := parseHeader(t); err != nil {
			return err
		}
	} else if _, err := Lex(t); err != nil {
		return err
	}
	i, found := slices.BinarySearchFunc(ed.lines, ed.cur, func(l edLine, n float64) int {
		switch {
		case l.num < n:
			return -1
		case l.num > n:
			return 1
		}
		return 0
	})
	if found {
		ed.lines[i].text = t
	} else {
		ed.lines = slices.Insert(ed.lines, i, edLine{ed.cur, t})
	}
	ed.cur = nextNum(ed.cur)
	return nil
}

// nextNum is the line number after n: n+1, or 1.6 after 1.5.
func nextNum(n float64) float64 {
	s := lineNum(n)
	i := strings.IndexByte(s, '.')
	if i < 0 {
		return n + 1
	}
	m := math.Pow(10, float64(len(s)-i-1))
	return (math.Round(n*m) + 1) / m
}

// showLines displays the lines numbered from..to; the whole function is
// framed by ∇ lines.
func (in *Interp) showLines(from, to float64) {
	in.endLine()
	ed := in.editing
	for _, l := range ed.lines {
		if l.num < from || l.num > to {
			continue
		}
		if l.num == 0 {
			fmt.Fprintln(in.out, "    ∇"+l.text)
		} else {
			fmt.Fprintf(in.out, "[%s]  %s\n", lineNum(l.num), l.text)
		}
	}
	if math.IsInf(to, 1) {
		fmt.Fprintln(in.out, "    ∇")
	}
}

// closeEditor fixes the edited function and leaves the editor.
func (in *Interp) closeEditor() error {
	ed := in.editing
	in.editing = nil
	var lines []string
	for _, l := range ed.lines {
		lines = append(lines, l.text)
	}
	f, line, err := fix(lines)
	if e, ok := err.(*Error); ok && line > 0 {
		e.Src, e.Where = lines[line], "["+itoa(line)+"]"
	}
	if err != nil {
		return err
	}
	if _, ok := in.vars[f.Name]; ok {
		return newError(DefnError)
	}
	in.fns[f.Name] = f
	in.fnGen++
	return nil
}
