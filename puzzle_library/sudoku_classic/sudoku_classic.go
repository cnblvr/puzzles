package sudoku_classic

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
	"math/rand"
	"strconv"
	"strings"
)

const (
	// size is the width and height measurement
	size = 9
	// sizeGrp is the number of rows or columns for the big line
	sizeGrp = 3
)

type puzzle [size][size]uint8

func parse(str string) (*puzzle, error) {
	if len(str) != size*size {
		return nil, errors.Errorf("invalid puzzle length: %d", len(str))
	}
	var p puzzle
	for i := 0; i < size*size; i++ {
		if '1' <= str[i] && str[i] <= '9' {
			p[i/size][i%size] = str[i] - '0'
		}
	}
	return &p, nil
}

// ParseAssistant parses str into an interface that can be used to work with the
// generated puzzle or user state of the puzzle.
func ParseAssistant(str string) (app.PuzzleAssistant, error) {
	return parse(str)
}

// ParseGenerator parses str into an interface that can be used to generate the
// puzzle.
func ParseGenerator(str string) (app.PuzzleGenerator, error) {
	return parse(str)
}

func (p puzzle) String() string {
	out := make([]byte, size*size)
	for i := 0; i < size*size; i++ {
		char := p[i/size][i%size]
		if char > 0 {
			out[i] = char + '0'
		} else {
			out[i] = '.'
		}
	}
	return string(out)
}

func (p puzzle) Type() app.PuzzleType {
	return app.PuzzleSudokuClassic
}

// NewRandomSolution generates a solution randomly for further extraction of
// digits.
func NewRandomSolution() (s app.PuzzleGenerator, seed int64) {
	seedBts := make([]byte, 8)
	if _, err := crand.Reader.Read(seedBts); err != nil {
		panic(err)
	}
	seed = int64(binary.LittleEndian.Uint64(seedBts))
	return NewSolutionBySeed(seed), seed
}

// NewSolutionBySeed generates a solution with a given seed for further
// extraction of digits.
func NewSolutionBySeed(seed int64) app.PuzzleGenerator {
	rnd := rand.New(rand.NewSource(seed))

	s := generateWithoutShuffling(rnd)

	return &s
}

func generateWithoutShuffling(rnd *rand.Rand) (s puzzle) {
	digits := []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9}
	// Generate first line randomly
	i := 0
	for len(digits) > 0 {
		idx := rnd.Int() % len(digits)
		s[0][i] = digits[idx]
		i++
		digits = append(digits[:idx], digits[idx+1:]...)
	}

	// The second and third lines is the offset of the previous line to the left by 3
	// The next "big" lines (d-f and g-i) are generated like this:
	//  lines d and g:    offset of the previous line to the left by 1
	//  lines e, f, h, i: offset of the previous line to the left by 3
	for l := 1; l < size; l++ {
		if l%sizeGrp == 0 {
			copy(s[l][:size-1], s[l-1][1:size])
			s[l][size-1] = s[l-1][0]
			continue
		}
		copy(s[l][:size-sizeGrp], s[l-1][sizeGrp:size])
		copy(s[l][size-sizeGrp:size], s[l-1][:sizeGrp])
	}

	return
}

func (p *puzzle) SwapLines(dir app.DirectionType, a, b int) error {
	switch dir {
	case app.Horizontal, app.Vertical:
	default:
		return errors.Errorf("dir unknown: %d", dir)
	}
	if 0 > a || a > size-1 {
		return errors.Errorf("a is incorrect line: %d", a)
	}
	if 0 > b || b > size-1 {
		return errors.Errorf("b is incorrect line: %d", b)
	}
	if a == b {
		return errors.Errorf("a == b == %d", a)
	}
	for i := 0; i < size; i++ {
		if dir == app.Horizontal {
			p[a][i], p[b][i] = p[b][i], p[a][i]
		} else {
			p[i][a], p[i][b] = p[i][b], p[i][a]
		}
	}
	return nil
}

func (p *puzzle) SwapBigLines(dir app.DirectionType, a, b int) error {
	switch dir {
	case app.Horizontal, app.Vertical:
	default:
		return errors.Errorf("dir unknown: %d", dir)
	}
	if 0 > a || a > sizeGrp-1 {
		return errors.Errorf("a is incorrect line: %d", a)
	}
	if 0 > b || b > sizeGrp-1 {
		return errors.Errorf("b is incorrect line: %d", b)
	}
	if a == b {
		return errors.Errorf("a == b == %d", a)
	}
	for l := 0; l < sizeGrp; l++ {
		for i := 0; i < size; i++ {
			la, lb := a*sizeGrp+l, b*sizeGrp+l
			if dir == app.Horizontal {
				p[la][i], p[lb][i] = p[lb][i], p[la][i]
			} else {
				p[i][la], p[i][lb] = p[i][lb], p[i][la]
			}
		}
	}
	return nil
}

