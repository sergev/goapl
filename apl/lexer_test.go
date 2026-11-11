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
	}
	for _, c := range cases {
		if got := lexString(t, c.in); got != c.want {
			t.Errorf("Lex(%q) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestLexErrors(t *testing.T) {
	cases := []struct {
		in   string
		code int
		pos  int
	}{
		{"¯A", 1, 1},
		{"1.+2", 2, 2},
		{"1 # 2", 4, 2},
		{"ABCDEFGHIJK←1", 70, 0},
	}
	for _, c := range cases {
		_, err := Lex(c.in)
		le, ok := err.(*LexError)
		if !ok {
			t.Errorf("Lex(%q): got %v, want LexError", c.in, err)
			continue
		}
		if le.Msg != Messages[c.code] || le.Pos != c.pos {
			t.Errorf("Lex(%q) = %q at %d, want %q at %d", c.in, le.Msg, le.Pos, Messages[c.code], c.pos)
		}
	}
}
