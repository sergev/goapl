# goapl

An APL interpreter written in Go.

It starts as a re-implementation of the APL interpreter published in
*Pascal News* #26 (July 1983) and *The BYTE Book of Pascal* (1979), written
in CDC Pascal by V. Dichristofano, A. Kaniss, T. Robinson and J. Santini
(NADC / Villanova University). The goal is to reproduce that language
first, then grow toward classic APL\360 / APL2.

## Build

Requires Go 1.27 or later.

```sh
make            # build ./goapl
make test       # go vet + unit and golden tests
make install    # install to /usr/local/bin (override with PREFIX=...)
make clean
```

## Usage

```sh
goapl                 # interactive, reads statements from stdin
goapl file.apl ...    # run script files
```

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

## Language (prototype level)

| Kind      | Glyphs |
|-----------|--------|
| Monadic   | `+` identity, `-` negate, `×` signum, `÷` reciprocal, `*` exp, `~` not, `⍳` index generator, `⍴` shape, `,` ravel |
| Dyadic    | `+ - × ÷ *`, `⍳` index of, `⍴` reshape, `,` catenate, `= ≠ < ≤ ≥ >`, `∧ ∨` |
| Operators | `f/` reduction (also `⌈ ⌊ ○`; `○/` is log to base, as in the original), `f.g` inner product, `∘.f` outer product |
| Added     | `⌈ ⌊` ceiling/floor, max/min; `\|` magnitude/residue; `⍟` natural log, log to base; `A/B` compress |
| Other     | `←` assignment, `A[I;J]` indexing and indexed assignment, `⎕` input/output, `→` branch, labels |
| Functions | `∇` definitions: niladic, monadic, dyadic, with or without explicit result |

Data are floating-point arrays of rank 0 to 3. Errors use the numbered
messages of the original `APLfile`. Comparisons use a tolerance of 1E¯13.
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
  statement. A function cannot be redefined (FUNCTION ALREADY DEFINED), as
  in the original.
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

`make test` runs unit tests and golden tests: each `testdata/*.apl` script
is run and its output compared with the matching `.out` file. After an
intended change in output, regenerate them with `go test ./apl -update`
and review the diff.

## Status

Step 4 done: the language of the original is implemented. Next: parity
tests against the Pascal prototype.
