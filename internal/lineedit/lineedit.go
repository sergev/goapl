// Package lineedit reads lines from a terminal with editing, history
// and an APL keyboard.
package lineedit

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// ErrInterrupt is returned when Ctrl-C abandons a line.
var ErrInterrupt = errors.New("interrupt")

// Keymap gives the APL glyph typed with the prefix key ` or with Alt.
var Keymap = map[rune]rune{
	'a': '⍺', 'b': '⊥', 'c': '∩', 'd': '⌊', 'e': '∊', 'f': '_', 'g': '∇',
	'h': '∆', 'i': '⍳', 'j': '∘', 'k': '\'', 'l': '⎕', 'm': '|', 'n': '⊤',
	'o': '○', 'p': '*', 'q': '?', 'r': '⍴', 's': '⌈', 't': '~', 'u': '↓',
	'v': '∪', 'w': '⍵', 'x': '⊃', 'y': '↑', 'z': '⊂',
	'1': '¨', '2': '¯', '3': '<', '4': '≤', '5': '=', '6': '≥', '7': '>',
	'8': '≠', '9': '∨', '0': '∧', '-': '×', '=': '÷', '[': '←', ']': '→',
	';': '⍎', '\'': '⍕', ',': '⍝', '.': '⍀', '/': '⌿', '`': '⋄',
	'^': '⍉', '&': '⊖', '%': '⌽', '$': '⍋', '#': '⍒', '*': '⍟', '<': '⍪',
	':': '≡', '+': '⌹', '(': '⍱', ')': '⍲', 'L': '⍞',
}

// Editor reads lines with editing and history.
type Editor struct {
	in      *os.File
	r       *bufio.Reader
	out     io.Writer
	History []string
	file    string // history file, or ""
}

// New returns an editor for terminal in, writing to out. If file is not
// empty, history is loaded from it and each line is appended to it.
func New(in *os.File, out io.Writer, file string) *Editor {
	e := &Editor{in: in, r: bufio.NewReader(in), out: out, file: file}
	if file != "" {
		if data, err := os.ReadFile(file); err == nil {
			for _, l := range strings.Split(string(data), "\n") {
				if l != "" {
					e.History = append(e.History, l)
				}
			}
		}
	}
	return e
}

// ReadLine shows prompt and reads a line; io.EOF at end of input.
func (e *Editor) ReadLine(prompt string) (string, error) {
	fd := int(e.in.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	line, err := e.edit(prompt)
	term.Restore(fd, state)
	if err == nil && line != "" {
		e.remember(line)
	}
	return line, err
}

func (e *Editor) remember(line string) {
	if n := len(e.History); n > 0 && e.History[n-1] == line {
		return
	}
	e.History = append(e.History, line)
	if e.file != "" {
		if f, err := os.OpenFile(e.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
			fmt.Fprintln(f, line)
			f.Close()
		}
	}
}

// edit runs the editing loop on a terminal in raw mode.
func (e *Editor) edit(prompt string) (string, error) {
	var buf []rune
	pos := 0
	hist := len(e.History)
	saved := ""
	redraw := func() {
		fmt.Fprint(e.out, "\r"+prompt+string(buf)+"\x1b[K")
		if n := len(buf) - pos; n > 0 {
			fmt.Fprintf(e.out, "\x1b[%dD", n)
		}
	}
	insert := func(c rune) {
		buf = append(buf[:pos], append([]rune{c}, buf[pos:]...)...)
		pos++
	}
	recall := func(i int) {
		if i < 0 || i > len(e.History) {
			return
		}
		if hist == len(e.History) {
			saved = string(buf)
		}
		hist = i
		if i == len(e.History) {
			buf = []rune(saved)
		} else {
			buf = []rune(e.History[i])
		}
		pos = len(buf)
	}
	redraw()
	for {
		c, _, err := e.r.ReadRune()
		if err != nil {
			return "", io.EOF
		}
		switch c {
		case '\r', '\n':
			fmt.Fprint(e.out, "\r\n")
			return string(buf), nil
		case 3: // Ctrl-C
			fmt.Fprint(e.out, "^C\r\n")
			return "", ErrInterrupt
		case 4: // Ctrl-D: delete, or end of input on an empty line
			if len(buf) == 0 {
				fmt.Fprint(e.out, "\r\n")
				return "", io.EOF
			}
			if pos < len(buf) {
				buf = append(buf[:pos], buf[pos+1:]...)
			}
		case 127, 8:
			if pos > 0 {
				buf = append(buf[:pos-1], buf[pos:]...)
				pos--
			}
		case 1:
			pos = 0
		case 5:
			pos = len(buf)
		case 2:
			pos = max(pos-1, 0)
		case 6:
			pos = min(pos+1, len(buf))
		case 11:
			buf = buf[:pos]
		case 21:
			buf, pos = buf[pos:], 0
		case 16:
			recall(hist - 1)
		case 14:
			recall(hist + 1)
		case 12:
			fmt.Fprint(e.out, "\x1b[H\x1b[2J")
		case '`':
			k, _, err := e.r.ReadRune()
			if err != nil {
				return "", io.EOF
			}
			if g, ok := Keymap[k]; ok {
				insert(g)
			} else {
				insert('`')
				if k >= ' ' {
					insert(k)
				}
			}
		case 27:
			e.escape(&buf, &pos, insert, recall, &hist)
		default:
			if c >= ' ' {
				insert(c)
			}
		}
		redraw()
	}
}

// escape handles an escape sequence: arrows, Home, End, Delete, or Alt
// with a key of the APL keyboard.
func (e *Editor) escape(buf *[]rune, pos *int, insert func(rune), recall func(int), hist *int) {
	c, _, err := e.r.ReadRune()
	if err != nil {
		return
	}
	if c != '[' && c != 'O' {
		if g, ok := Keymap[c]; ok {
			insert(g)
		}
		return
	}
	var params []rune // a CSI sequence ends with a byte in @..~
	k, _, err := e.r.ReadRune()
	for err == nil && (k < '@' || k > '~') {
		params = append(params, k)
		k, _, err = e.r.ReadRune()
	}
	if err != nil {
		return
	}
	if k == '~' {
		if string(params) == "3" && *pos < len(*buf) {
			*buf = append((*buf)[:*pos], (*buf)[*pos+1:]...)
		}
		return
	}
	switch k {
	case 'A':
		recall(*hist - 1)
	case 'B':
		recall(*hist + 1)
	case 'C':
		*pos = min(*pos+1, len(*buf))
	case 'D':
		*pos = max(*pos-1, 0)
	case 'H':
		*pos = 0
	case 'F':
		*pos = len(*buf)
	}
}
