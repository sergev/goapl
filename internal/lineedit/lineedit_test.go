package lineedit

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// run edits the keystrokes in input and gives the lines read.
func run(t *testing.T, e *Editor, input string) []string {
	t.Helper()
	e.r = bufio.NewReader(strings.NewReader(input))
	var lines []string
	for {
		l, err := e.edit("> ")
		if err == io.EOF {
			return lines
		}
		if err == ErrInterrupt {
			l = "^C"
		} else {
			e.remember(l)
		}
		lines = append(lines, l)
	}
}

func TestEdit(t *testing.T) {
	cases := []struct{ in, want string }{
		{"abc\r", "abc"},
		{"abd\x7fc\r", "abc"},                  // backspace
		{"bc\x01a\r", "abc"},                   // Ctrl-A
		{"ac\x02b\r", "abc"},                   // Ctrl-B
		{"ab\x01\x06\x05c\r", "abc"},           // Ctrl-F, Ctrl-E
		{"abXY\x02\x02\x0bc\r", "abc"},         // Ctrl-K
		{"XYabc\x02\x02\x02\x15\x05\r", "abc"}, // Ctrl-U
		{"ac\x1b[Db\x1b[C\r", "abc"},           // arrows
		{"bc\x1b[Ha\x1b[F!\r", "abc!"},         // Home, End
		{"aXbc\x1b[H\x1b[C\x1b[3~\r", "abc"},   // Delete
		{"aXc\x02\x02\x04\x04b\r", "ab"},       // Ctrl-D deletes
		{"a\x1b[1;5Cb\r", "ab"},                // unknown sequence
		{"`r`i5\r", "⍴⍳5"},                     // prefix key
		{"\x1br\x1bi5\r", "⍴⍳5"},               // Alt
		{"``x` y\r", "⋄x` y"},                  // `` is ⋄; ` then other keys
		{"`\x01\r", "`"},                       // ` then a control key
		{"\x03", "^C"},
	}
	for _, c := range cases {
		e := &Editor{out: &bytes.Buffer{}}
		got := run(t, e, c.in)
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("%q: got %q, want %q", c.in, got, c.want)
		}
	}
}

func TestHistory(t *testing.T) {
	e := &Editor{out: &bytes.Buffer{}}
	got := run(t, e, "one\rtwo\rtwo\r\x1b[A\x1b[A!\r\x10\x10\x0e\r\x1b[A\x1b[B\x1b[Bnew\r")
	want := []string{"one", "two", "two", "one!", "one!", "new"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("got %q, want %q", got, want)
	}
	if strings.Join(e.History, ",") != "one,two,one!,new" {
		t.Errorf("history %q", e.History)
	}
}

func TestRedraw(t *testing.T) {
	var out bytes.Buffer
	e := &Editor{out: &out}
	run(t, e, "ab\x02\r")
	want := "\r> \x1b[K\r> a\x1b[K\r> ab\x1b[K\r> ab\x1b[K\x1b[1D\r\n\r> \x1b[K"
	if out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

func TestEOF(t *testing.T) {
	e := &Editor{out: &bytes.Buffer{}}
	if got := run(t, e, "abc"); len(got) != 0 {
		t.Errorf("got %q", got)
	}
	if got := run(t, e, "\x04"); len(got) != 0 {
		t.Errorf("Ctrl-D: got %q", got)
	}
	if got := run(t, e, "`"); len(got) != 0 {
		t.Errorf("prefix at EOF: got %q", got)
	}
}

// History is loaded from its file, and new lines are appended to it.
func TestHistoryFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "history")
	os.WriteFile(file, []byte("old\n\n"), 0o600)
	e := New(os.Stdin, &bytes.Buffer{}, file)
	if strings.Join(e.History, ",") != "old" {
		t.Fatalf("history %q", e.History)
	}
	run(t, e, "\x1b[A!\r")
	data, _ := os.ReadFile(file)
	if string(data) != "old\n\nold!\n" {
		t.Errorf("file %q", data)
	}
}