func (p *puzzle) Rotate(r app.RotationType) error {
	// A 0-degree angle rotation has no effect.
	// The input argument r is the ring [0; 3].
	r = r % (app.RotateTo270 + 1)

	// The rotation of the board of this type of Sudoku to any angle is done by vertical, horizontal reflections and
	// reflection along the major diagonal. So this method is useless.

	var err1, err2 error
	switch r {
	case app.RotateTo90:
		err1 = p.Reflect(app.ReflectMajorDiagonal)
		err2 = p.Reflect(app.ReflectVertical)

	case app.RotateTo180:
		err1 = p.Reflect(app.ReflectVertical)
		err2 = p.Reflect(app.ReflectHorizontal)

	case app.RotateTo270:
		err1 = p.Reflect(app.ReflectMajorDiagonal)
		err2 = p.Reflect(app.ReflectHorizontal)
	}
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func (p *puzzle) Reflect(r app.ReflectionType) error {
	switch r {
	case app.ReflectHorizontal:
		for row := 0; row < size; row++ {
			for i := 0; i < size/2; i++ {
				p[row][i], p[row][size-1-i] = p[row][size-1-i], p[row][i]
			}
		}

	case app.ReflectVertical:
		for col := 0; col < size; col++ {
			for i := 0; i < size/2; i++ {
				p[i][col], p[size-1-i][col] = p[size-1-i][col], p[i][col]
			}
		}

	case app.ReflectMajorDiagonal:
		for diag := 0; diag < size; diag++ {
			for i := diag + 1; i < size; i++ {
				p[diag][i], p[i][diag] = p[i][diag], p[diag][i]
			}
		}

	case app.ReflectMinorDiagonal:
		for diag := 0; diag < size; diag++ {
			for i := 0; i < size-1-diag; i++ {
				p[diag][i], p[size-1-i][size-1-diag] = p[size-1-i][size-1-diag], p[diag][i]
			}
		}
	default:
		return errors.Errorf("reflection type unknown: %d", r)
	}
	return nil
}

func (p *puzzle) SwapDigits(a, b uint8) error {
	if 1 > a || a > size {
		return errors.Errorf("a is incorrect digit: %d", a)
	}
	if 1 > b || b > size {
		return errors.Errorf("b is incorrect digit: %d", b)
	}
	if a == b {
		return errors.Errorf("a == b == %d", a)
	}
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			switch p[row][col] {
			case a:
				p[row][col] = b
			case b:
				p[row][col] = a
			}
		}
	}
	return nil
}

func (p *puzzle) Solve(candidatesIn string, chanSteps chan<- app.PuzzleStep) (changed bool, candidatesOut string, err error) {
	var candidates puzzleCandidates
	if candidatesIn == "" {
		candidates = p.findSimpleCandidates()
	} else {
		candidates, err = decodeCandidates(candidatesIn)
		if err != nil {
			return
		}
		p.optimizeCandidates(candidates)
	}
	defer func(candidates puzzleCandidates) {
		candidatesOut = candidates.encode()
	}(candidates)

	changedOnIteration := true
	if chanSteps != nil {
		defer close(chanSteps)
	}
	for changedOnIteration {
		changedOnIteration = false

		var step app.PuzzleStep
		changedOnIteration, candidates, step, err = p.solveOneStep(candidates)
		if err != nil {
			return
		}
		if changedOnIteration {
			changed = true
			if chanSteps != nil {
				chanSteps <- step
			}
		}
	}
	return
}

func (p *puzzle) SolveOneStep(candidatesIn string) (candidatesOut string, step app.PuzzleStep, err error) {
	var candidates puzzleCandidates
	if candidatesIn == "" {
		candidates = p.findSimpleCandidates()
	} else {
		candidates, err = decodeCandidates(candidatesIn)
		if err != nil {
			return
		}
		p.optimizeCandidates(candidates)
	}
	defer func(candidates puzzleCandidates) {
		candidatesOut = candidates.encode()
	}(candidates)

	_, candidates, step, err = p.solveOneStep(candidates)
	if err != nil {
		return
	}
	return
}

