⍝ ⍕X gives the characters that display X
⍕123
⍴⍕123
⍕¯1.5 2E10
⍴⍕1 2 3
⍕2 3⍴⍳6
⍴⍕2 3⍴1 10 100
⍴⍕2 2 2⍴⍳8
⍕'TEXT'
⍕1,'A'
'X=',⍕42
⍴⍕⍳0
⍝ ⍕ follows ⎕PP
⎕PP←3
⍕÷3
⎕PP←10
⍝ A⍕B: width and decimals for all columns, or for each
8 2⍕3.14159 ¯2
6 2 4 0⍕2 2⍴1.5 2 3 4
⍝ A single number gives the decimals; the width fits the column
2⍕3.14159 2
0⍕2.5 3.5
⍴2⍕2 3⍴⍳6
⍝ Width 0 fits too
0 1⍕10 200
⍝ Negative decimals give significant digits in E notation
¯3⍕1234 0.5
10 ¯2⍕¯0.000123
⍝ A number too wide for its field shows stars
3 1⍕100
⍝ ⍎X runs X as a statement and gives its value
⍎'1+2'
X←⍎'2×3'
X
⍎'Y←5'
Y
2×⍎'⍳3'
⍎⍕42
S←'+/⍳10'
⍎S
⍝ An empty statement gives no value
⍎''
Z←⍎''
⍝ ⍎ of a function without a result gives no value
∇NOVAL
⎕←'RAN'
∇
⍎'NOVAL'
⍝ Errors inside ⍎ show the executed text
⍎'1÷0'
⍎'UNDEF'
⍎'1+'
⍎'→1'
⍎2 2⍴'AB'
⍎1 2
'A'⍕1
1 2 3⍕1 2
(2 2⍴1)⍕1
2⍕'A'
⍝ ⎕PW folds long output lines; the rest is indented by six blanks
⎕PW
⎕PW←30
⍳20
2 20⍴⍳40
40⍴'ABCDEFGHIJ'
⎕PW←29
⎕PW←1001
⎕PW←80
