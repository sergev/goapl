# Todo: toward APL2

goapl now implements the language of the Pascal prototype (steps 1–5 of
the README roadmap). This list grows it toward APL2, one task at a time.
Each task ends with: unit tests, a `testdata/` script (or additions to
one), README updates, and `make test` and `make compare` passing.

Already done: `⍝` comments; `⌈ ⌊ | ⍟` monadic and dyadic; compress `A/B`;
APL2 error classes and behavior where the original differs (T1);
arrays of any rank with general (character, mixed, nested) elements (T2);
`⎕IO ⎕CT ⎕PP ⎕RL` (T3); characters, strings and `⍞` (T4);
`○ ! ? ⍲ ⍱` (T5); scan, axis, N-wise reduction, expand (T6);
`, ⍪ ⌽ ⊖ ⍉ ↑ ↓ ∊ ~ ⍋ ⍒ ⊥ ⊤ ⌹` (T7); `⍕ ⍎ ⎕PW` (T8); nested arrays (T9);
`⋄`, function editing, `⎕FX ⎕CR ⎕NC ⎕EX`, functions as operands, defined
operators (T10); suspension, `)SI )RESET ⎕LC ⎕EA ⎕EM ⎕ET`, Ctrl-C (T11);
system commands and workspaces, `⎕LX`, `goapl -i` (T12); ASCII glyph
names, line editing with history and an APL keyboard (T13).

All tasks are done.
