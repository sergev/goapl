⍝ Defaults
⎕IO
⎕CT
⎕PP
⎕RL
⍝ ⎕IO: index origin of ⍳, dyadic ⍳ and indexing
⎕IO←0
⍳5
10 20 30⍳30 40
V←10 20 30
V[0]
V[2]←99
V
M←2 3⍴⍳6
M[1;0 2]
V[3]
⎕IO←1
⍳5
V[3]
⍝ ⎕CT: comparison tolerance
1=1+1E¯10
⎕CT←1E¯9
1=1+1E¯10
1 2 3⍳2.0000000001
⎕CT←0
1=1+1E¯15
⎕CT←1E¯13
⍝ Floor, ceiling and residue are tolerant
⌊2.99999999999999
⌈3.00000000000001
⌊2.9
0.1|0.3
⍝ ⎕PP: significant digits shown
÷3
⎕PP←4
÷3
2*40
⎕PP←17
÷3
⍝ Values above 17 are taken as 17
⎕PP←30
⎕PP
⎕PP←10
⍝ ⎕RL: random link, the seed of ?
⎕RL←42
⎕RL
⍝ A one-element vector may be assigned
⎕IO←,0
⎕IO
⎕IO←1
⍝ Localized in a function header; the global value is kept outside
∇R←F X;⎕IO
⎕IO←0
R←⍳X
∇
F 3
⎕IO
⍝ A localized system variable starts with the global value
∇G;⎕PP
⎕←⎕PP
⎕PP←3
⎕←÷3
∇
G
÷3
⍝ Branch targets and labels stay 1-based
∇R←H N;⎕IO
⎕IO←0
R←0
L:R←R+1
→(R<N)/L
∇
H 4
⍝ Errors
⎕IO←2
⎕IO←0.5
⎕IO←0 1
⎕CT←1
⎕CT←¯1
⎕PP←0
⎕PP←1.5
⎕RL←0
⎕RL←2*31
⎕XY
⎕XY←1
⍝ System variables cannot be function or argument names
∇⎕IO
∇R←F ⎕IO
∇R←⎕CT F X
⍝ Only known system variables can be localized
∇K;⎕XY
