⍝ Scan: item I is the reduction of the first I items
+\1 2 3 4 5
×\1 2 3 4 5
⌈\3 1 4 1 5
-\1 2 3 4
÷\1 2 3
∧\1 1 0 1
∨\0 0 1 0
≠\0 1 1 0
+\5
+\⍳0
⍴+\⍳0
+\2 3⍴⍳6
⍝ Reduce and scan along an axis; ⌿ and ⍀ use the first axis
M←2 3⍴⍳6
+/[1]M
+/[2]M
+⌿M
+\[1]M
+⍀M
-⌿3 1⍴10 3 2
A←2 3 4⍴⍳24
+/[2]A
+⌿A
⍴+/[1]A
⍴+/[2]A
⍴+/[3]A
⍝ Axis numbers follow ⎕IO
⎕IO←0
+/[0]M
⎕IO←1
⍝ The axis may be any expression
+/[⍴⍴M]M
⍝ Reducing an empty axis gives the identity element
+⌿0 3⍴0
×/[1]0 2⍴0
⍝ N-wise reduction: f/ over each window of N items
2+/1 2 3 4 5
3+/1 2 3 4 5
5+/1 2 3 4 5
6+/1 2 3 4 5
0+/1 2 3
2-/1 4 9 16
¯2-/1 4 9 16
2+/2 3⍴⍳6
2+/[1]3 2⍴⍳6
2+⌿3 2⍴⍳6
⍝ Compress along an axis; ⌿ uses the first axis
1 0 1/M
1 0/[1]M
0 1⌿M
1 0 1/[2]M
2 0 1/[2]M
⍝ A single number or a single item extends
2/1 2 3
1 0 1/5
1 1 0⌿1 3⍴7 8 9
⍝ Expand: zeros insert fill items
1 0 1 1 0\1 2 3
1 0 1\'AB'
1 0 1\5
1 0 1 1\M
1 0 1⍀M
1 0 1\[1]M
0 0\⍳0
⍴0 0⍀0 3⍴0
⍝ Inner and outer product on any rank
⍴(2 3 4⍴1)+.×4 5⍴1
(2 2 2⍴⍳8)+.×2 2⍴1 0 0 1
⍴(2 2⍴1)∘.×2 2 2⍴1
⍝ Errors
2+\1 2
+/[3]M
+/[0]M
+/[1.5]M
+/[1 2]M
+/['A']M
+/[]M
+[1]M
⍳[1]3
1 0/[3]M
1 0 1\[3]M
1 2\1 2
1 0 0\1 2
1 0 1\1 2 3
1 0/1 2 3
4+/1 2
1.5+/1 2
1 2+/1 2
