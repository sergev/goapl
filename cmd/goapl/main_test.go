package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStdin(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run(nil, strings.NewReader("1+2\n"), &out, &errOut, false, false)
	if code != 0 || out.String() != "3\n" {
		t.Errorf("got %d %q", code, out.String())
	}
}

func TestPrompt(t *testing.T) {
	var out bytes.Buffer
	run(nil, strings.NewReader("∇R←F\nR←1\n∇\n"), &out, &out, true, false)
	if want := "      [1] [2]       "; out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

// Files share one workspace; /* ends only the current file.
func TestFiles(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.apl")
	b := filepath.Join(dir, "b.apl")
	os.WriteFile(a, []byte("X←5\n/*\nX←0\n"), 0o644)
	os.WriteFile(b, []byte("X+1\n"), 0o644)
	var out, errOut bytes.Buffer
	code := run([]string{a, b}, nil, &out, &errOut, false, false)
	if code != 0 || out.String() != "6\n" {
		t.Errorf("got %d %q %q", code, out.String(), errOut.String())
	}
}

func TestMissingFile(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"/nonexistent/x.apl"}, nil, &out, &errOut, false, false)
	if code != 1 || errOut.Len() == 0 {
		t.Errorf("got %d %q", code, errOut.String())
	}
}

func TestColor(t *testing.T) {
	var out bytes.Buffer
	run(nil, strings.NewReader("1+2\n÷0\n"), &out, &out, false, true)
	want := "\x1b[32m3\x1b[0m\n" +
		"\x1b[1;31mDOMAIN ERROR\x1b[0m\n÷0\n\x1b[31m^\x1b[0m\n"
	if out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

// -i continues with a session after the files; ⎕LX runs before it.
func TestSession(t *testing.T) {
	dir := t.TempDir()
	ws := filepath.Join(dir, "ws.apl")
	os.WriteFile(ws, []byte("X←5\n⎕LX←'X+1'\n"), 0o644)
	var out, errOut bytes.Buffer
	code := run([]string{"-i", ws}, strings.NewReader("X×2\n"), &out, &errOut, false, false)
	if code != 0 || out.String() != "6\n10\n" {
		t.Errorf("got %d %q %q", code, out.String(), errOut.String())
	}
}

// )OFF ends the program, also in a file.
func TestOff(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.apl")
	b := filepath.Join(dir, "b.apl")
	os.WriteFile(a, []byte("1\n)OFF\n2\n"), 0o644)
	os.WriteFile(b, []byte("3\n"), 0o644)
	var out, errOut bytes.Buffer
	code := run([]string{a, b}, nil, &out, &errOut, false, false)
	if code != 0 || out.String() != "1\n" {
		t.Errorf("got %d %q", code, out.String())
	}
}

func TestUsage(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"-x"}, nil, &out, &errOut, false, false); code != 2 {
		t.Errorf("got %d %q", code, errOut.String())
	}
}
