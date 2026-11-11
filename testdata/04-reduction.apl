⍝ Every function that can be reduced
+/1 2 3 4
-/1 2 3 4
×/1 2 3 4
÷/1 2 3 4
*/2 3 2
=/1 1 1
≠/1 0 1
</1 2 3
≤/1 2 3
≥/3 2 1
>/3 2 1
∧/1 1 0
∨/0 0 1
⌈/3 9 2
⌊/3 9 2
|/3 10
⍟/2 8
!/3 5
○/2 8
⍲/1 1 0
⍱/0 0 1
⍝ Reduction runs along the last axis
+/2 3⍴⍳6
+/2 2 3⍴⍳12
-/3 1⍴5 6 7
⍝ Results can be used further
2×+/⍳10
+/+/2 3⍴⍳6
⍝ A scalar or one item reduces to itself
+/5
+/,5
⍴+/,5
⍝ An empty axis gives the identity element
+/⍳0
×/⍳0
∧/⍳0
∨/⍳0
⌈/⍳0
⌊/⍳0
+/2 0⍴0
⍴+/2 0⍴0
⍝ Errors
⍟/⍳0
÷/1 0
⍝ Only dyadic scalar functions can be reduced
~/1 0
⍳/1 2
,/1 2
/1 2
