# goapl

An APL2 interpreter written in Go.

It began as a re-implementation of the APL interpreter published in
*Pascal News* #26 (July 1983) and *The BYTE Book of Pascal* (1979), written
in CDC Pascal by V. Dichristofano, A. Kaniss, T. Robinson and J. Santini
(NADC / Villanova University), and implements APL2: nested arrays,
characters, the APL2 primitives and operators, defined operators, the
state indicator and workspaces.

The language is described in [Manual.md](Manual.md). Its examples are
session transcripts checked by `make test`.

## Build

Requires Go 1.27 or later. The only dependency is `golang.org/x/term`,
for the line editor.

```sh
make            # build ./goapl
make test       # go vet + unit and golden tests
make install    # install to ~/.local/bin (override with PREFIX=...)
make clean
make compare PROTO=/path/to/pascal-apl   # compare with the Pascal original
make pdf        # Manual.pdf from Manual.md (needs pandoc and typst)
```

## Usage

```sh
goapl                 # interactive session
goapl file.apl ...    # run script files
goapl -i ws.apl       # run a file (e.g. a saved workspace), then a session
```

Input is UTF-8 with standard APL glyphs. On a terminal the input line can
be edited, with history (kept in `~/.goapl_history`); glyphs are typed
with the prefix key `` ` `` or with Alt (`` `r `` gives `⍴`). Without an APL
keyboard, glyphs can be written by name: `2 3#rho#iota6`. Output to a
terminal is colored; set `NO_COLOR=1` to disable. `)OFF`, end of input or
a line starting with `/*` ends the session.

```apl
      +/⍳10
55
      ∇R←FACT N
[1] R←×/⍳N
[2] ∇
      FACT¨1 2 3 4 5
1 2 6 24 120
      'HELLO' 'WORLD'
 HELLO  WORLD
```

## Language

| Kind | Features |
|------|----------|
| Data | arrays of any rank; numbers, characters, mixed and nested arrays; strands `1 (2 3) 'AB'` |
| Scalar functions | `+ - × ÷ * ⍟ ○ ! ? ⌈ ⌊ \| ~ ∧ ∨ ⍲ ⍱ = ≠ < ≤ ≥ >`, pervasive into nested arrays |
| Structural | `⍴ , ⍪ ⌽ ⊖ ⍉ ↑ ↓ ⊂ ⊃ ∊ ⍳ ~ ⍋ ⍒ ⊥ ⊤ ⌹ ≡ ⍕ ⍎`, compress and expand `/ ⌿ \ ⍀`, indexing `A[I;J]` |
| Operators | reduce `f/`, N-wise reduce, scan `f\`, axis `[K]`, inner `f.g` and outer `∘.f` product, each `f¨`; any function as operand |
| Assignment | `←`, indexed `A[I]←`, strand `(A B)←`, selective `(2↑V)←` |
| Functions | `∇` definitions with result, arguments and locals; ambivalent functions; defined operators `∇R←A (F OP G) B`; labels, `→` branch, `⋄`; line editor `∇F[n]` |
| System variables | `⎕IO ⎕CT ⎕PP ⎕PW ⎕RL ⎕LX`, read-only `⎕LC ⎕EM ⎕ET` |
| System functions | `⎕FX ⎕CR ⎕NC ⎕EX ⎕EA`; I/O `⎕` and `⍞` |
| Errors | APL2 error classes; suspension on error, `→N` to resume, `)SI`, `)RESET`; Ctrl-C interrupts |
| Workspaces | `)FNS )VARS )ERASE )WSID )SAVE )LOAD )COPY )CLEAR )OFF`; saved workspaces are readable APL source |

## Tests

`make test` runs unit tests and golden tests. Each `testdata/*.apl` script
covers one area of the language, with `⍝` comments saying what each group
checks. It is run as an interactive session, and the transcript (input
after the prompt, then output) is compared with the matching `.out` file.
After an intended change in output, regenerate them with
`go test ./apl -update` and review the diff.

`apl/parity_test.go` ports every case of the original's test suite,
covering all 89 glyphs of its character set; deliberate differences are
marked there. `make compare` builds the Pascal original with Free Pascal
(into `build/`) and runs the cases of `tools/compare_cases.txt` through
both interpreters, comparing printed numbers and whether an error
occurred. Known differences, marked with `!`, come from bugs in the
original and from APL2 behavior that goapl follows instead.

## License

MIT License; see [LICENSE](LICENSE).
