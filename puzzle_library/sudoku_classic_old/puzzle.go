package sudoku_classic_old

import (
	"context"
	"fmt"
	"github.com/cnblvr/puzzles/app"
)

// sudokuPuzzle is 9 lines with 9 digits on each line. A correct puzzle satisfies the condition that each column, each
// row, and each of the nine box 3x3 contain all of the digits from 1 to 9.
//  ╔═══════╤═══════╤═══════╗
//  ║ 3 1 6 │ 5 7 9 │ 2 4 8 ║ a
//  ║ 5 7 9 │ 2 4 8 │ 3 1 6 ║ b
//  ║ 2 4 8 │ 3 1 6 │ 5 7 9 ║ c
//  ╟───────┼───────┼───────╢
//  ║ 4 8 3 │ 1 6 5 │ 7 9 2 ║ d
//  ║ 1 6 5 │ 7 9 2 │ 4 8 3 ║ e
//  ║ 7 9 2 │ 4 8 3 │ 1 6 5 ║ f
//  ╟───────┼───────┼───────╢
//  ║ 9 2 4 │ 8 3 1 │ 6 5 7 ║ g
//  ║ 8 3 1 │ 6 5 7 │ 9 2 4 ║ h
//  ║ 6 5 7 │ 9 2 4 │ 8 3 1 ║ i
//  ╚═══════╧═══════╧═══════╝
//    1 2 3   4 5 6   7 8 9
// Context of methods:
//  - the lines a-i are data.Horizontal lines;
//  - the lines 1-9 are data.Vertical lines;
//  - the lines [1-3], [4-6], [7-9], [a-c], [d-f], [g-i] are "big" lines;
//  - box 3x3 is a matrix with 3 rows and 3 columns, for example:
//    [[a1,a2,a3],[b1,b2,b3],[c1,c2,c3]].
type sudokuPuzzle [][]int8

// Swap of lines within one "big" line.
func (p sudokuPuzzle) swapLines(dir DirectionType, lineA, lineB int) {
	if lineA == lineB {
		return
	}

	switch dir {
	case Horizontal:
		// Swap two horizontal lines
		temp := make([]int8, 9)
		copy(temp, p[lineA])
		copy(p[lineA], p[lineB])
		copy(p[lineB], temp)

	case Vertical:
		// Swap two vertical lines
		for row := 0; row < 9; row++ {
			p[row][lineA], p[row][lineB] = p[row][lineB], p[row][lineA]
		}
	}
}

func (p sudokuPuzzle) swapBigLines(dir DirectionType, lineA, lineB int) {
	if lineA == lineB {
		return
	}

	switch dir {
	case Horizontal:
		// Swap two horizontal "big" lines
		temp := make([]int8, 9)
		for i := 0; i < 3; i++ {
			copy(temp, p[lineA*3+i])
			copy(p[lineA*3+i], p[lineB*3+i])
			copy(p[lineB*3+i], temp)
		}

	case Vertical:
		// Swap two vertical "big" lines
		for row := 0; row < 9; row++ {
			for i := 0; i < 3; i++ {
				p[row][lineA*3+i], p[row][lineB*3+i] = p[row][lineB*3+i], p[row][lineA*3+i]
			}
		}
	}
}

// Reflect the entire puzzle in the typ direction.
func (p sudokuPuzzle) reflect(typ DirectionType) {
	switch typ {

	case Horizontal:
		for col := 0; col < 4; col++ {
			for row := 0; row < 9; row++ {
				p[row][col], p[row][8-col] = p[row][8-col], p[row][col]
			}
		}

	case Vertical:
		for row := 0; row < 4; row++ {
			p[row], p[8-row] = p[8-row], p[row]
		}
	}
}

// Rotate the entire puzzle in the rotation r.
func (p sudokuPuzzle) rotate(r RotationType) {
	r = r % 4
	switch r {

	case Rotate90:
		temp := make([][]int8, 9)
		for col := 8; col >= 0; col-- {
			for row := 0; row < 9; row++ {
				temp[8-col] = append(temp[8-col], p[row][col])
			}
		}
		for row := 0; row < 9; row++ {
			p[row] = temp[row]
		}

	case Rotate180:
		for row := 0; row < 4; row++ {
			for col := 0; col < 9; col++ {
				p[row][col], p[8-row][8-col] = p[8-row][8-col], p[row][col]
			}
		}
		for col := 0; col < 4; col++ {
			p[4][col], p[4][8-col] = p[4][8-col], p[4][col]
		}

	case Rotate270:
		temp := make([][]int8, 9)
		for col := 0; col < 9; col++ {
			for row := 8; row >= 0; row-- {
				temp[col] = append(temp[col], p[row][col])
			}
		}
		for row := 0; row < 9; row++ {
			p[row] = temp[row]
		}
	}
}

