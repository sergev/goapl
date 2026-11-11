---
title: goapl Language Manual
subtitle: An APL interpreter in Go
---

# Introduction

goapl is an interpreter for APL, a language of arrays. It reproduces the
APL interpreter written in Pascal by V. Dichristofano, A. Kaniss,
T. Robinson and J. Santini (NADC / Villanova University), published in
*Pascal News* #26 (1983) and *The BYTE Book of Pascal* (1979), and grows
it toward APL2. Where the two differ, goapl follows APL2. This manual
describes the language as goapl implements it.

## Running goapl

```text
goapl                 interactive session on the terminal
goapl < file.apl      read statements from standard input
goapl a.apl b.apl     run script files in one workspace
```

In a session, goapl shows a prompt of six spaces, reads a line, runs it,
and prints the result flush left. Examples in this manual are such
sessions: indented lines are typed, the others are printed.

```apl
      1+2
3
```

A line starting with `/*` ends the session (or the current script file).
End of input ends it too.

When the output goes to a terminal, goapl uses colors: results in green,
error messages in red. Set the environment variable `NO_COLOR` to turn
colors off.

# Characters and tokens

Input is UTF-8 text with APL glyphs. Spaces separate tokens; they are
needed only between names and numbers.

## Numbers

A number is written in decimal. The high minus `¯` marks a negative
number; `-` is a function, not part of the number. A number may have a
fraction and an exponent `E`.

```apl
      42
42
      ¯7
¯7
      .5
0.5
      ¯2.25
¯2.25
      1.5E3
1500
      2E¯3
0.002
```

A digit must follow the decimal point, and something must follow `¯`.

## Numeric vectors

Numbers separated by spaces form one vector.

```apl
      1 2 3
1 2 3
      ¯1 0.5   7
¯1 0.5 7
```

## Names

A name starts with a letter, `∆` or `_` and continues with letters,
digits, `∆` or `_`, and may be of any length. Upper and lower case are
different.

```apl
      RATE∆1←5
      RATE∆1
5
```

## Comments

`⍝` starts a comment; the rest of the line is ignored.

```apl
      1+2   ⍝ three
3
```

## Glyphs

| Glyph | Monadic | Dyadic |
|:--|:--|:--|
| `+` | identity | add |
| `-` | negate | subtract |
| `×` | signum | multiply |
| `÷` | reciprocal | divide |
| `*` | exponential | power |
| `⌈` | ceiling | maximum |
| `⌊` | floor | minimum |
| \| | magnitude | residue |
| `⍟` | natural logarithm | logarithm |
| `~` | not | |
| `= ≠ < ≤ ≥ >` | | comparisons |
| `∧ ∨` | | and, or |
| `⍳` | index generator | index of |
| `⍴` | shape | reshape |
| `,` | ravel | catenate |
| `/` | | compress |
| `f/` | reduction | |
| `f.g` | | inner product |
| `∘.f` | | outer product |

| Symbol | Meaning |
|:--|:--|
| `←` | assignment |
| `→` | branch |
| `[ ; ]` | indexing |
| `( )` | grouping |
| `⎕` | input and output; starts system variable names |
| `∇` | function definition |
| `:` | label |
| `⍝` | comment |

# Arrays

All data are arrays of numbers. An array has a *shape*, the list of its
lengths along each axis, and a *rank*, the number of axes.

| Rank | Name | Example |
|:--|:--|:--|
| 0 | scalar | `5` |
| 1 | vector | `1 2 3` |
| 2 | matrix | `2 3⍴⍳6` |
| 3 | three-dimensional array | `2 2 3⍴⍳12` |

Rank has no limit. An array may be empty: `⍳0` is a vector of length 0.

```apl
      ⍴5

      ⍴1 2 3
3
      2 3⍴⍳6
1 2 3
4 5 6
```

Numbers are 64-bit floating point. Equality and the other comparisons
use a relative tolerance of 1E¯13 (`⎕CT`), so tiny rounding errors do
not matter.

```apl
      (0.1+0.2)=0.3
1
```

*Boolean* values are 0 and 1. Functions that need booleans (`~ ∧ ∨`)
reject other numbers.

