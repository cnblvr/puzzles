package sudoku_classic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
	"sort"
	"strings"
)

type puzzleCandidates [size][size]cellCandidates

func newPuzzleCandidates(fill bool) puzzleCandidates {
	var candidates puzzleCandidates
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			candidates[row][col] = newCellCandidatesEmpty()
			if fill {
				candidates[row][col].fill()
			}
		}
	}
	return candidates
}

func (p puzzle) findSimpleCandidates() puzzleCandidates {
	candidates := newPuzzleCandidates(true)
	p.forEach(func(point app.Point, val uint8, _ *bool) {
		if val == 0 {
			return
		}
		candidates[point.Row][point.Col] = newCellCandidatesEmpty()
		// delete from vertical and horizontal lines and from boxes 3x3
		rowBox, colBox := point.Row/sizeGrp*sizeGrp, point.Col/sizeGrp*sizeGrp
		for i := 0; i < size; i++ {
			candidates[point.Row][i].delete(val)
			candidates[i][point.Col].delete(val)
			candidates[rowBox+i%sizeGrp][colBox+i/sizeGrp].delete(val)
		}
	})
	return candidates
}

func (c puzzleCandidates) simpleRemoveAfterSet(point app.Point, value uint8) (removals []app.Point) {
	rowBox, colBox := point.Row/sizeGrp*sizeGrp, point.Col/sizeGrp*sizeGrp
	for i := 0; i < size; i++ {
		if _, ok := c[point.Row][i][value]; i != point.Col && ok {
			removals = append(removals, app.Point{Row: point.Row, Col: i})
			delete(c[point.Row][i], value)
		}
		if _, ok := c[i][point.Col][value]; i != point.Row && ok {
			removals = append(removals, app.Point{Row: i, Col: point.Col})
			delete(c[i][point.Col], value)
		}
		boxPoint := app.Point{Row: rowBox + i%sizeGrp, Col: colBox + i/sizeGrp}
		if _, ok := c[boxPoint.Row][boxPoint.Col][value]; boxPoint != point && ok {
			removals = append(removals, boxPoint)
			delete(c[rowBox+i%sizeGrp][colBox+i/sizeGrp], value)
		}
	}
	for digit := uint8(1); digit <= size; digit++ {
		delete(c[point.Row][point.Col], digit)
	}
	return
}

func (c puzzleCandidates) strategyNakedPair() (pairPoints []app.Point, pair []uint8, removals []app.Point, changed bool) {
	c.forEach(func(point1 app.Point, candidates1 cellCandidates, stop1 *bool) {
		if candidates1.len() != 2 {
			return
		}
		pairA := candidates1.slice()
		c.forEachInRow(point1.Row, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			if candidates2.len() != 2 {
				return
			}
			if !bytes.Equal(pairA, candidates2.slice()) {
				return
			}
			c.forEachInRow(point1.Row, func(point3 app.Point, candidates3 cellCandidates, _ *bool) {
				if candidates3.delete(pairA...) {
					removals = append(removals, point3)
					changed = true
				}
			}, point1.Col, point2.Col)
			if changed {
				pairPoints = []app.Point{point1, point2}
				pair = pairA
				*stop1, *stop2 = true, true
			}
		}, point1.Col)
		if changed {
			return
		}
		c.forEachInCol(point1.Col, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			if candidates2.len() != 2 {
				return
			}
			if !bytes.Equal(pairA, candidates2.slice()) {
				return
			}
			c.forEachInCol(point1.Col, func(point3 app.Point, candidates3 cellCandidates, _ *bool) {
				if candidates3.delete(pairA...) {
					removals = append(removals, point3)
					changed = true
				}
			}, point1.Row, point2.Row)
			if changed {
				pairPoints = []app.Point{point1, point2}
				pair = pairA
				*stop1, *stop2 = true, true
			}
		}, point1.Row)
		if changed {
			return
		}
		c.forEachInBox(point1, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			if candidates2.len() != 2 {
				return
			}
			if !bytes.Equal(pairA, candidates2.slice()) {
				return
			}
			c.forEachInBox(point1, func(point3 app.Point, candidates3 cellCandidates, _ *bool) {
				if candidates3.delete(pairA...) {
					removals = append(removals, point3)
					changed = true
				}
			}, point1, point2)
			if changed {
				pairPoints = []app.Point{point1, point2}
				pair = pairA
				*stop1, *stop2 = true, true
			}
		}, point1)
		if changed {
			return
		}
	})
	return
}

