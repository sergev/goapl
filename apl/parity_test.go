package apl

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// Cases ported from the Pascal prototype's tests/run_symbol_tests.py,
// which exercises every glyph of APLCharSet. The prototype printed
// numbers in 12.2 format; expected values here use goapl's format.

type parityCase struct {
	name  string
	lines []string
	want  string // expected output; ignored if class != 0
	class Class  // expected error class, or 0
}

func parityCases() []parityCase {
	var cases []parityCase
	for c := 'A'; c <= 'Z'; c++ {
		cases = append(cases, parityCase{
			name: "letter_" + string(c), lines: []string{string(c) + "←3", string(c)}, want: "3",
		})
	}
	for _, d := range "1234567890" {
		cases = append(cases, parityCase{name: "digit_" + string(d), lines: []string{string(d)}, want: string(d)})
	}
	cases = append(cases, []parityCase{
		// monadic
		{name: "monadic_plus", lines: []string{"+5"}, want: "5"},
		{name: "monadic_minus", lines: []string{"-5"}, want: "¯5"},
		{name: "monadic_times_signum", lines: []string{"×5"}, want: "1"},
		{name: "monadic_times_neg", lines: []string{"×¯3"}, want: "¯1"},
		{name: "monadic_divide", lines: []string{"÷5"}, want: "0.2"},
		{name: "monadic_star_exp", lines: []string{"*1"}, want: "2.718281828"},
		{name: "monadic_tilde_not1", lines: []string{"~1"}, want: "0"},
		{name: "monadic_tilde_not0", lines: []string{"~0"}, want: "1"},
		{name: "monadic_iota", lines: []string{"⍳3"}, want: "1 2 3"},
		{name: "monadic_rho", lines: []string{"⍴1 2 3"}, want: "3"},
		{name: "monadic_comma_ravel", lines: []string{",1 2"}, want: "1 2"},
		// dyadic
		{name: "dyadic_plus", lines: []string{"1+2"}, want: "3"},
		{name: "dyadic_minus", lines: []string{"5-2"}, want: "3"},
		{name: "dyadic_times", lines: []string{"3×4"}, want: "12"},
		{name: "dyadic_divide", lines: []string{"8÷4"}, want: "2"},
		{name: "dyadic_star", lines: []string{"2*3"}, want: "8"},
		{name: "dyadic_iota_indexof", lines: []string{"1 2 3⍳2"}, want: "2"},
		{name: "dyadic_rho_reshape", lines: []string{"3⍴1 2"}, want: "1 2 1"},
		{name: "dyadic_comma_cat", lines: []string{"1 2,3 4"}, want: "1 2 3 4"},
		{name: "dyadic_equals", lines: []string{"1=1"}, want: "1"},
		{name: "dyadic_notequal", lines: []string{"1≠2"}, want: "1"},
		{name: "dyadic_less", lines: []string{"1<2"}, want: "1"},
		{name: "dyadic_leq", lines: []string{"1≤1"}, want: "1"},
		{name: "dyadic_geq", lines: []string{"2≥1"}, want: "1"},
		{name: "dyadic_greater", lines: []string{"2>1"}, want: "1"},
		{name: "dyadic_and", lines: []string{"1∧1"}, want: "1"},
		{name: "dyadic_or", lines: []string{"1∨0"}, want: "1"},
		// reduction
		{name: "reduce_plus", lines: []string{"+/1 2 3"}, want: "6"},
		{name: "reduce_minus", lines: []string{"-/10 3 2"}, want: "9"},
		{name: "reduce_times", lines: []string{"×/2 3 4"}, want: "24"},
		{name: "reduce_divide", lines: []string{"÷/8 2"}, want: "4"},
		{name: "reduce_star", lines: []string{"*/2 3"}, want: "8"},
		{name: "reduce_equals", lines: []string{"=/1 1 1"}, want: "1"},
		{name: "reduce_neq", lines: []string{"≠/1 2 1"}, want: "0"},
		{name: "reduce_less", lines: []string{"</1 2 3"}, want: "0"},
		{name: "reduce_leq", lines: []string{"≤/1 2 3"}, want: "1"},
		{name: "reduce_geq", lines: []string{"≥/3 2 1"}, want: "1"},
		{name: "reduce_greater", lines: []string{">/3 2 1"}, want: "1"},
		{name: "reduce_and", lines: []string{"∧/1 1 0"}, want: "0"},
		{name: "reduce_or", lines: []string{"∨/0 0 1"}, want: "1"},
		{name: "reduce_ceiling", lines: []string{"⌈/1 3 2"}, want: "3"},
		{name: "reduce_floor", lines: []string{"⌊/1 3 2"}, want: "1"},
		{name: "slash_as_reduce", lines: []string{"+/1 2"}, want: "3"},
		// products
		{name: "inner_product", lines: []string{"1+.×2"}, want: "2"},
		{name: "outer_product", lines: []string{"2∘.×3"}, want: "6"},
		// syntax
		{name: "paren_group", lines: []string{"(1+2)×3"}, want: "9"},
		{name: "paren_close", lines: []string{"(1+2)"}, want: "3"},
		{name: "assign_leftarrow", lines: []string{"A←9", "A"}, want: "9"},
		{name: "space_separator", lines: []string{"1 2 3"}, want: "1 2 3"},
		{name: "colon_label", lines: []string{"L:1+1"}, want: "2"},
		{name: "high_minus_literal", lines: []string{"¯7"}, want: "¯7"},
		{name: "quit_slash_star", lines: []string{"1+0", "/*", "2"}, want: "1"},
		{name: "rightarrow_present", lines: []string{"→0"}, want: ""},
		{name: "quad_alone", lines: []string{"⎕", "5"}, want: "⎕:\n5"},
		// glyphs in APLCharSet that are not functions
		{name: "charset_only_DoubleQuote", lines: []string{`1"2`}, class: SyntaxError},
		{name: "charset_only_omega", lines: []string{"1ω2"}, class: SyntaxError},
		{name: "charset_only_epsilon", lines: []string{"1ε2"}, class: NonceError},
		{name: "charset_only_UpArrow", lines: []string{"1↑2"}, class: NonceError},
		{name: "charset_only_DownArrow", lines: []string{"1↓2"}, class: NonceError},
		{name: "charset_only_alpha", lines: []string{"1α2"}, class: SyntaxError},
		{name: "charset_only_UnderScore", lines: []string{"1_2"}, class: SyntaxError},
		{name: "charset_only_del", lines: []string{"1∇2"}, class: SyntaxError},
		{name: "charset_only_delta", lines: []string{"1∆2"}, class: SyntaxError},
		{name: "charset_only_SingleQuote", lines: []string{"1'2"}, class: SyntaxError},
		{name: "charset_only_EastCap", lines: []string{"1⊃2"}, class: SyntaxError},
		{name: "charset_only_WestCap", lines: []string{"1⊂2"}, class: SyntaxError},
		{name: "charset_only_NorthCap", lines: []string{"1∪2"}, class: SyntaxError},
		{name: "charset_only_ibeam", lines: []string{"1⊥2"}, class: NonceError},
		{name: "charset_only_TBeam", lines: []string{"1⊤2"}, class: NonceError},
		{name: "charset_only_LeftBracket", lines: []string{"1[2]"}, class: RankError},

		// Deliberate differences from the prototype.
		// Errors are reported by APL2 class, not by the prototype's
		// numbered messages. Unimplemented glyphs give NONCE ERROR.
		// ○/ was log to base; ○ is the circular functions.
		{name: "differs_reduce_circle", lines: []string{"○/2 8"}, want: "¯0.1455000338"},
		// 1\2 was an error; \ is expand.
		{name: "differs_BackwardSlash", lines: []string{`1\2`}, want: "2"},
		// 1?2 was an error; ? is deal.
		{name: "differs_QuestionMark", lines: []string{"1?2"}, want: "1"},
		// Reducing a scalar, one item or an empty axis were errors.
		{name: "differs_reduce_scalar", lines: []string{"+/5"}, want: "5"},
		{name: "differs_reduce_one", lines: []string{"+/,5"}, want: "5"},
		{name: "differs_reduce_empty", lines: []string{"×/⍳0"}, want: "1"},
		// 0÷0 was an error; it is 1 as in APL2.
		{name: "differs_zero_div_zero", lines: []string{"0÷0"}, want: "1"},
		// Names were limited to 10 characters.
		{name: "differs_long_name", lines: []string{"ABCDEFGHIJK←1", "ABCDEFGHIJK"}, want: "1"},
		// 1|2 was an error; | is residue here.
		{name: "differs_VerticalStroke", lines: []string{"1|2"}, want: "0"},
		// 1;2 printed 2; ; is only an index separator here.
		{name: "differs_semicolon_stmt", lines: []string{"1;2"}, class: SyntaxError},
		// 1]2 printed 2; a stray ] is an error here.
		{name: "differs_RightBracket", lines: []string{"1]2"}, class: SyntaxError},
		// 1∩2 printed 1; ∩ is not a known glyph here.
		{name: "differs_SouthCap", lines: []string{"1∩2"}, class: SyntaxError},
	}...)
	return cases
}

