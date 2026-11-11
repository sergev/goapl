# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

goapl is an APL interpreter in Go. It reimplements the 1978 Pascal APL
interpreter from *Pascal News* #26 (source in `../../Besm-6/pascal-apl/`,
`pnscanner.pas`), then grows toward APL2. `Todo.md` lists the next tasks
(T1–T13); `Manual.md` is the language specification.

## Commands

```sh
make                 # build ./goapl
make test            # go vet + go test ./...
go test ./apl -run TestParity/reduce_plus   # one test or subtest
go test ./apl -run TestGolden -update       # regenerate testdata/*.out
go test ./apl -run TestManual -update       # regenerate examples in Manual.md
make compare PROTO=/home/vak/Project/Besm-6/pascal-apl   # diff against the Pascal original (needs fpc)
make pdf             # Manual.pdf via pandoc + typst
```

After `-update`, review the diff: the expected outputs are generated from
goapl itself.

## Architecture

All of the interpreter is package `apl`; `cmd/goapl` is a thin CLI
(prompt when stdin is a tty, color when stdout is a tty and `NO_COLOR` unset).

Pipeline per line, driven by `Interp.Exec` (`workspace.go`):
`Lex` (`lexer.go`) → `Interp.parse` (`parser.go`) → `Interp.run`/`eval`
(`eval.go`). `Exec` also dispatches `∇` definition mode (`function.go`).

- **Values**: `Array{Shape, Data, Elems, Proto}` (`value.go`), any rank.
  A numeric array keeps `Data []float64` (the fast path); a general one
  keeps `Elems []Value` (`float64`, `Char` or nested `*Array`) and, when
  empty, its prototype in `Proto`. Build general arrays with `General`,
  which falls back to numeric when it can; `gather` selects elements for
  structural functions, `numeric` gives DOMAIN ERROR for general arrays,
  `axisSplit` views an array as frames along an axis. `format`
  prints characters touching and numbers set off by blanks. Arrays are
  immutable once built; indexed assignment copies.
- **Parser**: recursive descent read left to right that yields APL's
  right-to-left binding; grammar is in the comment at the top of
  `parser.go`. Parsing needs the workspace function table to classify
  names (variable vs niladic/monadic/dyadic user function), so user
  function bodies are lexed at definition but parsed at run time and cached
  per statement, invalidated by `Interp.fnGen`.
- **System variables** (`sysvar.go`): `⎕IO ⎕CT ⎕PP ⎕PW ⎕RL` live in `env`,
  embedded in `Interp`; primitives take `e *env`. `getVar`/`setVar`
  dispatch `⎕` names to `env`; a localized one keeps its value on entry.
- **Primitives**: scalar (pervasive) functions are table-driven in
  `scalar.go` (`scalarFns`: glyph → monad/dyad); structural ones in
  `structural.go` (`structFns`, with the functions in `rearrange.go`,
  `search.go` and `numeric.go`; `pick` and `gather` select elements);
  reduce, scan, inner and outer product in
  `operators.go`. Operators accept only dyadic scalar functions. An axis
  `[K]` is parsed into `fnNode.axis`; `axisIndex` resolves it, and
  structural functions take it through `structFn.dyadAxis`.
- **User functions**: `Interp.call` implements dynamic scoping by saving
  and restoring globals for args, result, `;locals` and labels; labels are
  local constants equal to their line number; `→` returns a target from
  `run`, out-of-range ends the call.
- **Errors**: `*Error{Class, Pos, Src, Where}` (`errors.go`); `Class` is an
  APL2 error class (`DomainError` prints `DOMAIN ERROR`). Errors propagate
  as Go errors to the top; `setPos`/`withSrc`/`locate` fill in position,
  source line and `F[n]`, and `Interp.report` prints class, line and caret
  `^`.

## Tests

- `apl/*_test.go`: unit tests, mostly checking error classes.
- `testdata/NN-*.apl` + `.out`: golden session transcripts per feature
  area, run with `Prompt` and `Echo` set (`golden_test.go`). `⍝` comments
  label each group.
- `apl/parity_test.go`: every case from the original's test suite;
  deliberate differences are marked `differs_*`.
- `apl/manual_test.go`: every ```` ```apl ```` block in `Manual.md` is run as a
  session and must match. Non-session blocks use ```` ```text ````. An output
  line starting with 6 spaces would be read as input, so avoid such examples.
- `tools/compare_cases.txt`: cases for `make compare`; `!` marks a known
  difference with the reason in the comment above it.

## Conventions

- Where the original and APL2 differ, follow APL2; record each deviation
  from the original in `parity_test.go`, `tools/compare_cases.txt` and the
  "Differences from the original" section of `Manual.md`.
- When the language changes, update `Manual.md` (examples are tested), the
  README language table, and add cases to the matching `testdata` script.
- Commit messages: a single short subject line, no body, no
  `Co-Authored-By`.
