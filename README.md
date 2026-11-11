# goapl

An APL interpreter written in Go.

It starts as a re-implementation of the APL interpreter published in
*Pascal News* #26 (July 1983) and *The BYTE Book of Pascal* (1979), written
in CDC Pascal by V. Dichristofano, A. Kaniss, T. Robinson and J. Santini
(NADC / Villanova University). It reproduced that language first and now
grows toward APL2; where the two differ, goapl follows APL2.

## Build

Requires Go 1.27 or later.

```sh
make            # build ./goapl
make test       # go vet + unit and golden tests
make install    # install to ~/.local/bin (override with PREFIX=...)
make clean
make compare PROTO=/path/to/pascal-apl   # compare with the Pascal original
make pdf        # Manual.pdf from Manual.md (needs pandoc and typst)
```

The language is described in [Manual.md](Manual.md). Its examples are
session transcripts checked by `make test`.

## Usage

```sh
goapl                 # interactive, reads statements from stdin
goapl file.apl ...    # run script files
```

Output to a terminal is colored; set `NO_COLOR=1` to disable.

Input is UTF-8 with standard APL glyphs. A line starting with `/*` ends the
session, as in the original.

```apl
      +/⍳10
55
      ∇R←FACT N
      R←×/⍳N
      ∇
      FACT 5
120
```

## Language

| Kind      | Glyphs |
|-----------|--------|
| Monadic   | `+` identity, `-` negate, `×` signum, `÷` reciprocal, `*` exp, `~` not, `⍳` index generator, `⍴` shape, `,` ravel |
| Dyadic    | `+ - × ÷ *`, `⍳` index of, `⍴` reshape, `,` catenate, `= ≠ < ≤ ≥ >`, `∧ ∨` |
| Operators | `f/` reduction (scalars, one element and empty axes as in APL2), `f.g` inner product, `∘.f` outer product |
| Added     | `⌈ ⌊` ceiling/floor, max/min; `\|` magnitude/residue; `⍟` natural log, log to base; `○` pi times, circular functions; `!` factorial, binomial; `?` roll, deal; `⍲ ⍱` nand, nor; `A/B` compress |
| Other     | `←` assignment, `A[I;J]` indexing and indexed assignment, `⎕` input/output, `⍞` character input/output, `→` branch, labels |
| System    | `⎕IO` index origin, `⎕CT` comparison tolerance, `⎕PP` print precision, `⎕RL` random link; localizable in headers |
| Functions | `∇` definitions: niladic, monadic, dyadic, with or without explicit result |

Data are arrays of any rank of numbers (floating point) and characters (`'IT''S'`); `=` and `≠` compare characters. Errors are reported by
APL2 class (`DOMAIN ERROR`, `LENGTH ERROR`, ...). `⍝` starts a comment.
Comparisons use a tolerance of 1E¯13 (`⎕CT`).
Indexing is more general than in the original: indices may be arrays
(`V[2 4]`, `M[1 3;]`) and an axis may be elided (`M[;2]`); the original
accepted scalar indices only. Output uses conventional APL notation (10 significant digits, `¯` for
negative numbers) rather than the original fixed-width columns.

## User functions

```apl
∇R←FACT N;I
R←1
I←0
LOOP:→(I≥N)/0
I←I+1
R←R×I
→LOOP
∇
```

- The header is `∇[R←][A] F [B]` with optional local names `;X;Y`.
  Arguments, result, locals and labels are local to the call (dynamic
  scope); other names are global.
- `→N` jumps to statement N; a target outside the function (e.g. `→0`)
  returns. `→` of an empty vector does nothing, so `→(cond)/L` branches
  conditionally. A label `L:` is a local constant equal to its line number.
- The definition ends with a line `∇`, or with `∇` at the end of the last
  statement. A function cannot be redefined yet (DEFN ERROR).
- An error inside a function aborts to the top level and shows where:
  `FACT[2] R←X÷0`.
- `⎕←X` prints X. `⎕` prompts with `⎕:`, reads a line from the same input
  and evaluates it as an expression.

## Roadmap

1. Skeleton: module, Makefile, tokenizer, REPL.
2. Core expressions: arrays, parser, scalar functions, output formatting.
3. Structural functions, reduction, inner/outer product, indexing.
4. User-defined functions, branching, `⎕` I/O.
5. Parity tests with the Pascal prototype.
6. Toward APL2: characters, arbitrary rank, more primitives and operators,
   comments, system commands.

## Tests

`make test` runs unit tests and golden tests. Each `testdata/*.apl` script
covers one area of the language (lexical rules, monadic and dyadic
functions, reduction, structural functions, products, indexing, syntax,
user functions, `⎕`, output format, edge cases, system variables, characters, more scalar functions), with `⍝` comments saying
what each group checks. It is run as an interactive session, and the
transcript (input after the prompt, then output) is compared with the
matching `.out` file.
`apl/parity_test.go` ports every case of the prototype's
`tests/run_symbol_tests.py`, covering all 89 glyphs of its character set;
the few deliberate differences are marked there.

`make compare` builds the Pascal prototype with Free Pascal (into `build/`)
and runs the cases of `tools/compare_cases.txt` through both interpreters,
comparing printed numbers (the original shows two decimals) and whether
an error occurred. Known differences are marked with `!` and explained in
that file. They come from bugs in the original (it cannot use or display a
function's explicit result, crashes on `⎕` input, computes `¯2*2` as 0.25,
and misses an unclosed parenthesis) and from APL2 behavior that goapl
follows instead. After an
intended change in output, regenerate them with `go test ./apl -update`
and review the diff.

## Status

Steps 1–5 done: the language of the original is implemented and passes
the prototype's test suite. Now growing toward APL2 (step 6); the tasks
are listed in [Todo.md](Todo.md).

## License

MIT License; see [LICENSE](LICENSE).
