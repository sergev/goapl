package apl

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
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
		{"+/5", "5"},
		{"+/,5", "5"},
		{"+/⍳0", "0"},
		{"×/⍳0", "1"},
		{"⌊/⍳0", "1.797693135E308"},
		{"⍴+/2 0⍴0", "2"},
		{"(2 0⍴0)+.×⍳0", "0 0"},
		{"0÷0", "1"},
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
		in    string
		class Class
	}{
		{"8÷0", DomainError},
		{"÷0", DomainError},
		{"~2", DomainError},
		{"2∧3", DomainError},
		{"1 2+1 2 3", LengthError},
		{"=5", ValenceError},
		{"(1+2", SyntaxError},
		{"1+2)", SyntaxError},
		{"X", ValueError},
		{"2+\\1 2", ValenceError},
		{"+/[2]1 2", AxisError},
		{"+[1]1 2", AxisError},
		{"⍳[1]3", AxisError},
		{"1 2+[1]3", AxisError},
		{"+/[1;2]1", AxisError},
		{"2+", SyntaxError},
		{"1 2 (3", SyntaxError},
		{"⍎/1 0", ValenceError},
		{"/1 2", SyntaxError},
		{"¯2⍟8", DomainError},
		{"⍟/⍳0", DomainError},
		{"!¯1", DomainError},
		{"8○1", DomainError},
		{"?0", DomainError},
		{"3?2", DomainError},
		{"1 2?3", LengthError},
		{"⍲1", ValenceError},
		{"2⍱1", DomainError},
		{"↓2", ValenceError},
		{"⍪2", ValenceError},
		{"/5", SyntaxError},
		{"1 2+2 2⍴1", RankError},
	}
	for _, c := range cases {
		err := New(&bytes.Buffer{}).Exec(c.in)
		e, ok := err.(*Error)
		if !ok || e.Class != c.class {
			t.Errorf("%s: got %v, want %v", c.in, err, c.class)
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
		in    string
		class Class
	}{
		{"⍳1 2", LengthError},
		{"⍳¯1", DomainError},
		{"⍳2.5", DomainError},
		{"(2 2⍴1)⍳1", RankError},
		{"(2 2⍴1)⍴1", RankError},
		{"(2 2 2⍴1),1 2", RankError},
		{"1 2+.×1 2 3", LengthError},
		{"+.,1", SyntaxError},
		{"1∘.⍳1", RankError},
		{"1∘.¨1", SyntaxError},
		{"∘.+3", SyntaxError},
		{"(⍳5)[6]", IndexError},
		{"(⍳5)[0]", IndexError},
		{"(⍳5)[1.5]", DomainError},
		{"(⍳5)[1;1]", RankError},
		{"(⍳5)[1", SyntaxError},
		{"V[1]←2", ValueError},
	}
	for _, c := range cases {
		err := New(&bytes.Buffer{}).Exec(c.in)
		e, ok := err.(*Error)
		if !ok || e.Class != c.class {
			t.Errorf("%s: got %v, want %v", c.in, err, c.class)
		}
	}
}

func TestIndexAssignErrors(t *testing.T) {
	cases := []struct {
		in    string
		class Class
	}{
		{"V[1 2]←1 2 3", LengthError},
		{"V[6]←1", IndexError},
		{"V[1;1]←1", RankError},
	}
	for _, c := range cases {
		in := New(&bytes.Buffer{})
		if err := in.Exec("V←⍳5"); err != nil {
			t.Fatal(err)
		}
		err := in.Exec(c.in)
		e, ok := err.(*Error)
		if !ok || e.Class != c.class {
			t.Errorf("%s: got %v, want %v", c.in, err, c.class)
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
	want := "⎕:\n⎕:\nDOMAIN ERROR\n1÷0\n ^\n⎕:\n2 4 6"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	err := New(&bytes.Buffer{}).Exec("⎕")
	if e, ok := err.(*Error); !ok || e.Class != Interrupt {
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
		in    string
		class Class
	}{
		{"∇MON Y", DefnError},
		{"∇", DefnError},
		{"∇R←", DefnError},
		{"∇A B C D", DefnError},
		{"MON←1", SyntaxError},
		{"DY 1", ValueError},
		{"1 MON 2", ValenceError},
		{"1+NORES", ValueError},
		{"NOVAL", ValueError},
		{"INF 1", SystemLimit},
		{"→1.5", DomainError},
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
		if !ok || e.Class != c.class {
			t.Errorf("%s: got %v, want %v", c.in, err, c.class)
		}
	}
	in := New(&bytes.Buffer{})
	in.Exec("∇EMPTY")
	if err := in.Exec("∇"); err != nil {
		t.Errorf("empty function: got %v", err)
	}
	if err := in.Exec("EMPTY"); err != nil {
		t.Errorf("empty function call: got %v", err)
	}
}

func TestErrorLocation(t *testing.T) {
	got := run(t, "∇R←BAD X", "R←X÷0", "∇", "BAD 1")
	want := "DOMAIN ERROR\nBAD[1] R←X÷0\n          ^"
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
		"\x1b[1;31mDOMAIN ERROR\x1b[0m\n" +
		"\x1b[33mF[1] \x1b[0mR←X÷0\n" +
		"        \x1b[31m^\x1b[0m\n" +
		"      \x1b[2m⎕:\x1b[0m\n            "
	if out.String() != want {
		t.Errorf("got %q\nwant %q", out.String(), want)
	}
}

// Interrupt stops a running loop and suspends the function.
func TestInterrupt(t *testing.T) {
	var out bytes.Buffer
	in := New(&out)
	r, w := io.Pipe()
	done := make(chan bool)
	go func() {
		in.Run(r)
		done <- true
	}()
	io.WriteString(w, "∇LOOP\nL:→L\n∇\nLOOP\n")
	time.Sleep(50 * time.Millisecond)
	in.Interrupt()
	io.WriteString(w, ")SI\n")
	w.Close()
	<-done
	want := "INTERRUPT\nLOOP[1] L:→L\nLOOP[1] *\n"
	if out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}
