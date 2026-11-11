⍝ Catenate joins along the last axis, ⍪ along the first, ,[K] along K
M←2 3⍴⍳6
M,7 8
M,0
M⍪7 8 9
M⍪0
M,[1]7 8 9
M,M
M⍪M
⍴(2 3 4⍴0),[2]2 1 4⍴1
'AB',[1]'CD'
1⍪2
⍝ A fractional axis laminates along a new axis
1 2 3,[0.5]4 5 6
1 2 3,[1.5]4 5 6
'AB',[0.5]'CD'
1,[0.5]2 3
⍴M,[0.5]M
⍴M,[1.5]M
⍴M,[2.5]M
⍝ Reverse and rotate; ⊖ and [K] work along the first axis or K
⌽1 2 3
⌽'HELLO'
⌽M
⊖M
⌽[1]M
⌽5
1⌽1 2 3 4
¯1⌽1 2 3 4
5⌽1 2 3 4
1⊖M
1 2⌽M
1 0 2⊖M
1⌽[1]M
⍝ Transpose reverses the axes; A⍉B moves axis I of B to A[I]
⍉M
⍉1 2 3
⍴⍉2 3 4⍴⍳24
2 1⍉M
1 1⍉3 3⍴⍳9
⍴2 1 3⍉2 3 4⍴⍳24
⍝ Take and drop; negative counts work from the end; take pads with fill
3↑1 2 3 4
¯2↑1 2 3 4
6↑1 2 3
¯6↑1 2 3
5↑'AB'
2 2↑M
¯1 ¯2↑M
3 4↑M
1↑M
3↑5
1↓1 2 3
¯1↓1 2 3
5↓1 2
1 1↓M
¯1↓M
0 ¯2↓M
⍝ First: ↑X is the first element
↑1 2 3
↑M
↑'ABC'
↑⍳0
⍝ Membership and without
2 3 4∊1 2 3
'HELLO'∊'AEIOU'
M∊2 4 6
5∊⍳0
1 2 3 4 5~2 4
'HELLO'~'L'
'A'~'B'
⍝ Grade up and down: indices that sort; equal items keep their order
⍋3 1 2 1
⍒3 1 2 1
X←5 2 8 1
X[⍋X]
X[⍒X]
⍋3 2⍴3 1 1 2 1 1
⍋'BCA'
⍝ With a collating sequence on the left
'ZYX'⍋'XYZ'
'ABC'⍋'CQB'
'ABC'⍒'CQB'
⍝ Decode and encode
10⊥1 2 3
2⊥1 0 1
24 60 60⊥1 2 3
1 2 3⊥4
10⊥2 3⍴1 2 3 4 5 6
10 10 10⊤123
2 2 2⊤5 6
0 10⊤123
10⊤123
2 2⊤5
⍝ Matrix inverse and matrix divide (least squares if more rows)
⌹2 2⍴2 0 0 4
(5 6)⌹2 2⍴1 2 3 4
⌹2
⌹1 2
⌹3 2⍴1 0 0 1 1 1
⌊0.5+(2 2⍴1 2 3 4)+.×⌹2 2⍴1 2 3 4
1 2 3⌹3 2⍴1 1 1 2 1 3
⍝ Errors
(2 2 2⍴1),1 2
M,1 2 3
M⍪1 2
M,[3]1
1 2,[2.5]3 4
1 2,[0.5]1 2 3
(2 2⍴1),[0.5]1 2
1.5⌽1 2
1 2 3⌽M
(2 2⍴1)⌽M
1 2 3⍉M
0 1⍉M
3 1⍉M
'A'↑1
1 2 3↑M
(2 2⍴1)↑M
1 2~2 2⍴1
⍋5
⍋1,'A'
1 2⍋'AB'
'AB'⍋1 2
1 2⊥1 2 3
'A'⊤1
⌹2 2⍴1 2 2 4
⌹1 2⍴1 2
1 2⌹1 2 3
⌹2 2 2⍴1
