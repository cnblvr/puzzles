package sudoku_classic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
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

func (c puzzleCandidates) clone() puzzleCandidates {
	clone := newPuzzleCandidates(false)
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			clone[row][col].add(c[row][col].slice()...)
		}
	}
	return clone
}

type puzzleCandidatesExternal struct {
	Base   map[string][]int8 `json:"base,omitempty"`
	Add    map[string][]int8 `json:"add,omitempty"`
	Delete map[string][]int8 `json:"del,omitempty"`
}

func (c puzzleCandidates) encode() string {
	out := puzzleCandidatesExternal{
		Base: make(map[string][]int8),
	}
	c.forEach(func(point app.Point, candidates cellCandidates, _ *bool) {
		if candidates.len() == 0 {
			return
		}
		out.Base[point.String()] = candidates.sliceInt8()
	})
	bts, err := json.Marshal(out)
	if err != nil {
		zlog.Warn().Err(err).Msg("failed to encode puzzleCandidates")
	}
	return string(bts)
}

func (c puzzleCandidates) encodeOnlyChanges(base puzzleCandidates) string {
	out := puzzleCandidatesExternal{
		Add:    make(map[string][]int8),
		Delete: make(map[string][]int8),
	}
	c.forEach(func(point app.Point, candidates cellCandidates, _ *bool) {
		del := base[point.Row][point.Col].complement(candidates)
		if del.len() > 0 {
			out.Delete[point.String()] = del.sliceInt8()
		}
		add := candidates.complement(base[point.Row][point.Col])
		if add.len() > 0 {
			out.Add[point.String()] = add.sliceInt8()
		}
	})
	if len(out.Add) == 0 {
		out.Add = nil
	}
	if len(out.Delete) == 0 {
		out.Delete = nil
	}
	bts, err := json.Marshal(out)
	if err != nil {
		zlog.Warn().Err(err).Msg("failed to encodeOnlyChanges puzzleCandidates")
	}
	return string(bts)
}

func decodeCandidates(s string) (puzzleCandidates, error) {
	in := puzzleCandidatesExternal{}
	if err := json.Unmarshal([]byte(s), &in); err != nil {
		return puzzleCandidates{}, errors.Wrap(err, "decode candidates error")
	}
	c := newPuzzleCandidates(false)
	for pointStr, candidates := range in.Base {
		point, err := app.PointFromString(pointStr)
		if err != nil {
			return puzzleCandidates{}, errors.Wrapf(err, "decode candidates error: point '%s'", pointStr)
		}
		if point.Row >= size || point.Col >= size {
			return puzzleCandidates{}, errors.Errorf("decode candidates error: wrong point format '%s'", pointStr)
		}
		for _, candidate := range candidates {
			if candidate < 1 || size < candidate {
				return puzzleCandidates{}, errors.Errorf("decode candidates error: wrong candidate '%d'", candidate)
			}
		}
		c[point.Row][point.Col].addInt8(candidates...)
	}
	return c, nil
}