func TestParity(t *testing.T) {
	for _, c := range parityCases() {
		t.Run(c.name, func(t *testing.T) {
			if c.class != 0 {
				in := New(&bytes.Buffer{})
				var err error
				for _, l := range c.lines {
					if err = in.Exec(l); err != nil {
						break
					}
				}
				if e, ok := err.(*Error); !ok || e.Class != c.class {
					t.Errorf("%v: got %v, want %v", c.lines, err, c.class)
				}
				return
			}
			if got := run(t, c.lines...); got != c.want {
				t.Errorf("%v = %q, want %q", c.lines, got, c.want)
			}
		})
	}
}

// TestParityCoverage checks that every APLCharSet glyph is exercised.
func TestParityCoverage(t *testing.T) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890:→←∘.()[];⎕ +-×÷*⍳⍴,~=≠<≤≥>∧∨⌈⌊○/\"¯?ωε↑↓α_∇∆'⊃⊂∩∪⊥⊤|\\"
	var all strings.Builder
	for _, c := range parityCases() {
		all.WriteString(strings.Join(c.lines, "\n"))
	}
	text := all.String()
	var missing []string
	for _, g := range charset {
		if !strings.ContainsRune(text, g) {
			missing = append(missing, fmt.Sprintf("%q", g))
		}
	}
	if len(missing) > 0 {
		t.Errorf("glyphs without cases: %s", strings.Join(missing, " "))
	}
}
