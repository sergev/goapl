⍝ Index generator
⍳5
⍳1
⍳0
⍴⍳0
⍳,3
⍝ Shape
⍴5
⍴⍴5
⍴1 2 3
⍴2 3⍴⍳6
⍴2 3 4⍴⍳24
⍝ Reshape cycles the data
2 3⍴⍳6
2 3⍴1 2
3⍴5
2 2 2⍴⍳8
(,4)⍴1 2
0⍴1 2
⍴0 3⍴1
⍝ Empty data fills with zeros
3⍴⍳0
⍝ Ravel
,5
,1 2 3
,2 2⍴⍳4
⍴,2 3 4⍴1
⍝ Index of
10 20 30⍳20
10 20 30⍳40
10 20 30⍳2 2⍴10 40 30 20
1 2 1⍳1
(⍳0)⍳1
⍝ Catenate
1 2,3 4
1,2 3
1 2,3
1,2
(⍳0),1
⍝ Compress and replicate (goapl addition)
1 0 1/1 2 3
1/5
0 1/2 3
2 0 1/1 2 3
3/7
⍴0/5
1 0 1/2 3⍴⍳6
⍝ Any rank; planes of rank 4 are separated by two blank lines
2 2 2 2⍴⍳16
⍴2 3 4 5⍴0
⍴,2 3 4 5⍴0
⍝ Errors
⍳1 2
⍳¯1
⍳2.5
(2 2⍴1)⍴5
¯1⍴5
1.5⍴5
(2 2⍴1)⍳1
5⍳1
(2 2⍴1),1
1,2 2⍴1
1 0/1 2 3
¯1/1
0.5/1
(2 2⍴1)/1 2
