package apl

import (
	"bytes"
	"strings"
	"testing"
)

// run executes lines and returns trimmed output.
func run(t *testing.T, lines ...string) string {
	t.Helper()
	var out bytes.Buffer
	New(&out).Run(strings.NewReader(strings.Join(lines, "\n")))
	return strings.TrimRight(out.String(), "\n")
}

func TestEval(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1+2", "3"},
		{"5-2", "3"},
		{"3×4", "12"},
		{"8÷4", "2"},
		{"2*3", "8"},
		{"2×3+4", "14"},
		{"(2×3)+4", "10"},
		{"1 2 3+4 5 6", "5 7 9"},
		{"10×1 2 3", "10 20 30"},
		{"1 2 3-1", "0 1 2"},
		{"-5", "¯5"},
		{"+5", "5"},
		{"×¯3 0 5", "¯1 0 1"},
		{"÷4", "0.25"},
		{"*0", "1"},
		{"~1 0", "0 1"},
		{"⌈2.5 ¯2.5", "3 ¯2"},
		{"⌊2.5 ¯2.5", "2 ¯3"},
		{"|¯4", "4"},
		{"3|7 ¯7", "1 2"},
		{"2⍟8", "3"},
		{"1=1", "1"},
		{"1≠2", "1"},
		{"1<2", "1"},
		{"1≤1", "1"},
		{"2≥1", "1"},
		{"2>1", "1"},
		{"1 1 0 0∧1 0 1 0", "1 0 0 0"},
		{"1 1 0 0∨1 0 1 0", "1 1 1 0"},
		{"0.1+0.2", "0.3"},
		{"(0.1+0.2)=0.3", "1"},
		{"+/1 2 3", "6"},
		{"-/10 3 2", "9"},
		{"×/2 3 4", "24"},
		{"÷/8 2", "4"},
		{"*/2 3", "8"},
		{"1++/1 2", "4"},
		{"=/1 1", "1"},
		{"⌈/3 9 2", "9"},
		{"⌊/3 9 2", "2"},
		{"○/2 8", "3"},
		{"∧/1 1 0", "0"},
		{"∨/1 0 0", "1"},
	}
	for _, c := range cases {
		if got := run(t, c.in); got != c.want {
			t.Errorf("%s = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestAssign(t *testing.T) {
	got := run(t, "A←2", "B←A×3", "A+B", "A←A+1", "A")
	if got != "8\n3" {
		t.Errorf("got %q", got)
	}
}

func TestErrors(t *testing.T) {
	cases := []struct {
		in   string
		code int
	}{
		{"8÷0", 20},
		{"÷0", 54},
		{"~2", 19},
		{"2∧3", 19},
		{"1 2+1 2 3", 55},
		{"=5", 74},
		{"(1+2", 15},
		{"1+2)", 15},
		{"X", 100},
		{"+/5", 24},
		{"+/1", 24},
		{"2+/1 2", 73},
		{"2+", 10},
		{"1 2 (3)", 10},
		{"~/1 0", 72},
		{"/1 2", 72},
		{"○/¯2 8", 50},
	}
	for _, c := range cases {
		err := New(&bytes.Buffer{}).Exec(c.in)
		e, ok := err.(*Error)
		if !ok || e.Code != c.code {
			t.Errorf("%s: got %v, want code %d", c.in, err, c.code)
		}
	}
}

func TestFormat(t *testing.T) {
	cases := []struct {
		v    float64
		want string
	}{
		{0, "0"},
		{-0.5, "¯0.5"},
		{1e20, "1E20"},
		{-1e-7, "¯1E¯7"},
		{1.0 / 3, "0.3333333333"},
		{123456789012, "1.23456789E11"},
	}
	for _, c := range cases {
		if got := FormatNumber(c.v); got != c.want {
			t.Errorf("FormatNumber(%v) = %q, want %q", c.v, got, c.want)
		}
	}
	m := &Array{Shape: []int{2, 2}, Data: []float64{1, -10, 100, 2}}
	if got := Format(m); got != "  1 ¯10\n100   2" {
		t.Errorf("matrix:\n%s", got)
	}
	c := &Array{Shape: []int{2, 1, 2}, Data: []float64{1, 2, 3, 4}}
	if got := Format(c); got != "1 2\n\n3 4" {
		t.Errorf("rank 3:\n%s", got)
	}
}