func (p *puzzle) solveOneStep(candidates puzzleCandidates) (changed bool, candidatesOut puzzleCandidates, step app.PuzzleStep, err error) {
	makeStep := func(s app.PuzzleStep) {
		switch s := s.(type) {
		case puzzleStepSet:
			p[s.point.Row][s.point.Col] = s.value
			s.removalsCandidates = candidates.simpleRemoveAfterSet(s.point, s.value)
		}
		step = s
		changed = true
	}
	defer func() {
		candidatesOut = candidates
	}()

	// strategy Naked Single
	p.forEach(func(point1 app.Point, val1 uint8, stop1 *bool) {
		if val1 > 0 {
			return
		}
		candidates1 := candidates[point1.Row][point1.Col]
		var candidate uint8
		switch count := candidates1.len(); {
		case count > 1:
			return
		case count == 0:
			err = errors.Errorf("candidates in %s is emtpy", point1.String())
			*stop1 = true
			return
		case count == 1:
			candidate = candidates1.slice()[0]
		}
		makeStep(puzzleStepSet{
			strategy: app.StrategyNakedSingle,
			point:    point1,
			value:    candidate,
		})
		*stop1 = true
		return
	})
	if changed || err != nil {
		return
	}

	// strategy Hidden Single
	p.forEach(func(point1 app.Point, val1 uint8, stop1 *bool) {
		if val1 > 0 {
			return
		}
		for _, candidate := range candidates.in(point1) {
			isHiddenSingle := uint8(0b111)
			candidates.forEachInRow(point1.Row, func(_ app.Point, candidates2 cellCandidates, stop2 *bool) {
				if candidates2.has(candidate) {
					isHiddenSingle &= 0b011
					*stop2 = true
				}
			}, point1.Col)
			candidates.forEachInCol(point1.Col, func(_ app.Point, candidates2 cellCandidates, stop2 *bool) {
				if candidates2.has(candidate) {
					isHiddenSingle &= 0b101
					*stop2 = true
				}
			}, point1.Row)
			candidates.forEachInBox(point1, func(_ app.Point, candidates2 cellCandidates, stop2 *bool) {
				if candidates2.has(candidate) {
					isHiddenSingle &= 0b110
					*stop2 = true
				}
			}, point1)
			if isHiddenSingle == 0 {
				continue
			}
			makeStep(puzzleStepSet{
				strategy: app.StrategyHiddenSingle,
				point:    point1,
				value:    candidate,
			})
			*stop1 = true
			return
		}
	})
	if changed {
		return
	}

	// strategy Naked Pair
	if points, pair, removals, ok := candidates.strategyNakedPair(); ok {
		makeStep(puzzleStepNakedStrategy{
			points:             points,
			set:                pair,
			removalsCandidates: removals,
		})
		return
	}
	// strategy Naked Triple
	if points, triple, removals, ok := candidates.strategyNakedTriple(); ok {
		makeStep(puzzleStepNakedStrategy{
			points:             points,
			set:                triple,
			removalsCandidates: removals,
		})
		return
	}
	// strategy Hidden Pair
	if points, pair, ok := candidates.strategyHiddenPair(); ok {
		makeStep(puzzleStepHiddenStrategy{
			points: points,
			set:    pair,
		})
		return
	}
	// strategy Hidden Triple
	if points, triple, ok := candidates.strategyHiddenTriple(); ok {
		makeStep(puzzleStepHiddenStrategy{
			points: points,
			set:    triple,
		})
		return
	}
	// strategy Pointing Pair or Triple
	if points, value, removals, ok := candidates.strategyPointingPairTriple(); ok {
		makeStep(puzzleStepPointingStrategy{
			points:             points,
			value:              value,
			removalsCandidates: removals,
		})
		return
	}
	// strategy Box Line Reduction Pair or Triple
	if points, value, removals, ok := candidates.strategyBoxLineReductionPairTriple(); ok {
		makeStep(puzzleStepBoxLineReductionStrategy{
			points:             points,
			value:              value,
			removalsCandidates: removals,
		})
		return
	}
	// strategy X-Wing
	if pairA, pairB, value, removals, ok := candidates.strategyXWing(); ok {
		makeStep(puzzleStepXWingStrategy{
			pairA:              pairA,
			pairB:              pairB,
			value:              value,
			removalsCandidates: removals,
		})
		return
	}
	return
}

func (p puzzle) forEach(fn func(point app.Point, val uint8, stop *bool), excludePoints ...app.Point) {
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
			fn(point, p[row][col], &stop)
		}
	}
}

func (p puzzle) forEachInRow(row int, fn func(point app.Point, val uint8, stop *bool), excludeColumns ...int) {
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
		fn(app.Point{Row: row, Col: col}, p[row][col], &stop)
	}
}

func (p puzzle) forEachInCol(col int, fn func(point app.Point, val uint8, stop *bool), excludeRows ...int) {
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
		fn(app.Point{Row: row, Col: col}, p[row][col], &stop)
	}
}

func (p puzzle) forEachInBox(point app.Point, fn func(point app.Point, val uint8, stop *bool), excludePoints ...app.Point) {
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
			fn(point, p[row][col], &stop)
		}
	}
}

// ASCII representation of the puzzle when debugging.
func (p puzzle) debug() string {
	var out strings.Builder
	out.WriteString("╔═══════╤═══════╤═══════╗  \n")
	for row := 0; row < size; row++ {
		out.WriteString("║ ")
		for col := 0; col < size; col++ {
			if clue := int(p[row][col]); clue == 0 {
				out.WriteByte(' ')
			} else {
				out.WriteString(strconv.Itoa(clue))
			}
			if col%sizeGrp == sizeGrp-1 && col != size-1 {
				out.WriteString(" │ ")
			} else {
				out.WriteByte(' ')
			}
		}
		out.WriteString(fmt.Sprintf("║ %s\n", string('a'+byte(row))))
		if row%sizeGrp == sizeGrp-1 && row != size-1 {
			out.WriteString("╟───────┼───────┼───────╢  \n")
		}
	}
	out.WriteString("╚═══════╧═══════╧═══════╝  \n")
	out.WriteString("  1 2 3   4 5 6   7 8 9    ")
	return out.String()
}
