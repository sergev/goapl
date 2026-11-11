⍝ An error in a function suspends it; its local names stay visible
∇R←DIV X;T
T←X×2
R←T÷X-1
∇
∇R←CALC X
R←1+DIV X
∇
CALC 1
X
T
⍝ )SI shows the functions in progress; * marks the suspended one
)SI
⎕LC
⍝ →N resumes at line N; here after fixing the local X
X←3
→2
⍝ The call completes and gives its result
)SI
⍝ → alone leaves the suspended function, abandoning its callers
CALC 1
→
)SI
X
⍝ Suspensions nest; )RESET clears them all
CALC 1
CALC 1
)SI
)RESET
)SI
⍝ →⎕LC runs the suspended line again
∇R←SAFE X
R←10÷X
∇
SAFE 0
X←5
→⎕LC
⍝ A syntax error suspends too
∇R←SYN X
R←X+
R←X
∇
SYN 4
→2
⍝ A branch past the end returns; here without a result
SYN 4
→3
)RESET
⍝ ⎕EM and ⎕ET describe the last error
⎕ET
⍴⎕EM
1÷0
⎕EM
⎕ET
⍝ A ⎕EA B executes B, or A if B fails; errors there do not suspend
'0' ⎕EA '2+3'
'0' ⎕EA '1÷0'
⎕ET
'''FAILED''' ⎕EA 'SAFE 0'
)SI
∇R←TRY X
R←'¯1' ⎕EA 'SAFE X'
∇
TRY 2
TRY 0
⍝ Errors
)FOO
⎕LC←1
⎕EA '1'
∇R←F X;⎕LC
