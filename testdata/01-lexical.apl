⍝ Numbers
0
42
¯7
3.25
.5
¯.5
0.001
1E3
1.5E¯2
¯2E2
⍝ Vectors are numbers separated by spaces
1 2 3
¯1 0 1
1   2    3
⍝ Names: letters, digits, ∆ and _; up to 10 characters
A←1
ABC←2
A1B2←3
∆X←4
X_Y←5
A+ABC+A1B2+∆X+X_Y
ABCDEFGHIJ←6
ABCDEFGHIJ
abc←7
abc
⍝ Comments
1+2 ⍝ trailing comment
⍝ Empty lines are ignored

⍝ Lexical errors
¯A
¯
1.
1.+2
1 # 2
'A'
ABCDEFGHIJK←1
1E
