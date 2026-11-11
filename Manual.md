---
title: goapl Language Manual
subtitle: An APL interpreter in Go
---

# Introduction

goapl is an interpreter for APL, a language of arrays. It reproduces the
APL interpreter written in Pascal by V. Dichristofano, A. Kaniss,
T. Robinson and J. Santini (NADC / Villanova University), published in
*Pascal News* #26 (1983) and *The BYTE Book of Pascal* (1979), with a few
additions. This manual describes the language as goapl implements it.

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
digits, `∆` or `_`. It is at most 10 characters long. Upper and lower case
are different.

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
| `⎕` | input and output |
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

Rank is at most 3; a larger rank gives *Greater than three dimensions*.
An array may be empty: `⍳0` is a vector of length 0.

```apl
      ⍴5

      ⍴1 2 3
3
      2 3⍴⍳6
1 2 3
4 5 6
```

Numbers are 64-bit floating point. Equality and the other comparisons
use a relative tolerance of 1E¯13, so tiny rounding errors do not
matter.

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
Variable not assigned a value
UNKNOWN+1
↑
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
Args. incompatible for dyadic operation
1 2+1 2 3
   ↑
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

Division by zero is an error, and so is a power that has no real value.

```apl
      1÷0
Attempted division by zero
1÷0
 ↑
      ÷0
Attempted inverse of zero
÷0
↑
      ¯8*0.5
Argument is not an integer
¯8*0.5
  ↑
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
Value not boolean
~2
↑
```

# Structural functions

## Index generator `⍳N`

The first N positive integers. N is a non-negative integer, a scalar or
one-element vector.

```apl
      ⍳5
1 2 3 4 5
      ⍴⍳0
0
```

## Index of `A⍳B`

For each element of B, its first position in vector A, or `1+⍴A` if it is
not there. The result has the shape of B.

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

`○/` is the same as `⍟/`, as in the original: `○/2 8` is the logarithm
of 8 to the base 2.

```apl
      ○/2 8
3
```

The argument must have at least two elements along its last axis (a
matrix with rows of one element is allowed).

```apl
      +/5
Argument is a scalar or empty vector
+/5
↑
      +/,5
Argument is a vector of length one
+/,5
↑
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

`X[I]` selects elements of vector X at positions I. Positions start at 1.
I may be any array of integers; the result has the shape of I.

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
Index out of range
V[6]
 ↑
      V[1.5]
Non-integer indices
V[1.5]
 ↑
      V[1;2]
Non-scalar indices
V[1;2]
 ↑
      V[1 2]←1 2 3
Assigned expression not a scalar
V[1 2]←1 2 3
↑
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

A matrix is printed by rows, with each column right-aligned. A
three-dimensional array is printed as a sequence of matrices separated by
a blank line. An empty array prints an empty line.

```apl
      2 3⍴1 ¯20 300 4 5 6
1 ¯20 300
4   5   6
      2 2 2⍴⍳8
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
Attempted division by zero
RATIO[1] R←X÷X-1
            ↑
```

# Errors

An error prints a message, the line, and a caret `↑` under the place of
the error. The statement has no effect after that point.

| Code | Message | Cause |
|:--|:--|:--|
| 1 | Invalid character follows negative sign | `¯` not followed by a number |
| 2 | Digit must follow a decimal point | `1.` |
| 3 | Extraneous chars. follow function header | too many names in a header |
| 4 | Invalid character encountered | glyph that goapl does not know |
| 5 | Function already defined | redefining a function |
| 6 | Illegal name to right of explicit result | `∇R←` without a name |
| 7 | Invalid function/argument name | header without a valid name |
| 8 | Result of assignment not valid variable | assigning to a function name |
| 10 | Invalid expression | syntax error, unknown function |
| 12 | Statement no. to branch to not integer | `→1.5` |
| 13 | Dyadic operator not preceded by primary | product used monadically |
| 15 | Mismatched parentheses | unbalanced `(` or `)` |
| 16 | Function calls nested too deep | more than 1000 nested calls |
| 18 | End of input | `⎕` at end of input |
| 19 | Value not boolean | `~ ∧ ∨` on a number other than 0, 1 |
| 20 | Attempted division by zero | `A÷0`, `0*¯1`, `1⍟B` |
| 21 | Argument not a scalar | `⍳1 2` |
| 22 | Argument is negative | `⍳¯1`, `¯1⍴5`, `⍟0` |
| 23 | Argument is not an integer | `⍳2.5`, `¯8*0.5` |
| 24 | Argument is a scalar or empty vector | `+/5`, `+/⍳0` |
| 26 | Invalid outer product expression | `∘.` with a non-scalar function |
| 27 | Invalid inner product expression | `f.g` with a non-scalar function |
| 29 | Left argument is not a vector | `A⍳B`, `A/B` with a matrix A |
| 34 | Invalid index expression | unclosed `[` |
| 35 | Non-scalar indices | wrong number of indices |
| 36 | Assigned expression not a scalar | wrong shape in `X[I]←Y` |
| 37 | Non-integer indices | `V[1.5]` |
| 38 | Index out of range | `V[0]`, `V[1+⍴V]` |
| 50 | Number and base of different sign | `¯2⍟8` |
| 51 | Argument is a vector of length one | `+/,5` |
| 52 | Args. not compatible for inner product | axis lengths differ |
| 53 | Argument(s) with rank greater than one | `,` with a matrix |
| 54 | Attempted inverse of zero | `÷0` |
| 55 | Args. incompatible for dyadic operation | shapes differ |
| 56 | Left argument not a vector | `A⍴B` with a matrix A |
| 60 | Greater than three dimensions | result of rank above 3 |
| 70 | Identifier too long | name over 10 characters |
| 72 | Invalid reduction operator | `f/` with a non-scalar function |
| 73 | Dyadic reduction reference | `A f/B` |
| 74 | Monadic reference to dyadic operator | `=5`, dyadic function with one argument |
| 75 | Function defined with no statements | `∇F` followed by `∇` |
| 100 | Variable not assigned a value | name without value; function without result |

# Differences from the original

goapl adds to the language of the Pascal original:

- `⌈ ⌊ | ⍟` as monadic and dyadic functions (the original had only
  `⌈/ ⌊/ ○/`), and compress `A/B`.
- Indexing with arrays and empty indices; indexed assignment of arrays.
- Labels usable as branch targets; local names `;X` in headers.
- `⎕` input evaluates any expression.
- `⍝` comments.
- Comparison tolerance.

goapl does not reproduce these defects of the original:

- A function result could not be used in an expression or displayed.
- `⎕` input crashed.
- `¯2*2` gave 0.25.
- An unclosed `(` went unnoticed.
- `1;2`, `1]2` and `1∩2` gave a value.

Output uses the usual APL format instead of fixed columns of `12.2`
numbers, and an undefined name gives *Variable not assigned a value*
instead of *Invalid expression*.

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
| `⍳` | 1 to Y | position of Y in X |
| `⍴` | shape of Y | Y reshaped to X |
| `,` | Y as vector | X joined to Y |
| `/` | | Y replicated by X |
| `f/` | reduce Y by f | |
| `f.g` | | inner product |
| `∘.f` | | outer product |