func (c puzzleCandidates) strategyNakedTriple() (points []app.Point, triple []uint8, removals []app.Point, changed bool) {
	c.forEach(func(point1 app.Point, candidates1 cellCandidates, stop1 *bool) {
		if l := candidates1.len(); l < 2 || 3 < l {
			return
		}
		uniqueA := candidates1.clone() // TODO .union

		// watch row
		c.forEachInRow(point1.Row, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			if l := candidates2.len(); l < 2 || 3 < l {
				return
			}
			uniqueB := uniqueA.cloneWith(candidates2.slice()...)
			if uniqueB.len() > 3 {
				return
			}
			c.forEachInRow(point1.Row, func(point3 app.Point, candidates3 cellCandidates, stop3 *bool) {
				if l := candidates3.len(); l < 2 || 3 < l {
					return
				}
				uniqueC := uniqueB.cloneWith(candidates3.slice()...)
				if uniqueC.len() > 3 {
					return
				}
				// triple found
				c.forEachInRow(point1.Row, func(point4 app.Point, candidates4 cellCandidates, _ *bool) {
					if candidates4.delete(uniqueC.slice()...) {
						removals = append(removals, point4)
						changed = true
					}
				}, point1.Col, point2.Col, point3.Col)
				if changed {
					points = []app.Point{point1, point2, point3}
					triple = uniqueC.slice()
					*stop1, *stop2, *stop3 = true, true, true
				}
			}, point1.Col, point2.Col)
		}, point1.Col)
		if changed {
			return
		}

		// watch column
		c.forEachInCol(point1.Col, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			if l := candidates2.len(); l < 2 || 3 < l {
				return
			}
			uniqueB := uniqueA.cloneWith(candidates2.slice()...)
			if uniqueB.len() > 3 {
				return
			}
			c.forEachInCol(point1.Col, func(point3 app.Point, candidates3 cellCandidates, stop3 *bool) {
				if l := candidates3.len(); l < 2 || 3 < l {
					return
				}
				uniqueC := uniqueB.cloneWith(candidates3.slice()...)
				if uniqueC.len() > 3 {
					return
				}
				// triple found
				c.forEachInCol(point1.Col, func(point4 app.Point, candidates4 cellCandidates, _ *bool) {
					if candidates4.delete(uniqueC.slice()...) {
						removals = append(removals, point4)
						changed = true
					}
				}, point1.Row, point2.Row, point3.Row)
				if changed {
					points = []app.Point{point1, point2, point3}
					triple = uniqueC.slice()
					*stop1, *stop2, *stop3 = true, true, true
				}
			}, point1.Row, point2.Row)
		}, point1.Row)
		if changed {
			return
		}

		// watch box
		c.forEachInBox(point1, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			if l := candidates2.len(); l < 2 || 3 < l {
				return
			}
			uniqueB := uniqueA.cloneWith(candidates2.slice()...)
			if uniqueB.len() > 3 {
				return
			}
			c.forEachInBox(point1, func(point3 app.Point, candidates3 cellCandidates, stop3 *bool) {
				if l := candidates3.len(); l < 2 || 3 < l {
					return
				}
				uniqueC := uniqueB.cloneWith(candidates3.slice()...)
				if uniqueC.len() > 3 {
					return
				}
				// triple found
				c.forEachInBox(point1, func(point4 app.Point, candidates4 cellCandidates, _ *bool) {
					if candidates4.delete(uniqueC.slice()...) {
						removals = append(removals, point4)
						changed = true
					}
				}, point1, point2, point3)
				if changed {
					points = []app.Point{point1, point2, point3}
					triple = uniqueC.slice()
					*stop1, *stop2, *stop3 = true, true, true
				}
			}, point1, point2)
		}, point1)
		if changed {
			return
		}
	})
	return
}

