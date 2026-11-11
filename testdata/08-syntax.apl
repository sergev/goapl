⍝ Parentheses
(1+2)×3
((1+2))
(((5)))
2×(3+4)×5
⍝ Assignment gives a value
A←B←3
A+B
1+C←2
C
⍝ The right argument is evaluated first
D+(D←4)
⍝ A variable can take a value of any shape
E←5
E←2 2⍴E
E
⍝ A label on an immediate line is ignored
L:1+1
⍝ A branch on an immediate line is ignored
→5
⍝ Errors
(1+2
1+2)
)
(
()
1 2 (3)
2+
+
←1
1←2
A B
UNDEF
1+UNDEF
→
→1.5
⍝ /* ends the session
/*
1+1