# Evaluation

## Monadic and dyadic

A function with an argument only on its right is *monadic*; with
arguments on both sides it is *dyadic*. The same glyph usually has both
uses.

```apl
      -5
¯5
      8-5
3
```

## Right to left

There is no precedence. A function takes as its right argument the
value of everything to its right. Read an expression from right to left.

```apl
      2×3+4
14
      10-5-2
7
```

Use parentheses to change the order.

```apl
      (2×3)+4
10
```

The right argument is evaluated before the left one.

```apl
      D+(D←4)
8
```

## Statements

A line holds one statement. Its value is printed, except when the
statement is an assignment, a branch, or `⎕←X` (which prints by itself).

# Variables

`←` assigns a value to a name. The assignment has the value it assigned,
so it can be used inside an expression.

```apl
      A←3
      A×2
6
      B←1+C←2
      B,C
3 2
```

A variable can take a value of any shape. Using a name with no value
gives *Variable not assigned a value*.

```apl
      UNKNOWN+1
VALUE ERROR
UNKNOWN+1
^
```

# Scalar functions

Scalar functions work on each element of their arguments independently.
The result has the shape of the arguments.

## Arguments of a dyadic scalar function

Both arguments must have the same shape, or one of them must be a scalar
or have one element. That one is used with every element of the other.

```apl
      1 2 3+10 20 30
11 22 33
      10×1 2 3
10 20 30
      (2 2⍴1 2 3 4)-1
0 1
2 3
      1 2+1 2 3
LENGTH ERROR
1 2+1 2 3
   ^
```

## Arithmetic

`+X` is X. `-X` negates. `×X` is the sign: ¯1, 0 or 1. `÷X` is 1 divided
by X. `*X` is e to the power X.

```apl
      +¯3 0 2
¯3 0 2
      -¯3 0 2
3 0 ¯2
      ×¯3 0 2
¯1 0 1
      ÷4 ¯2
0.25 ¯0.5
      *1
2.718281828
```

Dyadic `+ - × ÷` are the four operations; `A*B` is A to the power B.

```apl
      7+2
9
      7-2
5
      7×2
14
      7÷2
3.5
      2*10
1024
      2*0.5
1.414213562
```

Division by zero is an error, except `0÷0`, which is 1. A power that
has no real value is an error too.

```apl
      0÷0
1
      1÷0
DOMAIN ERROR
1÷0
 ^
      ÷0
DOMAIN ERROR
÷0
^
      ¯8*0.5
DOMAIN ERROR
¯8*0.5
  ^
```

## Ceiling, floor, maximum, minimum

`⌈X` rounds up, `⌊X` rounds down. `A⌈B` is the larger, `A⌊B` the smaller.

```apl
      ⌈2.3 ¯2.3
3 ¯2
      ⌊2.3 ¯2.3
2 ¯3
      3⌈1 5
3 5
      3⌊1 5
1 3
```

## Magnitude and residue

`|X` is the absolute value. `A|B` is the remainder of B divided by A; it
has the sign of A. `0|B` is B.

```apl
      |¯4 0 4
4 0 4
      3|7 ¯7
1 2
      ¯3|7
¯2
```

## Logarithms

`⍟X` is the natural logarithm; X must be positive. `A⍟B` is the logarithm
of B to the base A.

```apl
      ⍟10
2.302585093
      2⍟8
3
      10⍟1000
3
```

## Comparisons

`= ≠ < ≤ ≥ >` give 1 if the relation holds, else 0. They have no
monadic form.

```apl
      1 2 3=2
0 1 0
      1 2 3≠2
1 0 1
      1 2 3<2
1 0 0
      1 2 3≤2
1 1 0
      1 2 3≥2
0 1 1
      1 2 3>2
0 0 1
```

## Logical functions

`~X` is not. `A∧B` is and, `A∨B` is or. Arguments must be 0 or 1.

```apl
      ~1 0
0 1
      0 0 1 1∧0 1 0 1
0 0 0 1
      0 0 1 1∨0 1 0 1
0 1 1 1
      ~2
DOMAIN ERROR
~2
^
```