func (c puzzleCandidates) strategyHiddenPair() (points []app.Point, pair []uint8, changed bool) {
	c.forEach(func(point1 app.Point, candidates1 cellCandidates, stop1 *bool) {

		// watch row
		c.forEachInRow(point1.Row, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			intersection2 := candidates1.intersection(candidates2)
			if intersection2.len() < 2 {
				return
			}
			complement := intersection2.clone()
			c.forEachInRow(point1.Row, func(_ app.Point, candidates3 cellCandidates, _ *bool) {
				complement = complement.complement(candidates3)
			}, point1.Col, point2.Col)
			if complement.len() != 2 {
				return
			}
			// pair found
			for _, candidates := range []cellCandidates{candidates1, candidates2} {
				if candidates.deleteExcept(complement.slice()...) {
					changed = true
				}
			}
			if changed {
				points = []app.Point{point1, point2}
				pair = complement.slice()
				*stop1, *stop2 = true, true
			}
		}, point1.Col)

		// watch column
		c.forEachInCol(point1.Col, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			intersection2 := candidates1.intersection(candidates2)
			if intersection2.len() < 2 {
				return
			}
			complement := intersection2.clone()
			c.forEachInCol(point1.Col, func(_ app.Point, candidates3 cellCandidates, _ *bool) {
				complement = complement.complement(candidates3)
			}, point1.Row, point2.Row)
			if complement.len() != 2 {
				return
			}
			// pair found
			for _, candidates := range []cellCandidates{candidates1, candidates2} {
				if candidates.deleteExcept(complement.slice()...) {
					changed = true
				}
			}
			if changed {
				points = []app.Point{point1, point2}
				pair = complement.slice()
				*stop1, *stop2 = true, true
			}
		}, point1.Row)

		// watch box
		c.forEachInBox(point1, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			intersection2 := candidates1.intersection(candidates2)
			if intersection2.len() < 2 {
				return
			}
			complement := intersection2.clone()
			c.forEachInBox(point1, func(_ app.Point, candidates3 cellCandidates, _ *bool) {
				complement = complement.complement(candidates3)
			}, point1, point2)
			if complement.len() != 2 {
				return
			}
			// pair found
			for _, candidates := range []cellCandidates{candidates1, candidates2} {
				if candidates.deleteExcept(complement.slice()...) {
					changed = true
				}
			}
			if changed {
				points = []app.Point{point1, point2}
				pair = complement.slice()
				*stop1, *stop2 = true, true
			}
		}, point1)
	})
	return
}

