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

type puzzleCandidates [size][size]map[uint8]struct{}

func newPuzzleCandidates(fill bool) puzzleCandidates {
	var candidates puzzleCandidates
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			candidates[row][col] = make(map[uint8]struct{})
			if fill {
				for i := uint8(1); i <= size; i++ {
					candidates[row][col][i] = struct{}{}
				}
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
		candidates[point.Row][point.Col] = make(map[uint8]struct{})
		// delete from vertical and horizontal lines and from boxes 3x3
		rowBox, colBox := point.Row/sizeGrp*sizeGrp, point.Col/sizeGrp*sizeGrp
		for i := 0; i < size; i++ {
			delete(candidates[point.Row][i], val)
			delete(candidates[i][point.Col], val)
			delete(candidates[rowBox+i%sizeGrp][colBox+i/sizeGrp], val)
		}
	})
	return candidates
}

func (c puzzleCandidates) removeFrom(point app.Point, candidates []uint8) (removals []app.Point) {
	isRemoval := false
	for _, candidate := range candidates {
		if _, ok := c[point.Row][point.Col][candidate]; ok {
			isRemoval = true
			delete(c[point.Row][point.Col], candidate)
		}
	}
	if isRemoval {
		removals = append(removals, point)
	}
	return
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
	c.forEach(func(point1 app.Point, candidates1 map[uint8]struct{}, stop1 *bool) {
		if len(candidates1) != 2 {
			return
		}
		pairA := c.in(point1)
		c.forEachInRow(point1.Row, func(point2 app.Point, candidates2 map[uint8]struct{}, stop2 *bool) {
			if len(candidates2) != 2 {
				return
			}
			if !bytes.Equal(pairA, c.in(point2)) {
				return
			}
			c.forEachInRow(point1.Row, func(point3 app.Point, candidates3 map[uint8]struct{}, stop3 *bool) {
				r := c.removeFrom(point3, pairA)
				if len(r) > 0 {
					removals = append(removals, r...)
					changed = true
				}
			}, point1.Col, point2.Col)
			if changed {
				pairPoints = append(pairPoints, point1, point2)
				pair = pairA
				*stop1 = true
				*stop2 = true
			}
		}, point1.Col)
		if changed {
			return
		}
		c.forEachInCol(point1.Col, func(point2 app.Point, candidates2 map[uint8]struct{}, stop2 *bool) {
			if len(candidates2) != 2 {
				return
			}
			if !bytes.Equal(pairA, c.in(point2)) {
				return
			}
			c.forEachInCol(point1.Col, func(point3 app.Point, candidates3 map[uint8]struct{}, stop3 *bool) {
				r := c.removeFrom(point3, pairA)
				if len(r) > 0 {
					removals = append(removals, r...)
					changed = true
				}
			}, point1.Row, point2.Row)
			if changed {
				pairPoints = append(pairPoints, point1, point2)
				pair = pairA
				*stop1 = true
				*stop2 = true
			}
		}, point1.Row)
		if changed {
			return
		}
		c.forEachInBox(point1, func(point2 app.Point, candidates2 map[uint8]struct{}, stop2 *bool) {
			if len(candidates2) != 2 {
				return
			}
			if !bytes.Equal(pairA, c.in(point2)) {
				return
			}
			c.forEachInBox(point1, func(point3 app.Point, candidates3 map[uint8]struct{}, stop3 *bool) {
				r := c.removeFrom(point3, pairA)
				if len(r) > 0 {
					removals = append(removals, r...)
					changed = true
				}
			}, point1, point2)
			if changed {
				pairPoints = append(pairPoints, point1, point2)
				pair = pairA
				*stop1 = true
				*stop2 = true
			}
		}, point1)
		if changed {
			return
		}
	})
	return
}

func (c puzzleCandidates) String() string {
	s, _ := c.MarshalJSON()
	return string(s)
}

func (c puzzleCandidates) MarshalJSON() ([]byte, error) {
	keysFromCandidates := func(c map[uint8]struct{}) (out []int8) {
		for i := uint8(1); i <= size; i++ {
			if _, ok := c[i]; ok {
				out = append(out, int8(i))
			}
		}
		return
	}

	out := make(map[string][]int8)
	c.forEach(func(point app.Point, candidates map[uint8]struct{}, _ *bool) {
		if len(candidates) == 0 {
			return
		}
		out[point.String()] = keysFromCandidates(candidates)
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
			c[point.Row][point.Col][uint8(candidate)] = struct{}{}
		}
	}
	return nil
}

func (c puzzleCandidates) in(point app.Point) (candidates []uint8) {
	for candidate := range c[point.Row][point.Col] {
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i] < candidates[j]
	})
	return
}

func (c puzzleCandidates) forEach(fn func(point app.Point, candidates map[uint8]struct{}, stop *bool), excludePoints ...app.Point) {
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

func (c puzzleCandidates) forEachInRow(row int, fn func(point app.Point, candidates map[uint8]struct{}, stop *bool), excludeColumns ...int) {
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

func (c puzzleCandidates) forEachInCol(col int, fn func(point app.Point, candidates map[uint8]struct{}, stop *bool), excludeRows ...int) {
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

func (c puzzleCandidates) forEachInBox(point app.Point, fn func(point app.Point, candidates map[uint8]struct{}, stop *bool), excludePoints ...app.Point) {
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

func (c puzzleCandidates) debug(state *puzzle) string {
	is := func(candidates map[uint8]struct{}, d uint8) bool {
		_, ok := candidates[d]
		return ok
	}
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
						if digit := d + i; is(cell, digit) {
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
