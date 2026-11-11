⍝ ⎕← prints a value and passes it on
⎕←1 2 3
X←⎕←5
X
1+⎕←2
⍝ ⎕ reads a line and evaluates it
A←⎕
1 2 3
A
B←⎕
2×A
B
⎕+1
10
⍝ An error in the input line is reported and the line is asked again
C←⎕
1÷0
UNDEF
7
C
⍝ Empty input lines are skipped
D←⎕

8
D
⍝ ⎕ inside a function
∇R←ASK
R←⎕×2
∇
ASK
21
⍝ End of input while reading ⎕
⎕
