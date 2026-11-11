package apl

import (
	"strings"
	"testing"
)

func lexString(t *testing.T, line string) string {
	t.Helper()
	toks, err := Lex(line)
	if err != nil {
		t.Fatalf("Lex(%q): %v", line, err)
	}
	s := make([]string, len(toks))
	for i, tok := range toks {
		s[i] = tok.String()
	}
	return strings.Join(s, " ")
}

func TestLex(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1+2", "NUM(1) PRIM(+) NUM(2)"},
		{"1 2 3", "NUM(1 2 3)"},
		{"¯1.5 .25 1E3", "NUM(-1.5 0.25 1000)"},
		{"+/⍳10", "PRIM(+) PRIM(/) PRIM(⍳) NUM(10)"},
		{"A←B[1;2]", "NAME(A) SPEC(←) NAME(B) SPEC([) NUM(1) SPEC(;) NUM(2) SPEC(])"},
		{"∇R←X F Y", "SPEC(∇) NAME(R) SPEC(←) NAME(X) NAME(F) NAME(Y)"},
		{"L1:→L1", "NAME(L1) SPEC(:) SPEC(→) NAME(L1)"},
		{"1 2∘.×3", "NUM(1 2) SPEC(∘) SPEC(.) PRIM(×) NUM(3)"},
		{"⎕←X", "SPEC(⎕) SPEC(←) NAME(X)"},
		{"⎕IO←⎕ X", "NAME(⎕IO) SPEC(←) SPEC(⎕) NAME(X)"},
		{"'AB''C',''", "STR(AB'C) PRIM(,) STR()"},
		{"⍞←'⍝'", "SPEC(⍞) SPEC(←) STR(⍝)"},
		{"LongName_∆1←x", "NAME(LongName_∆1) SPEC(←) NAME(x)"},
	}
	for _, c := range cases {
		if got := lexString(t, c.in); got != c.want {
			t.Errorf("Lex(%q) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestLexErrors(t *testing.T) {
	cases := []struct {
		in    string
		class Class
		pos   int
	}{
		{"¯A", SyntaxError, 1},
		{"1.+2", SyntaxError, 2},
		{"1 # 2", SyntaxError, 2},
		{"1,'AB", SyntaxError, 2},
		{"'AB''", SyntaxError, 0},
	}
	for _, c := range cases {
		_, err := Lex(c.in)
		le, ok := err.(*Error)
		if !ok {
			t.Errorf("Lex(%q): got %v, want *Error", c.in, err)
			continue
		}
		if le.Class != c.class || le.Pos != c.pos {
			t.Errorf("Lex(%q) = %v at %d, want %v at %d", c.in, le.Class, le.Pos, c.class, c.pos)
		}
	}
}
