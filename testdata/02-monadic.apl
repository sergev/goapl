⍝ Identity, negation, signum, reciprocal, exponential
+¯3 0 2.5
-¯3 0 2.5
×¯3 0 2.5
÷¯4 1 2
*0 1 2
⍝ Not
~0 1
~1 0 1 1
⍝ Ceiling, floor, magnitude, natural log (goapl additions)
⌈¯2.5 ¯2 2.3 2
⌊¯2.5 ¯2 2.3 2
|¯3 0 3
⍟1 10
⍝ Functions apply to each element of any shape
-2 3⍴⍳6
×2 2 2⍴¯1 0 1 2
⍝ An empty argument gives an empty result
⍴-⍳0
⍝ Errors
÷0
÷1 0
~2
~0.5
⍟0
⍟¯1
⍝ Functions without a monadic form
=5
≠5
<5
≤5
≥5
>5
∧1
∨1
