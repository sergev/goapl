⍝ A new session is a clear workspace
)WSID
)FNS
)VARS
⍝ )FNS and )VARS list names, from a given one on
X←1 2 3
M←2 3⍴'ABCDEF'
N←1 (2 3) ('AB' (4 5))
E←⍳0
C←''
S←'Q'
V←,5
F←0.1 ¯2.5E¯20 1E300
G←2 0⍴0
H←0⍴⊂1 2
J←1 'A' 2.5
K←⊂⊂1 2
∇R←FACT N
R←×/⍳N ⍝ it's a comment
∇
∇R←(F TWICE) Y
R←F F Y
∇
)FNS
)VARS
)VARS M
⍝ A clear workspace needs a name to be saved
)SAVE
⎕IO←0
⎕LX←'''WELCOME'''
)SAVE DEMO
)WSID
⍝ )CLEAR starts afresh, with default system variables
)CLEAR
)VARS
⎕IO
⍝ )LOAD restores everything and runs ⎕LX
)LOAD DEMO
)WSID
⎕IO
⎕IO←1
FACT 5
-TWICE 3
X M N E C S V F G H J K
⍴¨E C S V G H
≡K
X≡1 2 3
N≡1 (2 3) ('AB' (4 5))
F=0.1 ¯2.5E¯20 1E300
⍝ )SAVE without a name uses the workspace name
X←4 5 6
)SAVE
⍝ )ERASE removes names
)ERASE X FACT ZZ
)VARS X
)FNS
⍝ )COPY brings objects from a saved workspace
)COPY DEMO X FACT QQ
X
FACT 4
)ERASE X
)COPY DEMO
X
⍝ )WSID changes the name
)WSID OTHER
)WSID
⍝ A suspended function prevents )SAVE; )CLEAR clears the suspension
∇R←BAD
R←1÷0
∇
BAD
)SAVE
)SI
)CLEAR
)SI
)VARS
⍝ Errors
)LOAD NOSUCH
)COPY NOSUCH
)WSID A B
)FOO
)LOAD
⎕LX←1 2
⍝ )OFF ends the session
)OFF
'NOT SHOWN'
