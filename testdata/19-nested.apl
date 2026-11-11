⍝ Strands: juxtaposed values make a vector; parentheses group an item
1 2 (3 4)
⍴1 2 (3 4)
'AB' 'CD'
⍴'AB' 'CD'
'A' 'B'
A←5
B←6 7
A B
⍴A B
(1 2)(3 4)
X←1 (2 3) 'ABC'
X
⍝ Depth: 0 for a simple scalar, 1 for a simple array, more for nesting
≡5
≡1 2
≡X
≡1 (2 (3 4))
⍝ Indexing gives scalars holding the items
X[2]
⍴X[2]
X[2 3]
⍝ Enclose makes a scalar; a simple scalar stays as it is
⊂1 2
⍴⊂1 2
≡⊂1 2
⊂5
(⊂1 2),⊂3 4
⍝ Disclose makes one array of the items, padding with fill
⊃(1 2)(3 4 5)
⊃'AB' 'CDE'
⊃1 (2 3)
⊃1 2 3
⍝ First and pick
↑X
↑(1 2)(3 4)
2⊃X
3⊃X
(2 1)⊃(1 2)(3 4 5)
(⊂2 1)⊃2 2⍴'ABCD'
⍝ Enlist: all simple scalars as a vector
∊1 (2 3) (4 (5 6))
∊'AB' 'CD'
∊5
⍝ Match
(1 2)≡1 2
(1 2)≡1 2 3
1 2 3≡,¨1 2 3
X≡1 (2 3) 'ABC'
'A'≡,'A'
⍝ Scalar functions reach into nested items
1 2 (3 4)+10
-1 (2 3)
(1 2)(3 4)×(10 20)
1 (2 3)=1 (2 4)
⍝ Each applies a function to every item
⍴¨1 2 (3 4) 'ABCD'
1 2 3⍴¨4 5 6
+/¨(1 2)(3 4 5)
⌽¨'ABC' 'DE'
2↑¨(1 2 3)(4 5 6)
1↓¨⊂'ABC'
∇R←SQ X
R←X×X
∇
SQ¨1 (2 3)
∇R←A ADD B
R←A+B
∇
1 2 ADD¨10 20
⍴¨⍳0
⍝ Strand assignment
(P Q)←10 20
P
Q
(P Q)←'X' (1 2)
Q
(P Q)←7
P+Q
⍝ Selective assignment: assign through a selection of a variable
V←⍳5
(2↑V)←0
V
((V>2)/V)←9
V
(⌽V)←⍳5
V
(¯2↑V)←'AB'
V
M←3 3⍴0
(1 1⍉M)←1
M
(,M)←⍳9
M
(M[2;])←0
M
U←1 2
(1 0 1\U)←7 8 9
U
⍝ Display of nested matrices
(2 2⍴⍳4) 'X'
2 2⍴(1 2)(3 4)'AB' 5
⍕1 (2 3)
⍴⍕1 (2 3)
⍝ Errors
¨1
1 2 3⍴¨1 2
(1 2)+(1 2 3)
3⊃1 2
(1 1)⊃1 2
(⊂1 1)⊃1 2
'A'⊃1 2
(2 2⍴1)⊃1 2
(P Q R)←1 2
(P Q)←2 2⍴1
(1+V)←1
(2↑V)←1 2 3
(2↑W)←1
1 2⊂3
SQ¨1 'A'