// Swap the digits digitA and digitB in the entire puzzle.
func (p sudokuPuzzle) swapDigits(digitA, digitB int) {
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			switch int(p[row][col]) {
			case digitA:
				p[row][col] = int8(digitB)
			case digitB:
				p[row][col] = int8(digitA)
			}
		}
	}
}

func sudokuPuzzleFromString(str string) sudokuPuzzle {
	out := make(sudokuPuzzle, 9)
	for row := 0; row < 9; row++ {
		out[row] = make([]int8, 9)
	}
	for idx, ch := range str {
		if idx >= 81 {
			return out
		}
		switch {
		case '0' <= ch && ch <= '9':
			out[idx/9][idx%9] = int8(byte(ch) - '0')
		default:
			out[idx/9][idx%9] = 0
		}

	}
	return out
}

var errTimeoutBreak = fmt.Errorf("timeout break")

func (p sudokuPuzzle) solveBruteForce(ctx context.Context, breakOn int) ([]sudokuPuzzle, error) {
	var solutions []sudokuPuzzle
	var recursion func(p sudokuPuzzle)
	timeoutBreak := false
	recursion = func(p sudokuPuzzle) {
		if timeoutBreak || (breakOn > 0 && breakOn <= len(solutions)) {
			return
		}
		select {
		case <-ctx.Done():
			timeoutBreak = true
			return
		default:
		}
		candidates := p.findCandidates()
		for row := 0; row < 9; row++ {
			for col := 0; col < 9; col++ {
				if p[row][col] > 0 {
					continue
				}
				point := app.Point{Row: row, Col: col}
				for _, c := range candidates.in(point) {
					p[row][col] = c
					recursion(p)
					p[row][col] = 0
				}
				return
			}
		}
		if !p.isCorrectSolve() {
			return
		}
		solutions = append(solutions, p.clone())
		return
	}

	puzzle := p.clone()
	recursion(puzzle)
	if timeoutBreak {
		return nil, errTimeoutBreak
	}
	return solutions, nil
}

func (p sudokuPuzzle) forEach(fn func(p app.Point, v int8, _break *bool), excludeCols ...int) {
	excludes := make(map[int]struct{})
	for _, e := range excludeCols {
		excludes[e] = struct{}{}
	}
	_break := false
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			if _, isExcluded := excludes[col]; isExcluded {
				continue
			}
			fn(app.Point{Row: row, Col: col}, p[row][col], &_break)
		}
	}
}

func (p sudokuPuzzle) forEachInRow(row int, fn func(p app.Point, v int8, _break *bool), excludeCols ...int) {
	excludes := make(map[int]struct{})
	for _, e := range excludeCols {
		excludes[e] = struct{}{}
	}
	_break := false
	for col := 0; !_break && col < 9; col++ {
		if _, isExcluded := excludes[col]; isExcluded {
			continue
		}
		fn(app.Point{Row: row, Col: col}, p[row][col], &_break)
	}
}

func (p sudokuPuzzle) forEachInCol(col int, fn func(p app.Point, v int8, _break *bool), excludeRows ...int) {
	excludes := make(map[int]struct{})
	for _, e := range excludeRows {
		excludes[e] = struct{}{}
	}
	_break := false
	for row := 0; !_break && row < 9; row++ {
		if _, isExcluded := excludes[row]; isExcluded {
			continue
		}
		fn(app.Point{Row: row, Col: col}, p[row][col], &_break)
	}
}

func (p sudokuPuzzle) forEachInBox(point app.Point, fn func(p app.Point, v int8, _break *bool), excludePoints ...app.Point) {
	excludes := make(map[app.Point]struct{})
	for _, e := range excludePoints {
		excludes[e] = struct{}{}
	}
	_break := false
	pBox := app.Point{Row: (point.Row / 3) * 3, Col: (point.Col / 3) * 3}
	for row := 0; !_break && row < 3; row++ {
		for col := 0; !_break && col < 3; col++ {
			pCurrent := app.Point{Row: pBox.Row + row, Col: pBox.Col + col}
			if _, isExcluded := excludes[pCurrent]; isExcluded {
				continue
			}
			fn(pCurrent, p[pCurrent.Row][pCurrent.Col], &_break)
		}
	}
}

