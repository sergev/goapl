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

func TestStructural(t *testing.T) {
	cases := []struct{ in, want string }{
		{"⍳3", "1 2 3"},
		{"⍳0", ""},
		{"⍳,4", "1 2 3 4"},
		{"2×⍳5", "2 4 6 8 10"},
		{"⍴1 2 3", "3"},
		{"⍴5", ""},
		{"⍴2 3⍴⍳6", "2 3"},
		{",2 2⍴⍳4", "1 2 3 4"},
		{",5", "5"},
		{"1 2 3⍳2", "2"},
		{"1 2 3⍳2 5", "2 4"},
		{"3⍴1 2", "1 2 1"},
		{"2 3⍴⍳6", "1 2 3\n4 5 6"},
		{"⍴0⍴5", "0"},
		{"3⍴0⍴5", "0 0 0"},
		{"1 2,3 4", "1 2 3 4"},
		{"5,⍳3", "5 1 2 3"},
		{"1,2", "1 2"},
		{"1 2 3+.×4 5 6", "32"},
		{"(2 3⍴⍳6)+.×3 2⍴⍳6", "22 28\n49 64"},
		{"1 2 3∧.=1 2 3", "1"},
		{"2+.×3", "6"},
		{"1 2∘.×1 2 3", "1 2 3\n2 4 6"},
		{"⍴(⍳2)∘.+2 3⍴⍳6", "2 2 3"},
		{"(⍳5)[3]", "3"},
		{"(⍳5)[2 2⍴1 2 3 4]", "1 2\n3 4"},
	}
	for _, c := range cases {
		if got := run(t, c.in); got != c.want {
			t.Errorf("%s = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIndex(t *testing.T) {
	got := run(t,
		"A←3 4⍴⍳12",
		"A[2;3]",
		"A[1 3;2 4]",
		"A[;1]",
		"A[2;3]←0",
		"A[2;]",
		"A[;4]←7 8 9",
		"A[;4]",
		"V←⍳5",
		"V[2 4]←0",
		"V",
		"V[V[1]]←9",
		"V",
	)
	want := "7\n 2  4\n10 12\n1 5 9\n5 6 0 8\n7 8 9\n1 0 3 0 5\n9 0 3 0 5"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestStructuralErrors(t *testing.T) {
	cases := []struct {
		in   string
		code int
	}{
		{"⍳1 2", 21},
		{"⍳¯1", 22},
		{"⍳2.5", 23},
		{"(2 2⍴1)⍳1", 29},
		{"(2 2⍴1)⍴1", 56},
		{"2 2 2 2⍴1", 60},
		{"(2 2⍴1),1", 53},
		{"1 2+.×1 2 3", 52},
		{"+.,1", 27},
		{"1∘.⍳1", 26},
		{"∘.+3", 13},
		{"(⍳5)[6]", 38},
		{"(⍳5)[0]", 38},
		{"(⍳5)[1.5]", 37},
		{"(⍳5)[1;1]", 35},
		{"(⍳5)[1", 34},
		{"V[1]←2", 100},
		{"(2 2⍴⍳4)∘.+2 2⍴1", 60},
	}
	for _, c := range cases {
		err := New(&bytes.Buffer{}).Exec(c.in)
		e, ok := err.(*Error)
		if !ok || e.Code != c.code {
			t.Errorf("%s: got %v, want code %d", c.in, err, c.code)
		}
	}
}

func TestIndexAssignErrors(t *testing.T) {
	cases := []struct {
		in   string
		code int
	}{
		{"V[1 2]←1 2 3", 36},
		{"V[6]←1", 38},
		{"V[1;1]←1", 35},
	}
	for _, c := range cases {
		in := New(&bytes.Buffer{})
		if err := in.Exec("V←⍳5"); err != nil {
			t.Fatal(err)
		}
		err := in.Exec(c.in)
		e, ok := err.(*Error)
		if !ok || e.Code != c.code {
			t.Errorf("%s: got %v, want code %d", c.in, err, c.code)
		}
	}
}

func TestFunctions(t *testing.T) {
	got := run(t,
		"∇R←FACT N",
		"R←×/⍳N",
		"∇",
		"FACT 5",
		"∇R←A PLUS B",
		"R←A+B",
		"∇",
		"2 PLUS 3",
		"1 PLUS FACT 3",
		"∇R←F N;I",
		"R←1",
		"I←0",
		"L:→(I≥N)/0",
		"I←I+1",
		"R←R×I",
		"→L",
		"∇",
		"N←99",
		"I←7",
		"F 6",
		"N,I",
		"∇HELLO",
		"⎕←42",
		"∇",
		"HELLO",
		"∇R←FIB N",
		"→(N>2)/REC",
		"R←1",
		"→0",
		"REC:R←(FIB N-1)+FIB N-2",
		"∇",
		"FIB 10",
		"∇R←DBL X",
		"R←2×X ∇",
		"DBL 21",
		"∇R←PI",
		"R←3.14",
		"∇",
		"2×PI",
	)
	want := "120\n5\n7\n720\n99 7\n42\n55\n42\n6.28"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestQuadInput(t *testing.T) {
	got := run(t, "X←⎕", "1 2 3", "Y←⎕", "1÷0", "2×X", "Y")
	want := "⎕:\n⎕:\nAttempted division by zero\n1÷0\n ↑\n⎕:\n2 4 6"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	err := New(&bytes.Buffer{}).Exec("⎕")
	if e, ok := err.(*Error); !ok || e.Code != 18 {
		t.Errorf("⎕ without input: got %v", err)
	}
}

func TestFunctionErrors(t *testing.T) {
	setup := []string{
		"∇R←MON X", "R←X", "∇",
		"∇R←A DY B", "R←A", "∇",
		"∇NORES", "1", "∇",
		"∇R←NOVAL", "1", "∇",
		"∇R←INF X", "R←INF X", "∇",
	}
	cases := []struct {
		in   string
		code int
	}{
		{"∇MON Y", 5},
		{"∇", 7},
		{"∇R←", 6},
		{"∇A B C D", 3},
		{"MON←1", 8},
		{"DY 1", 74},
		{"1 MON 2", 10},
		{"1+NORES", 100},
		{"NOVAL", 100},
		{"INF 1", 16},
		{"→1.5", 12},
	}
	for _, c := range cases {
		in := New(&bytes.Buffer{})
		for _, l := range setup {
			if err := in.Exec(l); err != nil {
				t.Fatalf("%s: %v", l, err)
			}
		}
		err := in.Exec(c.in)
		e, ok := err.(*Error)
		if !ok || e.Code != c.code {
			t.Errorf("%s: got %v, want code %d", c.in, err, c.code)
		}
	}
	in := New(&bytes.Buffer{})
	in.Exec("∇EMPTY")
	if err := in.Exec("∇"); err == nil || err.(*Error).Code != 75 {
		t.Errorf("empty function: got %v", err)
	}
}

func TestErrorLocation(t *testing.T) {
	got := run(t, "∇R←BAD X", "R←X÷0", "∇", "BAD 1")
	want := "Attempted division by zero\nBAD[1] R←X÷0\n          ↑"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestCompress(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1 0 1/⍳3", "1 3"},
		{"2/5", "5 5"},
		{"0 2 1/2 3⍴⍳6", "2 2 3\n5 5 6"},
		{"⍴0/5", "0"},
	}
	for _, c := range cases {
		if got := run(t, c.in); got != c.want {
			t.Errorf("%s = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestColorOutput(t *testing.T) {
	var out bytes.Buffer
	in := New(&out)
	in.Color = true
	in.Prompt = "      "
	in.Run(strings.NewReader("∇R←F X\nR←X÷0\n∇\nF 1\nX←⎕\n2\n"))
	want := "      \x1b[2m[1] \x1b[0m\x1b[2m[2] \x1b[0m      " +
		"\x1b[1;31mAttempted division by zero\x1b[0m\n" +
		"\x1b[33mF[1] \x1b[0mR←X÷0\n" +
		"        \x1b[31m↑\x1b[0m\n" +
		"      \x1b[2m⎕:\x1b[0m\n      "
	if out.String() != want {
		t.Errorf("got %q\nwant %q", out.String(), want)
	}
}
