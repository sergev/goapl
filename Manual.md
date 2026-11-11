---
title: goapl Language Manual
subtitle: An APL interpreter in Go
---

# Introduction

goapl is an interpreter for APL2, a language of arrays. It began as a
re-implementation of the APL interpreter written in Pascal by
V. Dichristofano, A. Kaniss, T. Robinson and J. Santini (NADC /
Villanova University), published in *Pascal News* #26 (1983) and *The
BYTE Book of Pascal* (1979); where that language and APL2 differ, goapl
follows APL2. This manual describes the language as goapl implements it.

## Running goapl

```text
goapl                 interactive session on the terminal
goapl < file.apl      read statements from standard input
goapl a.apl b.apl     run script files in one workspace
goapl -i ws.apl       run a file, then continue with a session
```

In a session, goapl shows a prompt of six spaces, reads a line, runs it,
and prints the result flush left. Examples in this manual are such
sessions: indented lines are typed, the others are printed.

```apl
      1+2
3
```

A line starting with `/*` ends the session (or the current script file).
End of input or `)OFF` ends it too.

When the output goes to a terminal, goapl uses colors: results in green,
error messages in red. Set the environment variable `NO_COLOR` to turn
colors off.

## Typing APL

On a terminal, goapl edits the input line:

| Key | Effect |
|:--|:--|
| Left, Right, Ctrl-B, Ctrl-F | move the cursor |
| Home, End, Ctrl-A, Ctrl-E | go to the start or end of the line |
| Backspace, Delete, Ctrl-D | delete a character |
| Ctrl-K, Ctrl-U | delete to the end or the start of the line |
| Up, Down, Ctrl-P, Ctrl-N | recall earlier lines |
| Ctrl-C | abandon the line |
| Ctrl-D on an empty line | end the session |

Lines are kept in `~/.goapl_history` for later sessions. Ctrl-C while a
function runs interrupts it.

APL glyphs are typed with a prefix: the backquote `` ` `` and then a
key, or the key with Alt. The keyboard follows the common APL layout:

| Key | Key | Key | Key | Key | Key |
|:--|:--|:--|:--|:--|:--|
| `a` ⍺ | `b` ⊥ | `c` ∩ | `d` ⌊ | `e` ∊ | `f` _ |
| `g` ∇ | `h` ∆ | `i` ⍳ | `j` ∘ | `k` `'` | `L` ⍞ |
| `l` ⎕ | `m` \| | `n` ⊤ | `o` ○ | `p` * | `q` ? |
| `r` ⍴ | `s` ⌈ | `t` ~ | `u` ↓ | `v` ∪ | `w` ⍵ |
| `x` ⊃ | `y` ↑ | `z` ⊂ | `#` ⍒ | `$` ⍋ | `%` ⌽ |
| `&` ⊖ | `'` ⍕ | `(` ⍱ | `)` ⍲ | `*` ⍟ | `+` ⌹ |
| `,` ⍝ | `-` × | `.` ⍀ | `/` ⌿ | `0` ∧ | `1` ¨ |
| `2` ¯ | `3` < | `4` ≤ | `5` = | `6` ≥ | `7` > |
| `8` ≠ | `9` ∨ | `:` ≡ | `;` ⍎ | `<` ⍪ | `=` ÷ |
| `[` ← | `]` → | `^` ⍉ | `` ` `` ⋄ |  |  |

Without an APL keyboard, glyphs can also be written by name: `#rho`
stands for `⍴`. Names work in any input except `⍞`, but not inside
strings or comments; `#comment` starts a comment. Case does not matter,
and the longest known name is taken, so `#iota5` is `⍳5`.

```apl
      2 3#rho#iota6
1 2 3
4 5 6
      +/#iota10
55
```

| Glyph | Names |
|:--|:--|
| × | `#times` |
| ÷ | `#div` `#divide` |
| ⍴ | `#rho` |
| ⍳ | `#iota` |
| ⌈ | `#max` `#ceil` |
| ⌊ | `#min` `#floor` |
| ⍟ | `#log` |
| ○ | `#circle` |
| ≤ | `#le` |
| ≥ | `#ge` |
| ≠ | `#ne` |
| ∧ | `#and` |
| ∨ | `#or` |
| ⍲ | `#nand` |
| ⍱ | `#nor` |
| ↑ | `#take` |
| ↓ | `#drop` |
| ⊂ | `#enclose` |
| ⊃ | `#disclose` `#pick` |
| ∊ | `#epsilon` `#in` |
| ⌽ | `#rotate` `#reverse` |
| ⊖ | `#flip` |
| ⍉ | `#transpose` |
| ⍋ | `#gradeup` |
| ⍒ | `#gradedown` |
| ⊥ | `#decode` |
| ⊤ | `#encode` |
| ⌹ | `#domino` |
| ⍕ | `#format` |
| ⍎ | `#execute` |
| ≡ | `#match` `#depth` |
| ¨ | `#each` |
| ∘ | `#jot` |
| ⌿ | `#slashbar` |
| ⍀ | `#backslashbar` |
| ⍪ | `#commabar` |
| ⎕ | `#quad` |
| ⍞ | `#quotequad` |
| ← | `#assign` `#gets` |
| → | `#goto` `#branch` |
| ∇ | `#del` |
| ∆ | `#delta` |
| ⍝ | `#comment` `#lamp` |
| ⋄ | `#diamond` |
| ¯ | `#neg` `#macron` |

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