func (c puzzleCandidates) strategyHiddenTriple() (points []app.Point, triple []uint8, changed bool) {
	c.forEach(func(point1 app.Point, candidates1 cellCandidates, stop1 *bool) {

		// watch row
		c.forEachInRow(point1.Row, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			intersection12 := candidates1.intersection(candidates2)
			if l := intersection12.len(); l < 2 {
				return
			}
			c.forEachInRow(point1.Row, func(point3 app.Point, candidates3 cellCandidates, stop3 *bool) {
				intersection13 := candidates1.intersection(candidates3)
				if l := intersection13.len(); l < 2 {
					return
				}
				intersection23 := candidates2.intersection(candidates3)
				if l := intersection23.len(); l < 2 {
					return
				}
				complement := intersection12.union(intersection13).union(intersection23)
				c.forEachInRow(point1.Row, func(_ app.Point, candidates4 cellCandidates, _ *bool) {
					complement = complement.complement(candidates4)
				}, point1.Col, point2.Col, point3.Col)
				if complement.len() != 3 {
					return
				}
				// triple found
				for _, candidates := range []cellCandidates{candidates1, candidates2, candidates3} {
					if candidates.deleteExcept(complement.slice()...) {
						changed = true
					}
				}
				if changed {
					points = []app.Point{point1, point2, point3}
					triple = complement.slice()
					*stop1, *stop2, *stop3 = true, true, true
				}
			}, point1.Col, point2.Col)
		}, point1.Col)

		// watch column
		c.forEachInCol(point1.Col, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			intersection12 := candidates1.intersection(candidates2)
			if l := intersection12.len(); l < 2 {
				return
			}
			c.forEachInCol(point1.Col, func(point3 app.Point, candidates3 cellCandidates, stop3 *bool) {
				intersection13 := candidates1.intersection(candidates3)
				if l := intersection13.len(); l < 2 {
					return
				}
				intersection23 := candidates2.intersection(candidates3)
				if l := intersection23.len(); l < 2 {
					return
				}
				complement := intersection12.union(intersection13).union(intersection23)
				c.forEachInCol(point1.Col, func(_ app.Point, candidates4 cellCandidates, _ *bool) {
					complement = complement.complement(candidates4)
				}, point1.Row, point2.Row, point3.Row)
				if complement.len() != 3 {
					return
				}
				// triple found
				for _, candidates := range []cellCandidates{candidates1, candidates2, candidates3} {
					if candidates.deleteExcept(complement.slice()...) {
						changed = true
					}
				}
				if changed {
					points = []app.Point{point1, point2, point3}
					triple = complement.slice()
					*stop1, *stop2, *stop3 = true, true, true
				}
			}, point1.Row, point2.Row)
		}, point1.Row)

		// watch box
		c.forEachInBox(point1, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
			intersection12 := candidates1.intersection(candidates2)
			if l := intersection12.len(); l < 2 {
				return
			}
			c.forEachInBox(point1, func(point3 app.Point, candidates3 cellCandidates, stop3 *bool) {
				intersection13 := candidates1.intersection(candidates3)
				if l := intersection13.len(); l < 2 {
					return
				}
				intersection23 := candidates2.intersection(candidates3)
				if l := intersection23.len(); l < 2 {
					return
				}
				complement := intersection12.union(intersection13).union(intersection23)
				c.forEachInBox(point1, func(_ app.Point, candidates4 cellCandidates, _ *bool) {
					complement = complement.complement(candidates4)
				}, point1, point2, point3)
				if complement.len() != 3 {
					return
				}
				// triple found
				for _, candidates := range []cellCandidates{candidates1, candidates2, candidates3} {
					if candidates.deleteExcept(complement.slice()...) {
						changed = true
					}
				}
				if changed {
					points = []app.Point{point1, point2, point3}
					triple = complement.slice()
					*stop1, *stop2, *stop3 = true, true, true
				}
			}, point1, point2)
		}, point1)
	})
	return
}

// strategy Pair or Triple Box/Line Reduction
func (c puzzleCandidates) strategyBLRPairTriple() (points []app.Point, value uint8, removals []app.Point, changed bool) {
	c.forEachBox(func(pointBox1 app.Point, stop1 *bool) {
		for digit := uint8(1); digit <= size; digit++ {
			rows, cols := newCellCandidatesEmpty(), newCellCandidatesEmpty() // TODO is Set, not cellCandidates
			pointsDigit := make([]app.Point, 0, 3)
			c.forEachInBox(pointBox1, func(point2 app.Point, candidates2 cellCandidates, _ *bool) {
				if candidates2.has(digit) {
					rows.add(uint8(point2.Row))
					cols.add(uint8(point2.Col))
					pointsDigit = append(pointsDigit, point2)
				}
			})
			if l := len(pointsDigit); l < 2 || 3 < l {
				continue
			}
			if rows.len() == 1 {
				c.forEachInRow(int(rows.slice()[0]), func(point3 app.Point, candidates3 cellCandidates, _ *bool) {
					if candidates3.delete(digit) {
						removals = append(removals, point3)
						changed = true
					}
				}, cols.sliceInt()...)
			}
			if changed {
				value = digit
				points = pointsDigit
				*stop1 = true
				return
			}
			if cols.len() == 1 {
				c.forEachInCol(int(cols.slice()[0]), func(point3 app.Point, candidates3 cellCandidates, _ *bool) {
					if candidates3.delete(digit) {
						removals = append(removals, point3)
						changed = true
					}
				}, rows.sliceInt()...)
			}
			if changed {
				value = digit
				points = pointsDigit
				*stop1 = true
				return
			}
		}
	})
	return
}

func (c puzzleCandidates) in(point app.Point) []uint8 {
	return c[point.Row][point.Col].slice()
}

func (c puzzleCandidates) String() string {
	s, _ := c.MarshalJSON()
	return string(s)
}

func (c puzzleCandidates) MarshalJSON() ([]byte, error) {
	out := make(map[string][]int8)
	c.forEach(func(point app.Point, candidates cellCandidates, _ *bool) {
		if candidates.len() == 0 {
			return
		}
		out[point.String()] = candidates.sliceInt8()
	})
	return json.Marshal(out)
}

