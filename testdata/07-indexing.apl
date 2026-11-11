V←10 20 30 40 50
V[1]
V[5]
V[2 4]
V[3 3 1]
V[2 2⍴1 2 3 4]
⍴V[⍳0]
M←3 4⍴⍳12
M
M[2;3]
M[2;]
M[;3]
M[1 3;2 4]
M[;]
M[3 1;]
⍴M[,2;]
⍴M[2;]
C←2 3 4⍴⍳24
C[2;3;4]
C[1;;2]
C[;2;]
⍝ Indices are full expressions
V[1+1]
V[⍴V]
M[1+1;⍳2]
⍝ Any value can be indexed
(⍳5)[3]
(2 2⍴⍳4)[2;1]
⍝ Indexed assignment
V[2]←0
V
V[1 5]←¯1
V
V[3 4]←7 8
V
M[2;2]←0
M
M[1;]←100
M
M[;4]←1 2 3
M
M[1 3;1 3]←2 2⍴¯1 ¯2 ¯3 ¯4
M
⍝ Other variables holding the same value are not changed
A←1 2 3
B←A
B[1]←9
A
B
⍝ The value of an indexed assignment is the assigned value
X←V[1]←5
X
V
⍝ Errors
V[6]
V[0]
V[1.5]
V[1;1]
M[2]
M[4;1]
M[1;5]
5[1]
V[2
V[2;
V[1]←1 2
M[1;]←1 2
W[1]←0
V[6]←0
