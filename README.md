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
| Added     | `⌈ ⌊` ceiling/floor, max/min; `\|` magnitude/residue; `⍟` natural log, log to base |
| Other     | `←` assignment, `A[I;J]` indexing and indexed assignment, `⎕` input/output, `→` branch, labels |
| Functions | `∇` definitions: niladic, monadic, dyadic, with or without explicit result |

Data are floating-point arrays of rank 0 to 3. Errors use the numbered
messages of the original `APLfile`. Comparisons use a tolerance of 1E¯13.
Output uses conventional APL notation (10 significant digits, `¯` for
negative numbers) rather than the original fixed-width columns.

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

Step 2 done: scalar functions, reduction, assignment, parentheses.
