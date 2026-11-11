⍝ Diamonds separate statements on a line; they run left to right
1+2 ⋄ 3+4
X←1 ⋄ Y←X+1 ⋄ X,Y
∇R←F N
R←0 ⋄ →(N>0)/MORE ⋄ R←¯1
→0
MORE:R←N
∇
F 5
F 0
⍝ A branch ends the rest of the line
∇R←G
R←1 ⋄ →0 ⋄ R←2
∇
G
⍝ Display a function with [⎕]; ∇ at the end closes the editor
∇F[⎕]∇
⍝ Edit: [n] goes to line n, [n.5] inserts, [∆n] deletes
∇F
[2] R←10 ⋄ →(N>0)/MORE
[1.5] ⍝ inserted
[5] ⍝ temporary
[⎕]
[∆5]
[2⎕]
∇
∇F[⎕]∇
⍝ Lines are renumbered on closing
∇F[⎕3]∇
⍝ Line 0 is the header
∇F[0] R←F N;I
[1] R←N×2
[⎕]
∇
F 4
⍝ Appending to a function goes after its last line
∇G
R←R+10
∇
G
⍝ ⎕CR gives the lines of a function; ⎕FX defines one
⎕CR 'G'
⍴⎕CR 'G'
⍴⎕CR 'NOSUCH'
⎕FX 'R←DBL X' 'R←2×X'
DBL 4
⎕FX 2 9⍴'R←INC X  R←X+1    '
INC 1
⎕FX ⎕CR 'G'
⍝ ⎕FX gives the index of a bad line
⎕FX 'R←' 'R←1'
⎕FX 'R←BAD X' 'L:1' 'L:2'
⍝ ⎕NC: 0 unused, 2 variable, 3 function, 4 operator, ¯1 invalid
⎕NC 'DBL'
⎕NC 'X'
⎕NC 'NOSUCH'
⎕NC '1A'
⎕NC '⎕IO'
⎕NC 2 3⍴'DBLX  '
⍝ ⎕EX erases names
⎕EX 'X'
⎕NC 'X'
⎕EX 'DBL'
⎕NC 'DBL'
⎕EX 2 4⍴'INC ⎕IO'
⍝ A function with two arguments may be called with one
∇R←A PL B
→(0=⎕NC 'A')/MON
R←A+B ⋄ →0
MON:R←B
∇
5 PL 3
PL 3
⍝ User and other functions as operands
∇R←A ADD B
R←A+B
∇
ADD/1 2 3 4
ADD\1 2 3
2 ADD/1 2 3 4
1 2 ADD.× 3 4
1 2∘.ADD 10 20
,/'AB' 'CD' 'EF'
⍴,/'AB' 'CD'
⍳/2 3⍴1 2 3
(2 2⍴⍳4)+.,2 2⍴⍳4
1 2∘.⍴1 2
ADD¨/(1 2)(3 4)
ADD/⍳0
⍝ A variable cannot be fixed as a function
DBL←0
⎕FX 'R←DBL X' 'R←2×X'
⎕EX 'DBL'
⍝ Defined operators: (F OP) or (F OP G) in the header
∇R←(F TWICE) X
R←F F X
∇
-TWICE 5
⌽¨TWICE 'AB' 'CD'
+/TWICE 2 2⍴⍳4
∇R←A (F AND G) B
R←(A F B),(A G B)
∇
6 +AND- 2
6 ×AND÷ 2
∇R←A (F WITH) B
R←A F B F A
∇
2 -WITH 5
⎕FX 'R←DBL X' 'R←2×X'
DBL TWICE 3
⍝ Operands may come from an enclosing operator
∇R←(F THRICE) X
R←F TWICE F X
∇
DBL THRICE 1
⎕NC 'TWICE'
⎕CR 'AND'
⍝ Errors
1 2 (3 ⋄ 4
∇F[⎕
∇NEW[2]
∇R←A (F) B
∇R←(F OP) (G X)
∇R←(F OP G H) X
∇PL B
1 ADD
-AND 1
TWICE 1
∇R←T
R←1
L:2
L:3
∇
T