func (c *puzzleCandidates) UnmarshalJSON(bts []byte) error {
	in := make(map[string][]int8)
	if err := json.Unmarshal(bts, &in); err != nil {
		return err
	}
	*c = newPuzzleCandidates(false)
	for pointStr, candidates := range in {
		point, err := app.PointFromString(pointStr)
		if err != nil {
			return errors.Wrapf(err, "point '%s' invalid", pointStr)
		}
		for _, candidate := range candidates {
			c[point.Row][point.Col].add(uint8(candidate))
		}
	}
	return nil
}

func (c puzzleCandidates) forEach(fn func(point app.Point, candidates cellCandidates, stop *bool), excludePoints ...app.Point) {
	excludes := make(map[app.Point]struct{})
	for _, point := range excludePoints {
		excludes[point] = struct{}{}
	}
	stop := false
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if stop {
				return
			}
			point := app.Point{Row: row, Col: col}
			if _, ok := excludes[point]; ok {
				continue
			}
			fn(point, c[row][col], &stop)
		}
	}
}

func (c puzzleCandidates) forEachBox(fn func(pointBox app.Point, stop *bool), excludeBoxes ...app.Point) {
	excludes := make(map[app.Point]struct{})
	for _, pointGiven := range excludeBoxes {
		pointBox := app.Point{Row: (pointGiven.Row / sizeGrp) * sizeGrp, Col: (pointGiven.Col / sizeGrp) * sizeGrp}
		excludes[pointBox] = struct{}{}
	}
	stop := false
	for row := 0; row < size; row += sizeGrp {
		for col := 0; col < size; col += sizeGrp {
			if stop {
				return
			}
			point := app.Point{Row: row, Col: col}
			if _, ok := excludes[point]; ok {
				continue
			}
			fn(point, &stop)
		}
	}
}

func (c puzzleCandidates) forEachInRow(row int, fn func(point app.Point, candidates cellCandidates, stop *bool), excludeColumns ...int) {
	excludes := make(map[int]struct{})
	for _, col := range excludeColumns {
		excludes[col] = struct{}{}
	}
	stop := false
	for col := 0; col < size; col++ {
		if stop {
			return
		}
		if _, ok := excludes[col]; ok {
			continue
		}
		fn(app.Point{Row: row, Col: col}, c[row][col], &stop)
	}
}

func (c puzzleCandidates) forEachInCol(col int, fn func(point app.Point, candidates cellCandidates, stop *bool), excludeRows ...int) {
	excludes := make(map[int]struct{})
	for _, row := range excludeRows {
		excludes[row] = struct{}{}
	}
	stop := false
	for row := 0; row < size; row++ {
		if stop {
			return
		}
		if _, ok := excludes[row]; ok {
			continue
		}
		fn(app.Point{Row: row, Col: col}, c[row][col], &stop)
	}
}

func (c puzzleCandidates) forEachInBox(point app.Point, fn func(point app.Point, candidates cellCandidates, stop *bool), excludePoints ...app.Point) {
	excludes := make(map[app.Point]struct{})
	for _, point := range excludePoints {
		excludes[point] = struct{}{}
	}
	stop := false
	pBox := app.Point{Row: (point.Row / sizeGrp) * sizeGrp, Col: (point.Col / sizeGrp) * sizeGrp}
	for row := pBox.Row; row < pBox.Row+sizeGrp; row++ {
		for col := pBox.Col; col < pBox.Col+sizeGrp; col++ {
			if stop {
				return
			}
			point := app.Point{Row: row, Col: col}
			if _, ok := excludes[point]; ok {
				continue
			}
			fn(point, c[row][col], &stop)
		}
	}
}

type cellCandidates map[uint8]struct{}

func newCellCandidatesEmpty() cellCandidates {
	return make(cellCandidates)
}

func newCellCandidatesFilled() cellCandidates {
	c := newCellCandidatesEmpty()
	c.fill()
	return c
}

func newCellCandidatesWith(digits ...uint8) cellCandidates {
	c := newCellCandidatesEmpty()
	c.add(digits...)
	return c
}

func (c cellCandidates) len() int {
	return len(c)
}

