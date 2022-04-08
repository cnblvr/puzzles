package sudoku_classic

import (
	"encoding/json"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
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

func (p puzzle) findCandidates() puzzleCandidates {
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
