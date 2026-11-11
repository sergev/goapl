⍝ Niladic, no result
∇HELLO
⎕←1 2 3
∇
HELLO
⍝ Niladic with result
∇R←PI
R←3.14159
∇
PI
2×PI
⍝ Monadic, no result
∇SHOW X
⎕←X×10
∇
SHOW 1 2
⍝ Monadic with result
∇R←SQ X
R←X×X
∇
SQ 3
SQ SQ 2
1+SQ 3
⍝ Dyadic, no result
∇A SHOW2 B
⎕←A,B
∇
1 SHOW2 2
⍝ Dyadic with result
∇R←A HYP B
R←((A*2)+B*2)*0.5
∇
3 HYP 4
(3 HYP 4)+5 HYP 12
⍝ Values of statements inside a function are printed
∇R←TRACE X
X
R←X+1
R
∇
Y←TRACE 5
Y
⍝ The definition may end with ∇ on the last line
∇R←DBL X
R←2×X∇
DBL 21
⍝ Arguments, result, locals and labels are local; other names are global
X←100
R←200
T←300
G←0
∇R←LOC X;T
T←X
G←X
R←T+1
∇
LOC 5
X
R
T
G
⍝ A called function sees the locals of its caller (dynamic scope)
∇R←OUTER X;V
V←X×10
R←INNER 1
∇
∇R←INNER Y
R←V+Y
∇
OUTER 4
V
⍝ Branch to a statement number
∇R←JUMP N
R←1
→4
R←99
R←R+N
∇
JUMP 5
⍝ →0 returns; a branch to an empty vector continues
∇R←SIGN X
R←0
→(X=0)/0
R←1
→(X>0)/0
R←¯1
∇
SIGN ¯5
SIGN 0
SIGN 7
⍝ A target beyond the last statement returns
∇R←OUT
R←1
→99
R←2
∇
OUT
⍝ A vector target uses its first element
∇R←FIRST
R←1
→4 1
R←2
R←R+10
∇
FIRST
⍝ Labels
∇R←SUM N;I
R←0
I←1
LOOP:→(I>N)/0
R←R+I
I←I+1
→LOOP
∇
SUM 100
⍝ Recursion
∇R←FACT N
→(N≤1)/ONE
R←N×FACT N-1
→0
ONE:R←1
∇
FACT 10
⍝ Comment lines count as statements
∇R←CMT X
⍝ line 1 is a comment
R←X ⍝ line 2
→5
R←0
∇
CMT 3
⍝ A function may call one defined later
∇R←CALLER X
R←LATER X
∇
∇R←LATER X
R←X+1
∇
CALLER 1
⍝ Header errors; after a rejected header, lines run as usual
∇R←HELLO X
∇
∇R←
∇R←A B C D
∇A B C D
∇1 F
∇R←F X;
⍝ A function may have no statements
∇EMPTY
∇
EMPTY
⍝ A function name cannot be assigned
SQ←1
⍝ Wrong number of arguments; a dyadic function called with one has no left argument
HYP 1
1 SQ 2
(SQ)
⍝ A function without result has no value
∇NORES
1
∇
NORES
1+NORES
⍝ The result must be assigned
∇R←NOVAL
1
∇
NOVAL
⍝ An error inside a function shows where it happened
∇R←BAD X
R←X÷0
∇
BAD 1
∇R←CALLBAD X
R←1+BAD X
∇
CALLBAD 1
⍝ An error suspends the function; )RESET leaves it and restores locals
)RESET
X
⍝ Branch target must be an integer
∇BADJUMP
→1.5
∇
BADJUMP
⍝ Unbounded recursion
∇R←INF X
R←INF X
∇
INF 1