## Strings

Characters between quotes form a character vector. A quote inside is
written twice. A single character in quotes is a scalar.

```apl
      'HELLO, WORLD'
HELLO, WORLD
      'IT''S'
IT'S
      ⍴'ABC'
3
      ⍴'A'

```

## Comments

`⍝` starts a comment; the rest of the line is ignored. A `⍝` inside
quotes is an ordinary character.

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
| `○` | pi times | circular functions |
| `!` | factorial | binomial |
| `?` | roll | deal |
| `~` | not | |
| `= ≠ < ≤ ≥ >` | | comparisons |
| `∧ ∨ ⍲ ⍱` | | and, or, nand, nor |
| `⍳` | index generator | index of |
| `⍴` | shape | reshape |
| `,` | ravel | catenate, laminate |
| `⍪` | | catenate along the first axis |
| `⌽ ⊖` | reverse | rotate |
| `⍉` | transpose | transpose |
| `↑` | first | take |
| `↓` | | drop |
| `∊` | enlist | membership |
| `~` | | without |
| `⍋ ⍒` | grade up, down | grade with collating sequence |
| `⊥` | | decode |
| `⊤` | | encode |
| `⌹` | matrix inverse | matrix divide |
| `⍕` | format | format with width and decimals |
| `⊂` | enclose | |
| `⊃` | disclose | pick |
| `≡` | depth | match |
| `f¨` | each | each |
| `⍎` | execute | |
| `/ ⌿` | | compress |
| `\ ⍀` | | expand |
| `f/ f⌿` | reduction | N-wise reduction |
| `f\ f⍀` | scan | |
| `f.g` | | inner product |
| `∘.f` | | outer product |

| Symbol | Meaning |
|:--|:--|
| `←` | assignment |
| `→` | branch |
| `[ ; ]` | indexing |
| `( )` | grouping |
| `⎕` | input and output; starts system variable names |
| `⍞` | character input and output |
| `' '` | string |
| `∇` | function definition |
| `:` | label |
| `⍝` | comment |
| `⋄` | statement separator |

# Arrays

