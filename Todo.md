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
operators (T10).

## T11. Errors and the state indicator

- Suspend on error instead of returning to top level.
- `)SI`, `)RESET`, `→` to resume, `⎕LC`.
- `⎕EA` execute alternate, `⎕EM`, `⎕ET`.
- Interrupt with Ctrl-C.

## T12. System commands and workspaces

- `)OFF`, `)FNS`, `)VARS`, `)ERASE`, `)CLEAR`, `)WSID`.
- `)SAVE`, `)LOAD`, `)COPY` with a readable text format (APL source
  that recreates the workspace).
- Run a workspace or script from the command line.

## T13. Input

- ASCII transliteration of glyphs (e.g. `#rho`), for terminals without
  an APL keyboard.
- Line editing and history in the REPL, with an APL keyboard map.
