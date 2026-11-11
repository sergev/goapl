⍝ Arithmetic
7+2
7-2
7×2
7÷2
2*10
2*0.5
4*¯0.5
0*0
⍝ Comparisons
1 2 3=2
1 2 3≠2
1 2 3<2
1 2 3≤2
1 2 3≥2
1 2 3>2
⍝ Logical
0 0 1 1∧0 1 0 1
0 0 1 1∨0 1 0 1
⍝ Maximum, minimum, residue, logarithm (goapl additions)
3⌈1 5 3
3⌊1 5 3
3|0 1 2 3 4 ¯1 ¯2
¯3|1 2 ¯1
0|5
0.5|1.25
10⍟100 1000
2⍟8
⍝ A scalar extends to the shape of the other argument
10+1 2 3
1 2 3×10
⍝ A one-element array extends too
(,10)+1 2 3
⍝ Matrices
(2 2⍴1 2 3 4)+2 2⍴10 20 30 40
10×2 3⍴⍳6
⍝ Comparison tolerance
0.1+0.2
(0.1+0.2)=0.3
1=1+1E¯14
1=1+1E¯10
⍝ Evaluation is right to left
2×3+4
10-5-2
2*3*2
⍝ Errors
1÷0
0÷0
1 2 3÷1 0 1
0*¯1
1 2+1 2 3
(2 2⍴1)+2 3⍴1
(2 2⍴1)+1 2 3 4
2∧3
0.5∨1
¯8*1÷3
¯2⍟8
1⍟5
2~1
