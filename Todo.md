# Todo: toward APL2

goapl now implements the language of the Pascal prototype (steps 1–5 of
the README roadmap). This list grows it toward APL2, one task at a time.
Each task ends with: unit tests, a `testdata/` script (or additions to
one), README updates, and `make test` and `make compare` passing.

Already done: `⍝` comments; `⌈ ⌊ | ⍟` monadic and dyadic; compress `A/B`.

## T1. Decide on prototype compatibility

Several APL2 features conflict with the original:

- `○/` means log to base in the original; in APL2 `○` is pi-times and
  the circular functions.
- `+/5`, `+/,5`, `+/⍳0` are errors in the original; APL2 gives 5, 5
  and the identity element 0.
- Catenate on rank > 1, ranks above 3, `0÷0`, etc. are errors there.
- Error reporting: numbered prototype messages vs APL2 error classes
  (`DOMAIN ERROR`, `LENGTH ERROR`, `RANK ERROR`, `INDEX ERROR`,
  `VALUE ERROR`, `SYNTAX ERROR`, `WS FULL`).

Options: drop the quirks, or keep them behind a `-compat` flag.
Update `parity_test.go` and `tools/compare_cases.txt` to match.

## T2. Data model

Rework `Array` before adding primitives, so each one is written once:

- Arbitrary rank (remove `MaxRank`).
- Element types: numbers, characters, and later nested arrays; keep a
  fast path for all-numeric data.
- Prototype (fill) element for empty arrays.
- Shared axis helpers (split an array into frames along axis k).

## T3. System variables

`⎕IO` (index origin; affects `⍳`, indexing, `⍋`, axis numbers, branch
targets stay 1-based), `⎕CT`, `⎕PP`, `⎕RL` (random seed). Assignment
and localization in function headers.

## T4. Characters

- Quoted strings `'ABC'`, `''''` for a quote; a one-character string is
  a scalar.
- Display of character arrays and of mixed numbers and characters.
- `=` `≠` on characters; other scalar functions give an error.
- Structural functions and indexing on characters.
- `⍞` character input and output.

## T5. Remaining scalar functions

- `○` pi times; `A○B` circular functions (`¯7` to `7`).
- `!` factorial/gamma; `A!B` binomial.
- `?` roll; `A?B` deal (uses `⎕RL` from T3, so tests are repeatable).
- `⍲ ⍱` nand, nor.
- `0÷0` gives 1 as in APL2 (if chosen in T1).

## T6. Operators and axis

- Scan `f\`.
- Reduce and scan along an axis: `f/[k]`, `f\[k]`, and `⌿ ⍀`.
- Identity elements for reducing empty arrays.
- Compress and expand with axis: `A/[k]B`, `A\B`, `A⍀B`.
- Inner and outer product on any rank.

## T7. Structural functions

- Catenate and laminate on any rank with axis: `,[k]`, `,[0.5]`, `⍪`.
- `⌽ ⊖` reverse and rotate, with axis.
- `⍉` transpose, monadic and dyadic.
- `↑ ↓` take and drop.
- `∊` membership; `A~B` without.
- `⍋ ⍒` grade up and down (numbers, then characters with a collating
  sequence).
- `⊥ ⊤` decode and encode.
- `⌹` matrix inverse and matrix divide.
- Dyadic `⍳` and reshape on any rank.

## T8. Format and execute

- `⍕` format, monadic and dyadic (width and precision).
- `⍎` execute.
- Output of large arrays: wrap long lines at a print width (`⎕PW`).

## T9. Nested arrays

- Strand notation `A B (C D)`.
- `⊂` enclose, `⊃` disclose and first, `≡` depth and match, `∊` enlist.
- `¨` each.
- Pick `A⊃B`; selective and strand assignment `(A B)←1 2`.
- Nested display with boxes or spacing, as in APL2.
- Scalar functions extend into nested arrays (pervasion).

## T10. Functions

- Redefine and edit functions: `∇F[n]` line editing, `∇F[⎕]∇` display.
- `⎕FX`, `⎕CR`, `⎕NC`, `⎕EX`.
- Diamond `⋄` as statement separator.
- User functions as operands of operators (`F/`, `F¨`).
- Ambivalent functions (`⎕NC` of the left argument).
- Defined operators (APL2 `∇R←(F OP G) B`).

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