func (p sudokuPuzzle) findCandidates() sudokuCandidates {
	c := newSudokuCandidates()
	p.forEach(func(p1 app.Point, v1 int8, _ *bool) {
		if v1 > 0 {
			return
		}
		for i := int8(1); i <= 9; i++ {
			c[p1.Row][p1.Col][i] = struct{}{}
		}
	})
	// delete from vertical and horizontal lines and from boxes 3x3
	p.forEach(func(p1 app.Point, v1 int8, _ *bool) {
		if v1 == 0 {
			return
		}
		boxRow, boxCol := p1.Row/3*3, p1.Col/3*3
		for i := 0; i < 9; i++ {
			delete(c[p1.Row][i], v1)
			delete(c[i][p1.Col], v1)
			delete(c[boxRow+i%3][boxCol+i/3], v1)
		}
	})
	return c
}

func (p sudokuPuzzle) isCorrectSolve() bool {
	isCorrect := true
	p.forEach(func(p1 app.Point, v1 int8, _break1 *bool) {
		if p[p1.Row][p1.Col] == 0 {
			isCorrect = false
			*_break1 = true
			return
		}
		p.forEachInRow(p1.Row, func(p2 app.Point, v2 int8, _break2 *bool) {
			if v1 == v2 {
				isCorrect = false
				*_break1, *_break2 = true, true
				return
			}
		}, p1.Col)
		p.forEachInCol(p1.Col, func(p2 app.Point, v2 int8, _break2 *bool) {
			if v1 == v2 {
				isCorrect = false
				*_break1, *_break2 = true, true
				return
			}
		}, p1.Row)
		p.forEachInBox(p1, func(p2 app.Point, v2 int8, _break2 *bool) {
			if v1 == v2 {
				isCorrect = false
				*_break1, *_break2 = true, true
				return
			}
		}, p1)
	})
	return isCorrect
}

// CountHints returns the number of hints in the current state.
func (p sudokuPuzzle) CountHints() (c int) {
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			if p[row][col] != 0 {
				c++
			}
		}
	}
	return
}

func (p sudokuPuzzle) clone() sudokuPuzzle {
	c := make([][]int8, len(p))
	for row := 0; row < len(p); row++ {
		c[row] = make([]int8, len(p[row]))
		copy(c[row], p[row])
	}
	return c
}

// Get the puzzle for user as a string of 81 characters, where the absence of a number is indicated by zero.
func (p sudokuPuzzle) String() string {
	return sudokuString(p)
}

func (p sudokuPuzzle) FindUserErrors() (listErrors []app.Point) {
	p.forEach(func(point1 app.Point, value1 int8, _ *bool) {
		if value1 == 0 {
			return
		}
		var point1Errors []app.Point
		findErrs := func(point2 app.Point, value2 int8, _ *bool) {
			if value1 == value2 {
				point1Errors = append(point1Errors, point2)
			}
		}
		p.forEachInRow(point1.Row, findErrs, point1.Col)
		p.forEachInCol(point1.Col, findErrs, point1.Row)
		p.forEachInBox(point1, findErrs, point1)
		if len(point1Errors) > 0 {
			point1Errors = append(point1Errors, point1)
			listErrors = append(listErrors, point1Errors...)
		}
	})
	return
}

//func (p sudokuPuzzle) FindErrors(target data.SudokuPuzzle) (listErrors []data.Point) {
//	p.forEach(func(point1 data.Point, value1 int8, _ *bool) {
//		if value1 == 0 {
//			return
//		}
//		userValue := target.In(point1)
//		if userValue == 0 {
//			return
//		}
//		if userValue != value1 {
//			findErrs := func(point2 data.Point, value2 int8, _ *bool) {
//				if value2 == userValue {
//					listErrors = append(listErrors, point2)
//				}
//			}
//			p.forEachInRow(point1.Row, findErrs, point1.Col)
//			p.forEachInCol(point1.Col, findErrs, point1.Row)
//			p.forEachInBox(point1, findErrs, point1)
//			listErrors = append(listErrors, point1)
//		}
//	})
//	return
//}

func (p sudokuPuzzle) In(point app.Point) int8 {
	return p[point.Row][point.Col]
}

func (p sudokuPuzzle) debug() string {
	return sudokuDebug(p)
}
