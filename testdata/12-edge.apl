⍝ Errors in any part of an expression stop it
A←1÷0
A
(1÷0)+1
1+1÷0
-1÷0
(1÷0)[1]
V←⍳5
V[1÷0]
V[1]←1÷0
V[1÷0]←1
V
⎕←1÷0
→1÷0
⍝ Errors inside products
1 2+.÷0 1
1 1÷.×0 0
⍝ Results above rank 3
(2 2 2⍴1)+.×2 2 2⍴1
M←2 2⍴⍳4
M[2 2⍴1;2 2⍴1]
+/2 2 2 2⍴⍳16
A←2 2 2 2⍴⍳16
A[2;1;2;2]
A[;;1;1]
⍝ Parse errors in nested parts
(1 2 (3))
-)
V[(]←1
V[1]←)
V[1 2)]
⎕←)
→(
⍝ Functions that are not available
↑5
1↓2
⍝ Infinity
*1000
-*1000
⍝ Numbers too large to read
1E999
⍝ Names repeated in a header are localized once
∇R←SAME X;X
R←X
∇
X←1
SAME 2
X
⍝ A lexical error while defining drops the line; definition goes on
∇R←LEX X
R←1 # 2
R←X+1
∇
LEX 1
⍝ A syntax error in a function shows up when the line runs
∇R←SYN X
R←1 2 (3)
∇
SYN 1
