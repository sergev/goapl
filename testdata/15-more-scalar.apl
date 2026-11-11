⍝ ○X is pi times X
○1
○0.5 2
⍝ A○B: 1 2 3 sin cos tan, 4 sqrt 1+B*2, 5 6 7 sinh cosh tanh, 0 sqrt 1-B*2
1○○÷6
2○○1
3○○÷4
0○0.6
4○0.75
5 6 7○1
⍝ Negative A gives the inverse functions
¯1○1
¯2○0
¯3○1
¯4○¯2 2
¯5 ¯6 ¯7○1 1 0.5
⍝ Circular functions are pervasive
1 2○○0 0.5
(⍳2)∘.○0 1
⍝ !N is the factorial, or Γ(N+1) for other numbers
!0 1 5 10
!0.5
!¯0.5
!170
⍝ A!B is the binomial coefficient
2!5
0 1 2 3 4!4
5!2
2!¯3
¯3!¯1
¯1!3
1.5!4
⍝ Pascal's triangle
(0,⍳4)∘.!0,⍳4
⍝ ⍲ nand and ⍱ nor
0 0 1 1⍲0 1 0 1
0 0 1 1⍱0 1 0 1
⍝ ?N rolls a random integer from 1 to N; ⎕RL makes it repeatable
⎕RL←16807
?6 6 6 6 6 6
?10 100 1000
⎕RL
⎕RL←16807
?6 6 6 6 6 6
⍝ A?B deals A distinct integers from 1 to B
5?5
X←10?100
⍴X
∧/(X≥1)∧X≤100
+/+/X∘.=X
⍝ ? follows ⎕IO
⎕IO←0
⎕RL←16807
?6 6 6 6 6 6
3?3
⎕IO←1
⍝ Errors
!¯1
!¯2
!200
8○1
¯8○1
1.5○1
¯1○2
¯6○0.5
?0
?2.5
?¯1
3?2
1 2?3
2?3 4
¯1?3
'A'○1
⍲1
1⍲2
2⍱0
