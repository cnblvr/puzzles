/*
Package sudoku_classic generates and assistants classic sudoku puzzle.

This puzzle is 9 lines with 9 digits on each line.
A correct puzzle satisfies the condition that each column, each row, and each of
the nine box 3x3 contain all the digits from 1 to 9.

 ╔═══════╤═══════╤═══════╗
 ║ 3 1 6 │ 5 7 9 │ 2 4 8 ║ a
 ║ 5 7 9 │ 2 4 8 │ 3 1 6 ║ b
 ║ 2 4 8 │ 3 1 6 │ 5 7 9 ║ c
 ╟───────┼───────┼───────╢
 ║ 4 8 3 │ 1 6 5 │ 7 9 2 ║ d
 ║ 1 6 5 │ 7 9 2 │ 4 8 3 ║ e
 ║ 7 9 2 │ 4 8 3 │ 1 6 5 ║ f
 ╟───────┼───────┼───────╢
 ║ 9 2 4 │ 8 3 1 │ 6 5 7 ║ g
 ║ 8 3 1 │ 6 5 7 │ 9 2 4 ║ h
 ║ 6 5 7 │ 9 2 4 │ 8 3 1 ║ i
 ╚═══════╧═══════╧═══════╝
   1 2 3   4 5 6   7 8 9

Context of methods:
 - the lines a-i are app.Horizontal lines;
 - the lines 1-9 are app.Vertical lines;
 - the lines [1-3], [4-6], [7-9], [a-c], [d-f], [g-i] are "big" lines;
 - box 3x3 is a matrix with 3 rows and 3 columns, for example:
   [[a1,a2,a3],[b1,b2,b3],[c1,c2,c3]].
*/
package sudoku_classic