func (p puzzle) GetCandidates() string {
	return p.findSimpleCandidates().encode()
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

func (p puzzle) optimizeCandidates(c *puzzleCandidates) {
	p.forEach(func(point app.Point, val uint8, _ *bool) {
		if val > 0 {
			c[point.Row][point.Col] = newCellCandidatesEmpty()
		}
	})
}

func (c puzzleCandidates) simpleRemoveAfterSet(point app.Point, value uint8) {
	rowBox, colBox := point.Row/sizeGrp*sizeGrp, point.Col/sizeGrp*sizeGrp
	for i := 0; i < size; i++ {
		if _, ok := c[point.Row][i][value]; i != point.Col && ok {
			delete(c[point.Row][i], value)
		}
		if _, ok := c[i][point.Col][value]; i != point.Row && ok {
			delete(c[i][point.Col], value)
		}
		boxPoint := app.Point{Row: rowBox + i%sizeGrp, Col: colBox + i/sizeGrp}
		if _, ok := c[boxPoint.Row][boxPoint.Col][value]; boxPoint != point && ok {
			delete(c[rowBox+i%sizeGrp][colBox+i/sizeGrp], value)
		}
	}
	for digit := uint8(1); digit <= size; digit++ {
		delete(c[point.Row][point.Col], digit)
	}
	return
}

func (c puzzleCandidates) strategyNakedPair() (pairPoints []app.Point, pair []uint8, changed bool) {
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

func (c puzzleCandidates) strategyNakedTriple() (points []app.Point, triple []uint8, changed bool) {
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

// strategy Pointing Pair or Triple
func (c puzzleCandidates) strategyPointingPairTriple() (points []app.Point, value uint8, changed bool) {
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

func (c puzzleCandidates) strategyBoxLineReductionPairTriple() (points []app.Point, value uint8, changed bool) {
	for row := 0; row < size; row++ {
		for digit := uint8(1); digit <= size; digit++ {
			boxes := newCellCandidatesEmpty()
			pointsDigit := make([]app.Point, 0, 3)
			c.forEachInRow(row, func(point2 app.Point, candidates2 cellCandidates, _ *bool) {
				if candidates2.has(digit) {
					boxes.add(BoxIdFrom(point2))
					pointsDigit = append(pointsDigit, point2)
				}
			})
			if boxes.len() != 1 {
				continue
			}
			// candidates in one box and in row found
			c.forEachInBox(pointsDigit[0], func(point3 app.Point, candidates3 cellCandidates, _ *bool) {
				if candidates3.delete(digit) {
					changed = true
				}
			}, pointsDigit...)
			if changed {
				points = pointsDigit
				value = digit
				return
			}
		}
	}
	for col := 0; col < size; col++ {
		for digit := uint8(1); digit <= size; digit++ {
			boxes := newCellCandidatesEmpty()
			pointsDigit := make([]app.Point, 0, 3)
			c.forEachInCol(col, func(point2 app.Point, candidates2 cellCandidates, _ *bool) {
				if candidates2.has(digit) {
					boxes.add(BoxIdFrom(point2))
					pointsDigit = append(pointsDigit, point2)
				}
			})
			if boxes.len() != 1 {
				continue
			}
			// candidates in one box and in column found
			c.forEachInBox(pointsDigit[0], func(point3 app.Point, candidates3 cellCandidates, _ *bool) {
				if candidates3.delete(digit) {
					changed = true
				}
			}, pointsDigit...)
			if changed {
				points = pointsDigit
				value = digit
				return
			}
		}
	}
	return
}

func (c puzzleCandidates) strategyXWing() (pairA, pairB []app.Point, value uint8, changed bool) {
	for digit := uint8(1); digit <= size; digit++ {
		type tPair struct {
			unit int
			pair []app.Point
		}
		var pairsOnRow []tPair
		for row := 0; row < size; row++ {
			rowPoints := make([]app.Point, 0)
			c.forEachInRow(row, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
				if candidates2.has(digit) {
					rowPoints = append(rowPoints, point2)
				}
				if len(rowPoints) > 2 {
					*stop2 = true
				}
			})
			if l := len(rowPoints); l != 2 {
				continue
			}
			pairsOnRow = append(pairsOnRow, tPair{
				unit: row,
				pair: rowPoints,
			})
		}
		if len(pairsOnRow) < 2 {
			continue
		}
		for a := 0; a < len(pairsOnRow); a++ {
			for b := a + 1; b < len(pairsOnRow); b++ {
				col1, col2 := -1, -1
				if pairsOnRow[a].pair[0].Col == pairsOnRow[b].pair[0].Col {
					col1 = pairsOnRow[a].pair[0].Col
				}
				if pairsOnRow[a].pair[1].Col == pairsOnRow[b].pair[1].Col {
					col2 = pairsOnRow[a].pair[1].Col
				}
				if col1 == -1 || col2 == -1 {
					continue
				}
				// pairs in row found
				removeCandidates := func(point2 app.Point, candidates2 cellCandidates, _ *bool) {
					if candidates2.delete(digit) {
						changed = true
					}
				}
				c.forEachInCol(col1, removeCandidates, pairsOnRow[a].unit, pairsOnRow[b].unit)
				c.forEachInCol(col2, removeCandidates, pairsOnRow[a].unit, pairsOnRow[b].unit)
				if changed {
					value = digit
					pairA, pairB = pairsOnRow[a].pair, pairsOnRow[b].pair
					return
				}
			}
		}

		// search in columns

		var pairsOnCol []tPair
		for col := 0; col < size; col++ {
			colPoints := make([]app.Point, 0)
			c.forEachInCol(col, func(point2 app.Point, candidates2 cellCandidates, stop2 *bool) {
				if candidates2.has(digit) {
					colPoints = append(colPoints, point2)
				}
				if len(colPoints) > 2 {
					*stop2 = true
				}
			})
			if l := len(colPoints); l != 2 {
				continue
			}
			pairsOnCol = append(pairsOnCol, tPair{
				unit: col,
				pair: colPoints,
			})
		}
		if len(pairsOnCol) < 2 {
			continue
		}
		for a := 0; a < len(pairsOnCol); a++ {
			for b := a + 1; b < len(pairsOnCol); b++ {
				row1, row2 := -1, -1
				if pairsOnCol[a].pair[0].Row == pairsOnCol[b].pair[0].Row {
					row1 = pairsOnCol[a].pair[0].Row
				}
				if pairsOnCol[a].pair[1].Row == pairsOnCol[b].pair[1].Row {
					row2 = pairsOnCol[a].pair[1].Row
				}
				if row1 == -1 || row2 == -1 {
					continue
				}
				// pairs in column found
				removeCandidates := func(point2 app.Point, candidates2 cellCandidates, _ *bool) {
					if candidates2.delete(digit) {
						changed = true
					}
				}
				c.forEachInRow(row1, removeCandidates, pairsOnCol[a].unit, pairsOnCol[b].unit)
				c.forEachInRow(row2, removeCandidates, pairsOnCol[a].unit, pairsOnCol[b].unit)
				if changed {
					value = digit
					pairA, pairB = pairsOnCol[a].pair, pairsOnCol[b].pair
					return
				}
			}
		}
	}
	return
}

func (p puzzle) GetWrongCandidates(candidates string) (string, error) {
	c, err := decodeCandidates(candidates)
	if err != nil {
		return "", errors.WithStack(err)
	}
	wrongs := p.getWrongCandidates(c)
	return wrongs.encode(), nil
}

func (p puzzle) getWrongCandidates(c puzzleCandidates) puzzleCandidates {
	wrongs := newPuzzleCandidates(false)
	c.forEach(func(point1 app.Point, candidates1 cellCandidates, _ *bool) {
		if p[point1.Row][point1.Col] > 0 {
			return
		}
		for candidate := range candidates1 {
			findErrs := func(_ app.Point, value2 uint8, stop2 *bool) {
				if candidate == value2 {
					wrongs[point1.Row][point1.Col].add(candidate)
					*stop2 = true
				}
			}
			p.forEachInRow(point1.Row, findErrs, point1.Col)
			p.forEachInCol(point1.Col, findErrs, point1.Row)
			p.forEachInBox(point1, findErrs, point1)
		}
	})
	return wrongs
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

func (c cellCandidates) addInt8(digits ...int8) {
	for _, digit := range digits {
		c[uint8(digit)] = struct{}{}
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

// BoxIdFrom returns 1, 2, 3, 4, 5, 6, 7, 8 or 9 as box 3x3 id.
func BoxIdFrom(point app.Point) uint8 {
	return uint8(point.Row/sizeGrp*sizeGrp + point.Col/sizeGrp + 1)
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
