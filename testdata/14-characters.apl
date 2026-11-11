⍝ A string is a character vector; one character is a scalar
'HELLO'
⍴'HELLO'
⍴'A'
⍴,'A'
⍴''
'A ⍝ not a comment'
⍝ Character matrices print without blanks between columns
2 3⍴'ABCDEF'
2 2 3⍴'ABCDEFGHIJKL'
3 4⍴'AB'
⍝ Mixed arrays: numbers are set off by blanks, characters touch
1,'AB',2
'X',1 2 3
2 3⍴'A',10,'B'
⍝ Structural functions and indexing work on characters
'ABCDE'[2 4]
'ABCDE'[3]
(2 3⍴'ABCDEF')[2;]
(2 3⍴'ABCDEF')[;1]
'HELLO',' ','WORLD'
,2 2⍴'ABCD'
1 0 1 0 1/'ABCDE'
3/'AB'
'ABCDE'⍳'DAZ'
5⍴'AB'
⍝ Fill element of an empty character array is a blank
3⍴''
⍴3⍴''
⍝ Indexed assignment may mix characters and numbers
S←'ABCDE'
S[2 4]←'XY'
S
S[1]←7
S
⍝ = and ≠ compare characters; a character never equals a number
'ABC'='ABD'
'ABC'≠'ABD'
'A'=65
'A'='A'
'ABCA'='A'
+/'MISSISSIPPI'='S'
'ABC'∘.='CBA'
'ABC'∧.='ABC'
=/'AA'
⍝ Other scalar functions reject characters
'A'+1
-'A'
'A'<'B'
⌈/'AB'
'A'∧1
'AB'+.×1 2
'AB'∘.+1 2
⍳'A'
'A'⍴1
1 2['A']
→'A'
⍝ ⎕ shows strings, ⎕← prints them
⎕←'DONE'
X←⎕
'ABC'
X,X
⍝ ⍞← prints without ending the line; ⍞ reads a line as characters
⍞←'NO NEWLINE'
2+2
∇R←ASK P
⍞←P
R←⍞
∇
N←ASK 'NAME? '
JOHN SMITH
⍴N
N
⍝ An empty line gives an empty vector
E←⍞

⍴E
⍝ A matrix through ⍞←
⍞←2 2⍴'ABCD'
'END'