All data are arrays. An element of an array is a number, a character,
or another array (see [Nested arrays](#nested-arrays)).
An array has a *shape*, the list of its lengths along each axis, and a
*rank*, the number of axes.

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

An array may mix numbers and characters. Arithmetic needs numbers;
structural functions and indexing take any elements.

```apl
      1,'AB',2
1 AB 2
      'ABCDE'[2 4]
BD
```

The *fill element* of an array is 0 if its first element is a number,
and a blank if it is a character. An empty array keeps the fill element
of the array it was made from. `A⍴B` fills with it when B is empty.

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
gives a `VALUE ERROR`.

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

## Circular functions

`○X` is pi times X. `A○B` applies to B the function numbered A:

| A | `A○B` | A | `A○B` |
|:--|:--|:--|:--|
| 0 | `(1-B*2)*0.5` | | |
| 1 | sine | ¯1 | arcsine |
| 2 | cosine | ¯2 | arccosine |
| 3 | tangent | ¯3 | arctangent |
| 4 | `(1+B*2)*0.5` | ¯4 | `(¯1+B*2)*0.5`, with the sign of B |
| 5 | hyperbolic sine | ¯5 | inverse hyperbolic sine |
| 6 | hyperbolic cosine | ¯6 | inverse hyperbolic cosine |
| 7 | hyperbolic tangent | ¯7 | inverse hyperbolic tangent |

Angles are in radians. A result that is not a real number is a
`DOMAIN ERROR`.

```apl
      ○1
3.141592654
      1○○÷6
0.5
      2○○1
¯1
      ¯3○1
0.7853981634
      0○0.6
0.8
      ¯1○2
DOMAIN ERROR
¯1○2
  ^
```

## Factorial and binomial

`!N` is the factorial of N. For a number that is not a non-negative
integer it is the gamma function of N+1; negative integers are a
`DOMAIN ERROR`.

`A!B` is the number of ways to choose A things out of B: `(!B)÷(!A)×!B-A`.
It is 0 if A is greater than B, and is extended to negative integers and
to other numbers as in APL2.

```apl
      !5
120
      !0.5
0.8862269255
      2!5
10
      (0,⍳4)∘.!0,⍳4
1 1 1 1 1
0 1 2 3 4
0 0 1 3 6
0 0 0 1 4
0 0 0 0 1
```

## Random numbers

`?N` is a random integer from 1 to N (from `⎕IO`). It is a scalar
function: `?6 6 6` rolls three dice. `A?B` deals A different integers
from 1 to B, in random order.

The numbers come from a generator whose state is `⎕RL`. Setting `⎕RL`
repeats the sequence.

```apl
      ⎕RL←16807
      ?6 6 6 6 6 6
1 5 3 4 2 1
      5?5
4 1 5 2 3
      ⎕RL←16807
      ?6 6 6
1 5 3
```

## Comparisons

`= ≠ < ≤ ≥ >` give 1 if the relation holds, else 0. They have no
monadic form. `=` and `≠` also compare characters; a character is never
equal to a number. The others need numbers.

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
      'ABCA'='A'
1 0 0 1
      'A'=65
0
```

## Logical functions

`~X` is not. `A∧B` is and, `A∨B` is or, `A⍲B` is nand (not and),
`A⍱B` is nor (not or). Arguments must be 0 or 1.

```apl
      ~1 0
0 1
      0 0 1 1∧0 1 0 1
0 0 0 1
      0 0 1 1∨0 1 0 1
0 1 1 1
      0 0 1 1⍲0 1 0 1
1 1 1 0
      0 0 1 1⍱0 1 0 1
1 0 0 0
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
needed. If B is empty, the result is filled with its fill element: 0 or
blank. A is a scalar or vector of non-negative integers.

```apl
      2 3⍴1 2
1 2 1
2 1 2
      5⍴7
7 7 7 7 7
      3⍴⍳0
0 0 0
      2 5⍴'HELLO'
HELLO
HELLO
```

## Ravel `,X`

The elements of X as a vector.

```apl
      ,2 2⍴⍳4
1 2 3 4
```

## Catenate `A,B` and `A⍪B`

`A,B` joins A and B along the last axis, `A⍪B` along the first, and
`A,[K]B` along axis K. The other axes must agree. An argument with one
axis less is joined as a single item, and a scalar is repeated to fit.

```apl
      1 2,3 4
1 2 3 4
      0,⍳3
0 1 2 3
      M←2 3⍴⍳6
      M,7 8
1 2 3 7
4 5 6 8
      M⍪0
1 2 3
4 5 6
0 0 0
```

## Laminate `A,[K]B`

With a fractional K, `A,[K]B` joins A and B along a new axis of length
2, placed before axis ⌈K. A and B have the same shape, or one is a
scalar.

```apl
      1 2 3,[0.5]4 5 6
1 2 3
4 5 6
      1 2 3,[1.5]4 5 6
1 4
2 5
3 6
```

## Reverse and rotate `⌽X`, `A⌽B`

`⌽X` reverses the order along the last axis; `⊖X` along the first;
`⌽[K]X` along axis K. `A⌽B` rotates each line left by A (right if A is
negative); A is a scalar, or an array with one number per line.

```apl
      ⌽'HELLO'
OLLEH
      ⊖2 3⍴⍳6
4 5 6
1 2 3
      1⌽1 2 3 4
2 3 4 1
      ¯1⌽1 2 3 4
4 1 2 3
      1 2⌽2 3⍴⍳6
2 3 1
6 4 5
```

## Transpose `⍉X`, `A⍉B`

`⍉X` reverses the order of the axes: a matrix is flipped over its
diagonal. In `A⍉B`, axis I of B becomes axis `A[I]` of the result. If
two axes go to the same place, the result takes their diagonal.

```apl
      ⍉2 3⍴⍳6
1 4
2 5
3 6
      1 1⍉3 3⍴⍳9
1 5 9
      ⍴2 1 3⍉2 3 4⍴⍳24
3 2 4
```

## Take and drop `A↑B`, `A↓B`

`A↑B` takes the first `A[K]` items along axis K, or the last if `A[K]`
is negative. Taking more than there are pads with fill elements. `A↓B`
drops them instead. A has at most one number per axis of B; the other
axes are kept whole.

```apl
      3↑1 2 3 4
1 2 3
      ¯2↑1 2 3 4
3 4
      6↑1 2 3
1 2 3 0 0 0
      1↓1 2 3
2 3
      2 2↑3 3⍴⍳9
1 2
4 5
      ¯1 ¯1↓3 3⍴⍳9
1 2
4 5
```

## First `↑X`

The first element of X, or its fill element if X is empty.

```apl
      ↑5 6 7
5
      ↑'ABC'
A
      ↑⍳0
0
```

## Membership and without `A∊B`, `A~B`

`A∊B` is 1 for each element of A that occurs in B. `A~B` is the vector A
without the elements that occur in B. Comparisons use `⎕CT`.

```apl
      'HELLO'∊'AEIOU'
0 1 0 0 1
      1 2 3 4 5~2 4
1 3 5
```

## Grade `⍋X`, `⍒X`

`⍋X` gives the indices that would sort X ascending; `⍒X` descending.
Equal items keep their order. For a matrix, rows are compared. Numbers
compare by value, characters by their Unicode codes; with a character
vector on the left, `A⍋B` orders characters by their position in A.

```apl
      ⍋3 1 2 1
2 4 3 1
      X←5 2 8 1
      X[⍒X]
8 5 2 1
      'ZYX'⍋'XYZ'
3 2 1
```

## Decode and encode `A⊥B`, `A⊤B`

`A⊥B` is the value of the digits B in the number system with radices A.
`A⊤B` gives the digits of B in radices A; a radix 0 takes all that is
left. A single radix applies to every digit.

```apl
      10⊥1 2 3
123
      24 60 60⊥1 2 3
3723
      2⊥1 0 1
5
      10 10 10⊤123
1 2 3
      0 10⊤123
12 3
      2 2 2⊤5 6
1 1
0 1
1 0
```

## Matrix inverse and divide `⌹X`, `A⌹B`

`⌹X` is the inverse of a square matrix X. `A⌹B` solves the linear
equations `B+.×X` = A for X. If B has more rows than columns, the
result is the least-squares fit. A singular matrix is a `DOMAIN
ERROR`. A vector is a one-column matrix.

```apl
      ⌹2 2⍴2 0 0 4
0.5    0
  0 0.25
      (5 6)⌹2 2⍴1 2 3 4
¯4 4.5
      1 2 3⌹3 2⍴1 1 1 2 1 3
0 1
```

## Compress and replicate `A/B`

Each element of B is repeated as many times as the corresponding element
of A. With a boolean A this selects elements. A scalar A applies to every
element, and a single element of B is repeated to match A. For a matrix,
it works along the rows; `A⌿B` works along the columns, and `A/[K]B`
along axis K.

```apl
      1 0 1 0/⍳4
1 3
      2 0 1/7 8 9
7 7 9
      1 0 1/5
5 5
      1 0 1/2 3⍴⍳6
1 3
4 6
      0 1⌿2 3⍴⍳6
4 5 6
```

## Expand `A\B`

A is boolean. The elements of B go where A has 1, in order, and fill
elements (0 or blank) go where A has 0. The number of 1s in A must
match the length of B. `A⍀B` and `A\[K]B` work along the first axis and
axis K.

```apl
      1 0 1 1 0\1 2 3
1 0 2 3 0
      1 0 1\'AB'
A B
      1 0 1⍀2 3⍴⍳6
1 2 3
0 0 0
4 5 6
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

## Scan `f\X`

Item I of the result is the reduction of the first I items of X:
`+\X` gives running sums. It works along the last axis, like reduction.

```apl
      +\1 2 3 4 5
1 3 6 10 15
      ×\1 2 3 4 5
1 2 6 24 120
      ⌈\3 1 4 1 5
3 3 4 4 5
      -\1 2 3 4
1 ¯1 2 ¯2
      +\2 3⍴⍳6
1 3  6
4 9 15
```

## Axis

Reduction and scan work along the last axis. `f⌿` and `f⍀` work along
the first axis. `f/[K]` and `f\[K]` work along axis K, counted from
`⎕IO`. A bad axis is an `AXIS ERROR`.

```apl
      M←2 3⍴⍳6
      +/M
6 15
      +⌿M
5 7 9
      +/[1]M
5 7 9
      +\[1]M
1 2 3
5 7 9
      +/[3]M
AXIS ERROR
+/[3]M
^
```

## N-wise reduction `N f/X`

Reduces each window of N consecutive items. A negative N reverses each
window first; N may be at most one more than the length.

```apl
      2+/1 2 3 4 5
3 5 7 9
      3+/1 2 3 4 5
6 9 12
      2-/1 4 9 16
¯3 ¯5 ¯7
      ¯2-/1 4 9 16
3 5 7
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

# Nested arrays

An element of an array may itself be an array that is not a simple
scalar. Such an element is an *item*, and the array is *nested*.

## Strands

Values written next to each other form a vector of those values. Use
parentheses to make an item of a vector or an expression. The numbers
of a constant like `1 2` are separate items.

```apl
      1 2 (3 4)
1 2  3 4
      ⍴1 2 (3 4)
3
      'AB' 'CD'
 AB  CD
      A←5
      B←6 7
      ⍴A B
2
```

## Display

A nested item is shown with a blank before it, so items stand apart.
Items that take several lines line up by rows.

```apl
      X←1 (2 3) 'ABC'
      X
1  2 3  ABC
      (2 2⍴⍳4) 'X'
 1 2 X
 3 4
```

## Depth `≡X`

0 for a simple scalar, 1 for a simple array, and one more than the
deepest item for a nested array.

```apl
      ≡5
0
      ≡1 2
1
      ≡1 (2 (3 4))
3
```

## Enclose and disclose `⊂X`, `⊃X`

`⊂X` is a scalar holding X; a simple scalar stays as it is. `⊃X` makes
one array of the items of X: their axes follow the axes of X, and
smaller items are padded with fill elements.

```apl
      ⊂1 2
 1 2
      ⍴⊂1 2

      ⊃(1 2)(3 4 5)
1 2 0
3 4 5
      ⊃'ABC' 'DEF'
ABC
DEF
```

Indexing a nested array gives scalars holding the items; `↑X` gives the
first item itself.

```apl
      X←1 (2 3) 'ABC'
      X[2]
 2 3
      ↑(1 2)(3 4)
1 2
```

## Pick `A⊃B`

Each element of A selects an item at the next level of B. For an item
of a matrix, the element of A is a vector of two indices, so it is
enclosed.

```apl
      2⊃1 (2 3) 'ABC'
2 3
      (2 1)⊃(1 2)(3 4 5)
3
      (⊂2 1)⊃2 2⍴'ABCD'
C
```

## Enlist `∊X`

All simple scalars of X, at any depth, as a vector.

```apl
      ∊1 (2 3) (4 (5 6))
1 2 3 4 5 6
```

## Match `A≡B`

1 if A and B have the same shape and the same items at every depth.

```apl
      (1 2)≡1 2
1
      'A'≡,'A'
0
```

## Scalar functions on nested arrays

Scalar functions apply to the simple scalars inside the items.

```apl
      1 2 (3 4)+10
11 12  13 14
      (1 2)(3 4)×10 20
 10 20  60 80
```

## Each `f¨`

`f¨X` applies f to each item of X; `A f¨B` to corresponding items, a
single item being used for all. f is any function, including user
functions and derived functions like `+/`.

```apl
      ⍴¨1 2 (3 4) 'ABCD'
     2  4
      +/¨(1 2)(3 4 5)
3 12
      1 2 3⍴¨4 5 6
 4  5 5  6 6 6
      ⌽¨'ABC' 'DE'
 CBA  ED
```

## Strand assignment `(A B)←X`

Assigns the items of X to the names in turn; a single item goes to
every name.

```apl
      (P Q)←10 (20 30)
      Q
20 30
```

## Selective assignment `(f V)←X`

The left side is an expression that selects elements from a variable:
the variable at its right end, with indexing and the functions
`↑ ↓ , ⍪ ⌽ ⊖ ⍉ / ⌿ \ ⍀ ⍴` (other parts, like the left argument of `/`,
are ordinary expressions). The selected elements get the items of X.

```apl
      V←⍳5
      (2↑V)←0
      V
0 0 3 4 5
      ((V>2)/V)←9
      V
0 0 9 9 9
      M←3 3⍴0
      (1 1⍉M)←1
      M
1 0 0
0 1 0
0 0 1
```

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

Characters are printed as they are, with no blanks between them. In a
mixed array, numbers are set off from their neighbors by a blank.

```apl
      2 3⍴'ABCDEF'
ABC
DEF
      2 3⍴'A',10,'B'
A 10 B
A 10 B
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

## Character input and output with `⍞`

`⍞←X` prints X without ending the line. `⍞` reads a line and gives it
as a character vector, with no evaluation; output of `⍞←` on the same
line serves as a prompt. Lines read by `⍞` have no indent.

```text
      ∇R←ASK P
[1] ⍞←P
[2] R←⍞
[3] ∇
      N←ASK 'NAME? '
NAME? JOHN SMITH
      ⍴N
10
```

## Format `⍕X`, `A⍕B`

`⍕X` gives the characters that display X: a character vector for a
scalar or vector, a matrix for a matrix. It follows `⎕PP`.

```apl
      ⍴⍕123
3
      'X=',⍕42
X=42
      ⍴⍕2 3⍴1 10 100
2 8
```

`A⍕B` formats numbers in fixed columns. A is a width and a number of
decimals, used for every column, or one such pair per column. A single
number gives just the decimals, and each column gets the width it needs
plus a blank; a width of 0 does the same. Negative decimals give that
many significant digits in `E` notation. A number too wide for its
field shows `*`.

```apl
      8 2⍕3.14159 ¯2
    3.14   ¯2.00
      2⍕3.14159 2
 3.14 2.00
      ¯3⍕1234 0.5
 1.23E3 5.00E¯1
      3 1⍕100
***
```

## Execute `⍎X`

`⍎X` runs the character vector X as a statement and gives its value. An
error inside shows the executed text after `⍎`. A branch cannot be
executed.

```apl
      ⍎'1+2'
3
      S←'+/⍳10'
      ⍎S
55
      ⍎'1÷0'
DOMAIN ERROR
⍎ 1÷0
   ^
```

## Print width `⎕PW`

A displayed line longer than `⎕PW` characters is folded. The
continuation lines are indented by six blanks, and lines break between
numbers where possible.

```text
      ⎕PW←30
      ⍳20
1 2 3 4 5 6 7 8 9 10 11 12 13
      14 15 16 17 18 19 20
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
| `⎕PW` | print width, in characters | 80 | 30 to 1000 |
| `⎕RL` | random link, the seed of `?` | 16807 | 1 to ¯2+2*31 |
| `⎕LC` | line numbers of running functions | | read only |
| `⎕EM` | message of the last error | | read only |
| `⎕ET` | type of the last error | 0 0 | read only |
| `⎕LX` | latent expression, run after loading | `''` | string |

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

The closing `∇` may also end the last statement. A line may hold several
statements separated by `⋄`; they run left to right, and a branch skips
the rest of the line.

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
gives a `VALUE ERROR`. The same happens when the result variable is never
assigned.

A function with two arguments may also be called with one: its left
argument then has no value, which `⎕NC` tells.

```apl
      ∇R←A PLUS B
[1] →(0=⎕NC 'A')/ONE
[2] R←A+B ⋄ →0
[3] ONE:R←B
[4] ∇
      2 PLUS 3
5
      PLUS 3
3
```

A header with a new name defines a function; a header that repeats an
existing function's name with other arguments is a `DEFN ERROR`. To
change a function, open it with its name alone (see
[Editing functions](#editing-functions)).

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

## Editing functions

`∇F` with the name of an existing function opens it for editing; the
prompt shows the number of the next line, after the last one. A line
may start with a command in brackets:

| Command | Effect |
|:--|:--|
| `[n]` | go to line n; text after it replaces that line |
| `[1.5]` | a fractional number inserts a line between 1 and 2 |
| `[0]` | line 0 is the header |
| `[⎕]` | show the whole function |
| `[⎕n]` | show from line n to the end |
| `[n⎕]` | show line n |
| `[∆n]` | delete line n |

Commands may also follow the name on the `∇` line, as in `∇F[⎕]∇`. The
closing `∇` numbers the lines again from 1.

```apl
      ∇R←SQUARE X
[1] R←X×X
[2] ∇
      ∇SQUARE
[2] [0] R←SQUARE X;T
[1] [0.5] T←X
[0.6] [1] R←T×T
[2] [⎕]
    ∇R←SQUARE X;T
[0.5]  T←X
[1]  R←T×T
    ∇
[2] ∇
      ∇SQUARE[⎕]∇
    ∇R←SQUARE X;T
[1]  T←X
[2]  R←T×T
    ∇
```

## Function representation `⎕CR`, `⎕FX`

`⎕CR 'F'` gives the lines of function F, header first, as a character
matrix; it is empty if F is not a function. `⎕FX` defines (or replaces)
a function from its lines: a character matrix, or a vector of strings.
It gives the function's name, or the index of the first line in error.

```apl
      ⎕FX 'R←DBL X' 'R←2×X'
DBL
      DBL 4
8
      ⍴⎕CR 'DBL'
2 7
      ⎕FX 'R←BAD X' 'L:1' 'L:2'
3
```

## Name class `⎕NC` and erase `⎕EX`

`⎕NC` tells what a name denotes: 0 nothing, 2 a variable, 3 a function,
4 an operator, ¯1 not a valid name. `⎕EX` erases a variable or function
and gives 1 (0 for a name it cannot erase). Both take a string, or a
matrix of names, one per row, and then give a vector.

```apl
      X←5
      ⎕NC 'X'
2
      ⎕FX 'R←DBL X' 'R←2×X'
DBL
      ⎕NC 'DBL'
3
      ⎕EX 'X'
1
      ⎕NC 2 3⍴'X  DBL'
0 3
```

## Functions as operands

Operators accept any dyadic function, including user functions and
derived functions. Such functions apply to items: each element is
passed as an array, and each result becomes an element. Reducing an
empty axis by such a function is a `DOMAIN ERROR`.

```apl
      ∇R←A ADD B
[1] R←A+B
[2] ∇
      ADD/1 2 3 4
10
      1 2∘.ADD 10 20
11 21
12 22
      ,/'AB' 'CD' 'EF'
 ABCDEF
      ADD¨/(1 2)(3 4)
 4 6
```

## Defined operators

An operator header puts the operator's name between its operands in
parentheses: `∇R←(F OP) X` for an operator taking one function,
`∇R←A (F OP G) B` for one taking two. In the body, F and G are used as
functions. The derived function, like `+OP`, may be monadic or dyadic
as the header says.

```apl
      ∇R←(F TWICE) X
[1] R←F F X
[2] ∇
      -TWICE 5
5
      +/TWICE 2 2⍴⍳4
10
      ∇R←A (F AND G) B
[1] R←(A F B),(A G B)
[2] ∇
      6 +AND- 2
8 4
```

## Errors in functions

An error in a function *suspends* it. The message shows the function and
statement number, and the session continues inside the function: its
local names are visible and can be changed. The functions that called it
wait, unfinished.

```apl
      ∇R←RATIO X
[1] R←X÷X-1
[2] ∇
      ∇R←CALC X
[1] R←1+RATIO X
[2] ∇
      CALC 1
DOMAIN ERROR
RATIO[1] R←X÷X-1
            ^
      X
1
      X←3
      →⎕LC
2.5
```

From a suspension:

- `→N` resumes the suspended function at line N; `→⎕LC` runs the failed
  line again, as above. The calls then complete as if nothing happened.
- `→` alone abandons the suspended function and its callers, back to the
  previous suspension or the session.
- `)RESET` abandons all suspensions.

Errors in a script suspend too; the lines after it run inside the
function until a `→` or `)RESET`.

## State indicator `)SI` and `⎕LC`

`)SI` lists the functions in progress, most recent first, with the
current line number; `*` marks a suspended one. `⎕LC` gives those line
numbers as a vector.

```apl
      ∇R←F X
[1] R←X÷0
[2] ∇
      ∇R←G X
[1] R←F X
[2] ∇
      G 1
DOMAIN ERROR
F[1] R←X÷0
        ^
      )SI
F[1] *
G[1]
      ⎕LC
1 1
      )RESET
```

## Interrupt

Ctrl-C stops a running function before its next line, with an
`INTERRUPT`, and suspends it like an error.

## Error handling `⎕EA`, `⎕EM`, `⎕ET`

`A ⎕EA B` executes the string B and gives its value; if B fails, it
executes A instead. Errors inside B do not suspend.

`⎕EM` holds the report of the last error as a character matrix, and
`⎕ET` its type as two numbers (0 0 if there was none):

| Class | `⎕ET` | Class | `⎕ET` |
|:--|:--|:--|:--|
| INTERRUPT | 1 1 | VALENCE ERROR | 5 1 |
| SYSTEM LIMIT | 1 4 | RANK ERROR | 5 2 |
| NONCE ERROR | 1 6 | LENGTH ERROR | 5 3 |
| SYNTAX ERROR | 2 1 | DOMAIN ERROR | 5 4 |
| DEFN ERROR | 2 2 | INDEX ERROR | 5 5 |
| VALUE ERROR | 3 1 | AXIS ERROR | 5 6 |

```apl
      '0' ⎕EA '1÷0'
0
      ⎕ET
5 4
      ⍴⎕EM
3 12
      ∇R←SAFEDIV X
[1] R←'0' ⎕EA '1÷X'
[2] ∇
      SAFEDIV 4
0.25
      SAFEDIV 0
0
```

# Workspaces

The variables, functions and system variables of a session form its
*workspace*. System commands, lines starting with `)`, manage it. A
command that cannot be done prints a message, not an error.

| Command | Effect |
|:--|:--|
| `)FNS [A]` | list functions and operators, from name A on |
| `)VARS [A]` | list variables, from name A on |
| `)ERASE A B …` | erase names |
| `)WSID [A]` | show the workspace name, or set it to A |
| `)SAVE [A]` | save the workspace to file A.apl, by default its name |
| `)LOAD A` | replace the workspace by the one saved in A.apl |
| `)COPY A [B …]` | copy all, or the named, objects of saved workspace A |
| `)CLEAR` | start a clear workspace |
| `)SI` | show the state indicator |
| `)RESET` | clear the state indicator |
| `)OFF` | end the session |

```apl
      X←1 2 3
      ∇R←SQ Y
[1] R←Y×Y
[2] ∇
      )VARS
X
      )FNS
SQ
      )WSID
CLEAR WS
      )SAVE DEMO
SAVED DEMO
      )CLEAR
CLEAR WS
      )VARS
      )LOAD DEMO
LOADED DEMO
      SQ X
1 4 9
      )COPY NOSUCH
WS NOT FOUND
```

A clear workspace cannot be saved without a name, and a workspace with
suspended functions cannot be saved. `)CLEAR` and `)LOAD` also clear the
state indicator.

## Workspace files

A saved workspace is APL source that recreates it: system variables and
variables as assignments of expressions for their exact values, and
functions in `∇` form. It can be read and edited, and run as a script.

```text
⍝ goapl workspace
⎕IO←1
⎕CT←1E¯13
⎕PP←10
⎕PW←80
⎕RL←16807
⎕LX←''
X←1 2 3
∇R←SQ Y
R←Y×Y
∇
```

## Latent expression `⎕LX`

`⎕LX` is a string run after the workspace is loaded, by `)LOAD` or from
the command line.

```apl
      ⎕LX←'''HELLO'''
      )SAVE GREET
SAVED GREET
      )LOAD GREET
LOADED GREET
HELLO
```

## Command line

`goapl file …` runs the files in one workspace, then `⎕LX` if set, and
ends. `goapl -i file …` continues with a session afterwards, so
`goapl -i DEMO.apl` loads a saved workspace and works in it. `)OFF` in a
file ends the program.

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
| AXIS ERROR | `+/[3]M` for a matrix M; an axis on a function without one |
| DEFN ERROR | bad function header; a header for an existing function; a bad editor command; a label used twice |
| NONCE ERROR | a glyph that goapl knows but does not implement yet |
| SYSTEM LIMIT | more than 1000 nested calls |
| INTERRUPT | Ctrl-C; `⎕` or `⍞` at end of input |

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
- Characters, strings and `⍞`.
- System variables `⎕IO ⎕CT ⎕PP ⎕PW ⎕RL`.
- `○ ! ?` monadic and dyadic, `⍲ ⍱`.
- Scan, axis, `⌿ ⍀`, N-wise reduction, expand.
- Catenate on any rank, laminate, `⍪ ⌽ ⊖ ⍉ ↑ ↓ ∊ ~ ⍋ ⍒ ⊥ ⊤ ⌹`.
- Format `⍕`, execute `⍎`, folding of long lines.
- Nested arrays: strands, `⊂ ⊃ ≡ ∊`, each, strand and selective
  assignment.
- `⋄`, function editing, `⎕FX ⎕CR ⎕NC ⎕EX`, any function as an operand,
  ambivalent functions, defined operators.
- Suspension on error, `)SI`, `)RESET`, `⎕LC`, `⎕EA ⎕EM ⎕ET`, Ctrl-C.
- System commands and workspaces saved as APL source; `⎕LX`.
- Line editing with history and an APL keyboard; ASCII glyph names.

goapl follows APL2 where the original differs:

- Errors are reported by APL2 class (`DOMAIN ERROR`) instead of the
  original's numbered messages.
- Reduction of a scalar, of one element, and of an empty axis gives a
  value instead of an error.
- `0÷0` is 1.
- `○/` is not the logarithm but the circular function; use `⍟/`.
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
line       := [ label : ] statement { ⋄ statement }
statement  := → expr | → | expr
expr       := fn expr
            | name ← expr
            | name [ index ] ← expr
            | ⎕ ← expr
            | ⍞ ← expr
            | ( name { name } ) ← expr
            | ( selection ) ← expr
            | strand [ fn expr ]
strand     := operand { operand }
operand    := ( number... | 'string' | name | niladic-fn | ⎕ | ⍞ | ( expr ) )
              { [ index ] }
index      := [ expr ] { ; [ expr ] }
fn         := ( f [ op ] [ [ expr ] ] | f . f | ∘ . f ) { ¨ | op | defined-op [ f ] }
f          := primitive | user-fn | ⎕FX | ⎕CR | ⎕NC | ⎕EX
op         := / | ⌿ | \ | ⍀
name       := variable | ⎕IO | ⎕CT | ⎕PP | ⎕PW | ⎕RL
header     := ∇ [ name ← ] [ name ] fname [ name ] { ; name }
fname      := name | ( name name [ name ] )
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
| `○` | pi times Y | circular function X of Y |
| `!` | factorial of Y | X things out of Y |
| `?` | random integer to Y | X random integers to Y |
| `~` | not Y | |
| `= ≠ < ≤ ≥ >` | | comparison |
| `∧ ∨ ⍲ ⍱` | | and, or, nand, nor |
| `⍳` | first Y indices | index of Y in X |
| `⍴` | shape of Y | Y reshaped to X |
| `,` | Y as vector | X joined to Y |
| `⍪` | | X joined to Y along the first axis |
| `⌽ ⊖` | Y reversed | Y rotated by X |
| `⍉` | Y transposed | Y with axes moved to X |
| `↑` | first of Y | first X of Y |
| `↓` | | Y without first X |
| `∊` | simple scalars of Y | X occurs in Y |
| `~` | | X without Y |
| `⍋ ⍒` | sorting indices of Y | same, collating sequence X |
| `⊥` | | Y in radix X |
| `⊤` | | digits of Y in radix X |
| `⌹` | inverse of Y | X divided by matrix Y |
| `⍕` | characters displaying Y | Y in fixed format X |
| `⊂` | Y as a scalar | |
| `⊃` | items of Y as one array | item X of Y |
| `≡` | depth of Y | X and Y are the same |
| `f¨` | f on each item of Y | f on each pair of items |
| `⎕FX` | define function from lines Y | |
| `⎕CR` | lines of function named Y | |
| `⎕NC` | class of name Y | |
| `⎕EX` | erase name Y | |
| `⎕EA` | | execute Y, or X on error |
| `⍎` | value of statement Y | |
| `/ ⌿` | | Y replicated by X |
| `\ ⍀` | | Y expanded by X |
| `f/ f⌿` | reduce Y by f | reduce windows of X items |
| `f\ f⍀` | scan Y by f | |
| `f.g` | | inner product |
| `∘.f` | | outer product |