# Structural functions

## Index generator `⍳N`

The first N indices: `1 2 … N`, or `0 1 … N-1` when `⎕IO` is 0. N is a
non-negative integer, a scalar or one-element vector.

```apl
      ⍳5
1 2 3 4 5
      ⍴⍳0
0
```

## Index of `A⍳B`

For each element of B, the index of its first occurrence in vector A,
or `⎕IO+⍴A` if it is not there. The result has the shape of B.

```apl
      10 20 30⍳20 40
2 4
```

## Shape `⍴X`

The lengths of the axes of X, as a vector. A scalar has shape `⍳0`.

```apl
      ⍴2 3⍴⍳6
2 3
      ⍴⍴5
0
```

## Reshape `A⍴B`

An array of shape A filled with the elements of B in order, repeated as
needed. If B is empty, the result is filled with zeros. A is a scalar or
vector of non-negative integers.

```apl
      2 3⍴1 2
1 2 1
2 1 2
      5⍴7
7 7 7 7 7
      3⍴⍳0
0 0 0
```

## Ravel `,X`

The elements of X as a vector.

```apl
      ,2 2⍴⍳4
1 2 3 4
```

## Catenate `A,B`

Joins A and B into one vector. A and B are scalars or vectors.

```apl
      1 2,3 4
1 2 3 4
      0,⍳3
0 1 2 3
```

## Compress and replicate `A/B`

Each element of B is repeated as many times as the corresponding element
of A. With a boolean A this selects elements. A scalar A applies to every
element. For a matrix, it works along the rows.

```apl
      1 0 1 0/⍳4
1 3
      2 0 1/7 8 9
7 7 9
      1 0 1/2 3⍴⍳6
1 3
4 6
```

# Operators

An operator takes a function and makes a new function. Only dyadic scalar
functions can be used with operators.

## Reduction `f/X`

Puts f between the elements of X. As always, it evaluates from the right:
`-/1 2 3` is `1-(2-3)`.

```apl
      +/1 2 3 4
10
      -/1 2 3
2
      ×/⍳5
120
      ⌈/3 9 2
9
      ∧/1 1 0
0
```

For a matrix, reduction works along each row.

```apl
      +/2 3⍴⍳6
6 15
```

A scalar reduces to itself, and so does a single element.

```apl
      +/5
5
      +/,5
5
```

Reducing an empty axis gives the identity element of the function: the
value e such that `e f X` is X. For `⌈` and `⌊` it is the smallest and
largest number; `⍟` has none.

```apl
      +/⍳0
0
      ×/⍳0
1
      ⌊/⍳0
1.797693135E308
      +/2 0⍴0
0 0
      ⍟/⍳0
DOMAIN ERROR
⍟/⍳0
^
```

## Inner product `A f.g B`

Combines the last axis of A with the first axis of B: g is applied to
pairs of elements, and the results are reduced with f. `+.×` is the
matrix product. The two axes must have the same length; a scalar
argument extends.

```apl
      1 2 3+.×4 5 6
32
      (2 3⍴⍳6)+.×3 2⍴⍳6
22 28
49 64
      1 2 3∧.=1 2 3
1
```

## Outer product `A ∘.f B`

Applies f to every pair: one element from A, one from B. The shape of the
result is `(⍴A),⍴B`.

```apl
      1 2 3∘.×1 2 3 4
1 2 3  4
2 4 6  8
3 6 9 12
      (⍳3)∘.=⍳3
1 0 0
0 1 0
0 0 1
```

# Indexing

## Selecting elements