func (c cellCandidates) slice() (candidates []uint8) {
	for candidate := range c {
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i] < candidates[j]
	})
	return
}

func (c cellCandidates) sliceInt8() []int8 {
	u8 := c.slice()
	out := make([]int8, len(u8))
	for idx, u := range u8 {
		out[idx] = int8(u)
	}
	return out
}

func (c cellCandidates) sliceInt() []int {
	u8 := c.slice()
	out := make([]int, len(u8))
	for idx, u := range u8 {
		out[idx] = int(u)
	}
	return out
}

func (c cellCandidates) has(value uint8) bool {
	_, ok := c[value]
	return ok
}

func (c cellCandidates) delete(digits ...uint8) bool {
	out := false
	for _, digit := range digits {
		if c.has(digit) {
			delete(c, digit)
			out = true
		}
	}
	return out
}

func (c cellCandidates) deleteExcept(digits ...uint8) bool {
	forDelete := c.complement(newCellCandidatesWith(digits...))
	return c.delete(forDelete.slice()...)
}

func (c cellCandidates) add(digits ...uint8) {
	for _, digit := range digits {
		c[digit] = struct{}{}
	}
}

func (c cellCandidates) fill() {
	for i := uint8(1); i <= size; i++ {
		c[i] = struct{}{}
	}
}

func (c cellCandidates) clone() cellCandidates {
	return newCellCandidatesWith(c.slice()...)
}

func (c cellCandidates) cloneWith(digits ...uint8) cellCandidates {
	clone := newCellCandidatesEmpty()
	clone.add(c.slice()...)
	clone.add(digits...)
	return clone
}

// c ⋂ with
func (c cellCandidates) intersection(with cellCandidates) cellCandidates {
	intersection := newCellCandidatesEmpty()
	for candidate := range c {
		if with.has(candidate) {
			intersection.add(candidate)
		}
	}
	return intersection
}

// c ⋃ with
func (c cellCandidates) union(with cellCandidates) cellCandidates {
	return c.cloneWith(with.slice()...)
}

// c \ of
func (c cellCandidates) complement(of cellCandidates) cellCandidates {
	complement := newCellCandidatesEmpty()
	for candidate := range c {
		if !of.has(candidate) {
			complement.add(candidate)
		}
	}
	return complement
}

func (c puzzleCandidates) debug(state *puzzle) string {
	var out strings.Builder
	out.WriteString("╔═══════╤═══════╤═══════╦═══════╤═══════╤═══════╦═══════╤═══════╤═══════╗  \n")
	for row := 0; row < size; row++ {
		for d := uint8(1); d <= size; d += sizeGrp {
			out.WriteString("║ ")
			for col := 0; col < size; col++ {
				cell := c[row][col]
				if state != nil && state[row][col] > 0 {
					clue := state[row][col]
					switch {
					case d == 1:
						out.WriteString("      ")
					case d == 4:
						out.WriteString(fmt.Sprintf(" (%s)  ", string(clue+'0')))
					case d == 7:
						out.WriteString("      ")
					}
				} else {
					for i := uint8(0); i < sizeGrp; i++ {
						if digit := d + i; cell.has(digit) {
							out.WriteString(fmt.Sprintf("%d ", digit))
						} else {
							out.WriteString("  ")
						}
					}
				}
				if col%sizeGrp == sizeGrp-1 {
					if col != size-1 {
						out.WriteString("║ ")
					}
				} else {
					out.WriteString("│ ")
				}
			}
			if d == 4 {
				out.WriteString(fmt.Sprintf("║ %s\n", string(byte(row)+'a')))
			} else {
				out.WriteString("║  \n")
			}
		}
		if row%sizeGrp == sizeGrp-1 {
			if row != size-1 {
				out.WriteString("╠═══════╪═══════╪═══════╬═══════╪═══════╪═══════╬═══════╪═══════╪═══════╣  \n")
			}
		} else {
			out.WriteString("╟───────┼───────┼───────╫───────┼───────┼───────╫───────┼───────┼───────╢  \n")
		}
	}
	out.WriteString("╚═══════╧═══════╧═══════╩═══════╧═══════╧═══════╩═══════╧═══════╧═══════╝  \n")
	out.WriteString("    1       2       3       4       5       6       7       8       9      \n")
	return out.String()
}