`X[I]` selects elements of vector X at positions I. Positions start at 1
(or at `⎕IO`; see [System variables](#system-variables)). I may be any
array of integers; the result has the shape of I.

```apl
      V←10 20 30 40
      V[2]
20
      V[4 1 1]
40 10 10
      V[2 2⍴1 2 3 4]
10 20
30 40
```

An array of rank R takes R indices separated by `;`. An empty index
selects the whole axis.

```apl
      M←3 4⍴⍳12
      M[2;3]
7
      M[2;]
5 6 7 8
      M[;1]
1 5 9
      M[1 3;2 3]
 2  3
10 11
```

Any value can be indexed, and indices are full expressions.

```apl
      (⍳5)[1+2]
3
```

## Indexed assignment

`X[I]←Y` replaces the selected elements. Y is a scalar, or has the shape
of the selection. Other variables holding the same value do not change.

```apl
      V←⍳5
      W←V
      V[2 4]←0
      V
1 0 3 0 5
      W
1 2 3 4 5
      M←2 3⍴0
      M[1;]←7 8 9
      M
7 8 9
0 0 0
```

## Index errors

```apl
      V←⍳5
      V[6]
INDEX ERROR
V[6]
 ^
      V[1.5]
DOMAIN ERROR
V[1.5]
 ^
      V[1;2]
RANK ERROR
V[1;2]
 ^
      V[1 2]←1 2 3
LENGTH ERROR
V[1 2]←1 2 3
^
```

The wrong number of indices also gives *Non-scalar indices*, as in the
original.

# Input and output

## Display

A statement's value is printed with up to 10 significant digits. Large
and small numbers use `E`. Negative numbers have `¯`.

```apl
      ÷3
0.3333333333
      2*40
1.099511628E12
      ¯1E¯7
¯1E¯7
```

A matrix is printed by rows, with each column right-aligned. An array of
rank 3 is printed as a sequence of matrices separated by a blank line;
each further axis adds one more blank line between its parts. An empty
array prints an empty line.

```apl
      2 3⍴1 ¯20 300 4 5 6
1 ¯20 300
4   5   6
      2 2 2⍴⍳8
1 2
3 4

5 6
7 8
      2 2 1 2⍴⍳8
1 2

3 4


5 6

7 8
```

## Output with `⎕←X`

`⎕←X` prints X and has the value X, so it shows intermediate results.

```apl
      1+⎕←2×3
6
7
```

## Input with `⎕`

`⎕` prints `⎕:`, reads a line, evaluates it as an expression, and gives
its value. If the line has an error, the message is printed and the line
is asked for again.

```apl
      X←⎕
⎕:
      2×⍳3
      X
2 4 6
```

# System variables

System variables are names starting with `⎕` that control the
interpreter. They are read and assigned like variables, and always have
a value.

| Name | Meaning | Default | Values |
|:--|:--|:--|:--|
| `⎕IO` | index origin | 1 | 0 or 1 |
| `⎕CT` | comparison tolerance | 1E¯13 | 0 to just below 1 |
| `⎕PP` | print precision, in digits | 10 | 1 or more; above 17 means 17 |
| `⎕RL` | random link, the seed of `?` | 16807 | 1 to ¯2+2*31 |

A value outside the allowed range is a `DOMAIN ERROR`; an unknown `⎕`
name is a `SYNTAX ERROR`.

```apl
      ⎕PP
10
      ⎕IO←2
DOMAIN ERROR
⎕IO←2
^
```

## Index origin `⎕IO`

`⎕IO` is the first index: of `⍳`, of the result of `A⍳B`, and of
indexing. Statement numbers, labels and branch targets do not depend on
it.

```apl
      ⎕IO←0
      ⍳5
0 1 2 3 4
      10 20 30⍳30
2
      V←10 20 30
      V[0]
10
      ⎕IO←1
```

## Comparison tolerance `⎕CT`

Two numbers are equal if they differ by at most `⎕CT` times the larger
magnitude. The tolerance applies to comparisons, `A⍳B`, floor and
ceiling (`⌊X` is rounded up if X is that close to the next integer), and
residue. With `⎕CT←0` comparisons are exact.

```apl
      1=1+1E¯10
0
      ⎕CT←1E¯9
      1=1+1E¯10
1
      ⎕CT←1E¯13
      ⌊2.99999999999999
3
```

## Print precision `⎕PP`

`⎕PP` is the number of significant digits shown for a number.

```apl
      ⎕PP←4
      ÷3
0.3333
      ⎕PP←10
```

## Localizing system variables

A system variable may be listed among the local names of a function.
Unlike other local names, it keeps its value at the start of the call;
assignments inside the function are undone when it returns.

```apl
      ∇R←ORIGIN0 N;⎕IO
[1] ⎕IO←0
[2] R←⍳N
[3] ∇
      ORIGIN0 3
0 1 2
      ⍳3
1 2 3
```

# User functions

## Definition

A function is defined by a header line starting with `∇`, the statements
of its body, and a line `∇`. While defining, the prompt shows the number
of the next statement.

```apl
      ∇R←AVG X
[1] R←(+/X)÷⍴X
[2] ∇
      AVG 1 2 3 4
2.5
```

The closing `∇` may also end the last statement.

```apl
      ∇R←DOUBLE X
[1] R←2×X∇
      DOUBLE 21
42
```

## Headers

A header names the function, its arguments, and optionally a result
variable. There are six forms:

| Header | Kind | Call |
|:--|:--|:--|
| `∇F` | niladic, no result | `F` |
| `∇R←F` | niladic, result | `F` |
| `∇F Y` | monadic, no result | `F 5` |
| `∇R←F Y` | monadic, result | `F 5` |
| `∇X F Y` | dyadic, no result | `1 F 2` |
| `∇R←X F Y` | dyadic, result | `1 F 2` |

```apl
      ∇R←A HYP B
[1] R←((A*2)+B*2)*0.5
[2] ∇
      3 HYP 4
5
      (3 HYP 4)+5 HYP 12
18
```

A function with a result is used like a primitive function. A function
without a result can only be called as a whole statement; using its value
gives *Variable not assigned a value*. The same happens when the result
variable is never assigned.

A function cannot be redefined: defining an existing name gives
*Function already defined*. A function must have at least one statement.

## Statements

Each statement of the body is numbered from 1. Statements are run in
order; their values are printed as in a session.

```apl
      ∇R←TRACE X
[1] X
[2] R←X+1
[3] ∇
      Y←TRACE 5
5
      Y
6
```

## Local names

The arguments, the result, and the labels are *local*: they exist only
while the function runs. More local names can be listed after the
header, each preceded by `;`. Other names are global.

When a function is called, the values that local names had outside are
saved, and they are restored when the function returns, even after an
error.

```apl
      I←100
      ∇R←SUM N;I
[1] R←0
[2] I←N
[3] L:R←R+I
[4] I←I-1
[5] →(I>0)/L
[6] ∇
      SUM 4
10
      I
100
```

Scope is dynamic: a function sees the local names of the functions that
called it.

```apl
      ∇R←OUTER X;V
[1] V←X×10
[2] R←INNER 1
[3] ∇
      ∇R←INNER Y
[1] R←V+Y
[2] ∇
      OUTER 4
41
```

## Branching

`→N` continues with statement N. If N is outside the function, for
example `→0`, the function returns. If N is an empty vector, nothing
happens and the next statement runs. If N is a vector, its first element
is used. N must be an integer.

This makes `→(condition)/L` a conditional branch: `1/L` is `L`, `0/L` is
empty.

```apl
      ∇R←SIGN X
[1] R←0
[2] →(X=0)/0
[3] R←1
[4] →(X>0)/0
[5] R←¯1
[6] ∇
      SIGN ¯5
¯1
      SIGN 0
0
```

A branch typed in a session is ignored.

## Labels

A statement may start with a label `NAME:`. Inside the function, the
label is a constant equal to the statement number. A label in a session
line is ignored.

```apl
      ∇R←FACT N
[1] R←1
[2] LOOP:→(N≤1)/0
[3] R←R×N
[4] N←N-1
[5] →LOOP
[6] ∇
      FACT 5
120
```

## Recursion

A function may call itself, and functions may call functions defined
later. Calls may nest up to 1000 deep.

```apl
      ∇R←FIB N
[1] R←N
[2] →(N≤2)/ONE
[3] R←(FIB N-1)+FIB N-2
[4] →0
[5] ONE:R←1
[6] ∇
      FIB 10
55
```

## Errors in functions

An error stops all running functions and returns to the session. The
message shows the function and statement number where it happened.

```apl
      ∇R←RATIO X
[1] R←X÷X-1
[2] ∇
      RATIO 1
DOMAIN ERROR
RATIO[1] R←X÷X-1
            ^
```

# Errors

An error prints the error class, the line, and a caret `^` under the
place of the error. The statement has no effect after that point.

```apl
      1 2+1 2 3
LENGTH ERROR
1 2+1 2 3
   ^
```

| Class | Cause |
|:--|:--|
| SYNTAX ERROR | malformed number, unknown glyph, unbalanced parentheses or brackets, misplaced function, assignment to a function name |
| VALUE ERROR | name without a value; function without a result used as a value |
| VALENCE ERROR | `=5`; a function given the wrong number of arguments |
| DOMAIN ERROR | `1÷0`, `~2`, `⍳¯1`, `¯8*0.5`, `V[1.5]`, `→1.5`, `⍟/⍳0`, `⎕IO←2` |
| LENGTH ERROR | shapes of the same rank differ; `X[I]←Y` with the wrong shape |
| RANK ERROR | ranks differ; a matrix where a vector is needed; wrong number of indices |
| INDEX ERROR | `V[0]`, `V[1+⍴V]` |
| DEFN ERROR | bad function header; redefining a function |
| NONCE ERROR | a glyph that goapl knows but does not implement yet |
| SYSTEM LIMIT | more than 1000 nested calls |
| INTERRUPT | `⎕` at end of input |

# Differences from the original

goapl adds to the language of the Pascal original:

- `⌈ ⌊ | ⍟` as monadic and dyadic functions (the original had only
  `⌈/ ⌊/ ○/`), and compress `A/B`.
- Indexing with arrays and empty indices; indexed assignment of arrays.
- Labels usable as branch targets; local names `;X` in headers.
- `⎕` input evaluates any expression.
- `⍝` comments.
- Comparison tolerance.
- Arrays of any rank (the original stopped at 3).

goapl follows APL2 where the original differs:

- Errors are reported by APL2 class (`DOMAIN ERROR`) instead of the
  original's numbered messages.
- Reduction of a scalar, of one element, and of an empty axis gives a
  value instead of an error.
- `0÷0` is 1.
- `○/` is not the logarithm; use `⍟/`.
- Names may be longer than 10 characters.
- A function may have no statements.

goapl does not reproduce these defects of the original:

- A function result could not be used in an expression or displayed.
- `⎕` input crashed.
- `¯2*2` gave 0.25.
- An unclosed `(` went unnoticed.
- `1;2`, `1]2` and `1∩2` gave a value.

Output uses the usual APL format instead of fixed columns of `12.2`
numbers.

# Quick reference

## Grammar

```text
statement  := → expr | label : expr | expr
expr       := fn expr
            | name ← expr
            | name [ index ] ← expr
            | ⎕ ← expr
            | operand [ fn expr ]
operand    := ( number... | name | niladic-fn | ⎕ | ( expr ) ) { [ index ] }
index      := [ expr ] { ; [ expr ] }
fn         := primitive | primitive / | primitive . primitive
            | ∘ . primitive | user-fn
name       := variable | ⎕IO | ⎕CT | ⎕PP | ⎕RL
header     := ∇ [ name ← ] [ name ] name [ name ] { ; name }
```

## Functions

| Function | Monadic `f Y` | Dyadic `X f Y` |
|:--|:--|:--|
| `+` | Y | X plus Y |
| `-` | negative Y | X minus Y |
| `×` | sign of Y | X times Y |
| `÷` | 1÷Y | X divided by Y |
| `*` | e to the Y | X to the Y |
| `⌈` | round up | larger |
| `⌊` | round down | smaller |
| \| | absolute value | Y modulo X |
| `⍟` | ln Y | log of Y base X |
| `~` | not Y | |
| `= ≠ < ≤ ≥ >` | | comparison |
| `∧ ∨` | | and, or |
| `⍳` | first Y indices | index of Y in X |
| `⍴` | shape of Y | Y reshaped to X |
| `,` | Y as vector | X joined to Y |
| `/` | | Y replicated by X |
| `f/` | reduce Y by f | |
| `f.g` | | inner product |
| `∘.f` | | outer product |
