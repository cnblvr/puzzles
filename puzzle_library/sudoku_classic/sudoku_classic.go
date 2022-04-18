package sudoku_classic

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
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

func (p puzzle) clone() (out puzzle) {
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			out[row][col] = p[row][col]
		}
	}
	return
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

type SudokuClassic struct{}

func (sc SudokuClassic) Type() app.PuzzleType {
	return app.PuzzleSudokuClassic
}

// NewRandomSolution generates a solution randomly for further extraction of
// digits.
func (sc SudokuClassic) NewRandomSolution() (s app.PuzzleGenerator, seed int64) {
	seedBts := make([]byte, 8)
	if _, err := crand.Reader.Read(seedBts); err != nil {
		panic(err)
	}
	seed = int64(binary.LittleEndian.Uint64(seedBts))
	return sc.NewSolutionBySeed(seed), seed
}

// NewSolutionBySeed generates a solution with a given seed for further
// extraction of digits.
func (sc SudokuClassic) NewSolutionBySeed(seed int64) app.PuzzleGenerator {
	rnd := rand.New(rand.NewSource(seed))

	s := generateWithoutShuffling(rnd)

	s.shuffle(rnd)

	return &s
}

func (p puzzle) shuffle(rnd *rand.Rand) {

	// swap of horizontal or vertical lines within one "big" line
	// TODO: imperfect randomization
	for i := 0; i < (rnd.Int()%1024)+1024; i++ {
		typ := app.Horizontal
		if rnd.Int()%2 == 1 {
			typ = app.Vertical
		}
		line := rnd.Int() % 9
		_ = p.SwapLines(typ, line, neighborLine(line, rnd.Int()%2))
	}

	// TODO: swap "big" lines

	// horizontal reflection
	if rnd.Int()%2 == 1 {
		_ = p.Reflect(app.ReflectHorizontal)
	}
	// vertical reflection
	if rnd.Int()%2 == 1 {
		_ = p.Reflect(app.ReflectVertical)
	}

	// rotate the puzzle by a random angle
	_ = p.Rotate(app.RotationType(rnd.Int() % 4))
}

// Calculation of the neighboring line.
// lineIdx in the range [0,8].
// neighbor can take values {0,1}.
func neighborLine(lineIdx int, neighbor int) int {
	switch neighbor {
	case 0:
		switch lineIdx % 3 {
		case 0:
			return lineIdx + 1
		case 1:
			return lineIdx - 1
		case 2:
			return lineIdx - 2
		default:
			return lineIdx
		}
	case 1:
		switch lineIdx % 3 {
		case 0:
			return lineIdx + 2
		case 1:
			return lineIdx + 1
		case 2:
			return lineIdx - 1
		default:
			return lineIdx
		}
	default:
		return lineIdx
	}
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

func (p *puzzle) solve(candidates puzzleCandidates, chanSteps chan<- app.PuzzleStep, strategies app.PuzzleStrategy) (changed bool, candidatesOut string, err error) {
	changedOnIteration := true
	if chanSteps != nil {
		defer close(chanSteps)
	}
	for changedOnIteration {
		changedOnIteration = false

		var step app.PuzzleStep
		candidatesBase := candidates.clone()
		changedOnIteration, step, err = p.solveOneStep(candidates, candidatesBase, strategies)
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

func (p *puzzle) Solve(candidatesIn string, chanSteps chan<- app.PuzzleStep, strategies app.PuzzleStrategy) (changed bool, candidatesOut string, err error) {
	var candidates puzzleCandidates
	if candidatesIn == "" {
		candidates = p.findSimpleCandidates()
	} else {
		candidates, err = decodeCandidates(candidatesIn)
		if err != nil {
			return
		}
		p.optimizeCandidates(&candidates)
	}
	defer func(candidates puzzleCandidates) {
		candidatesOut = candidates.encode()
	}(candidates)

	return p.solve(candidates, chanSteps, strategies)
}

func (p *puzzle) SolveOneStep(candidatesIn string, strategies app.PuzzleStrategy) (candidatesChanges string, step app.PuzzleStep, err error) {
	var candidates, candidatesBase puzzleCandidates
	if candidatesIn == "" {
		candidates = p.findSimpleCandidates()
	} else {
		candidates, err = decodeCandidates(candidatesIn)
		if err != nil {
			return
		}
		p.optimizeCandidates(&candidates)
	}
	candidatesBase = candidates.clone()
	defer func(candidates puzzleCandidates) {
		candidatesChanges = candidates.encodeOnlyChanges(candidatesBase)
	}(candidates)

	_, step, err = p.solveOneStep(candidates, candidatesBase, strategies)
	if err != nil {
		return
	}

	return
}

func (p *puzzle) solveOneStep(candidates puzzleCandidates, candidatesBase puzzleCandidates, strategies app.PuzzleStrategy) (changed bool, step puzzleStepSetter, err error) {
	makeStep := func(s puzzleStepSetter) {
		switch s := s.(type) {
		case *puzzleStepSet:
			p[s.point.Row][s.point.Col] = s.value
			candidates.simpleRemoveAfterSet(s.point, s.value)
		}
		s.setCandidateChanges(candidates.encodeOnlyChanges(candidatesBase))
		step = s
		changed = true
	}

	// strategy Naked Single
	if strategies.Has(app.StrategyNakedSingle) {
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
			makeStep(&puzzleStepSet{
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
	}

	// strategy Hidden Single
	if strategies.Has(app.StrategyHiddenSingle) {
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
				makeStep(&puzzleStepSet{
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
	}

	// strategy Naked Pair
	if strategies.Has(app.StrategyNakedPair) {
		if points, pair, ok := candidates.strategyNakedPair(); ok {
			makeStep(&puzzleStepNakedStrategy{
				points: points,
				set:    pair,
			})
			return
		}
	}
	// strategy Naked Triple
	if strategies.Has(app.StrategyNakedTriple) {
		if points, triple, ok := candidates.strategyNakedTriple(); ok {
			makeStep(&puzzleStepNakedStrategy{
				points: points,
				set:    triple,
			})
			return
		}
	}
	// strategy Hidden Pair
	if strategies.Has(app.StrategyHiddenPair) {
		if points, pair, ok := candidates.strategyHiddenPair(); ok {
			makeStep(&puzzleStepHiddenStrategy{
				points: points,
				set:    pair,
			})
			return
		}
	}
	// strategy Hidden Triple
	if strategies.Has(app.StrategyHiddenTriple) {
		if points, triple, ok := candidates.strategyHiddenTriple(); ok {
			makeStep(&puzzleStepHiddenStrategy{
				points: points,
				set:    triple,
			})
			return
		}
	}
	// strategy Pointing Pair or Triple
	if strategies.Has(app.StrategyPointingPair) || strategies.Has(app.StrategyPointingTriple) {
		if points, value, ok := candidates.strategyPointingPairTriple(); ok {
			makeStep(&puzzleStepPointingStrategy{
				points: points,
				value:  value,
			})
			return
		}
	}
	// strategy Box Line Reduction Pair or Triple
	if strategies.Has(app.StrategyBoxLineReductionPair) || strategies.Has(app.StrategyBoxLineReductionTriple) {
		if points, value, ok := candidates.strategyBoxLineReductionPairTriple(); ok {
			makeStep(&puzzleStepBoxLineReductionStrategy{
				points: points,
				value:  value,
			})
			return
		}
	}
	// strategy X-Wing
	if strategies.Has(app.StrategyXWing) {
		if pairA, pairB, value, ok := candidates.strategyXWing(); ok {
			makeStep(&puzzleStepXWingStrategy{
				pairA: pairA,
				pairB: pairB,
				value: value,
			})
			return
		}
	}
	return
}

func getRandomCountCluesBy(rnd *rand.Rand, level app.PuzzleLevel) int {
	min, max := 0, 0
	switch level {
	case app.PuzzleLevelEasy:
		min, max = 33, 37
	case app.PuzzleLevelNormal:
		min, max = 28, 32
	case app.PuzzleLevelHard:
		min, max = 17, 27
	}
	return (rnd.Int() % (max - min + 1)) + min
}

func (p *puzzle) GenerateLogic(seed int64, strategies app.PuzzleStrategy) (app.PuzzleStrategy, error) {
	rnd := rand.New(rand.NewSource(seed))
	givenStrategies := app.StrategyUnknown
	limitClues := getRandomCountCluesBy(rnd, strategies.Level())
	removedClues := 0
	for _, point := range getRandomPoints(rnd) {
		oneRemoveStrategies := givenStrategies
		if 81-removedClues <= limitClues {
			return givenStrategies, nil
		}
		digit := p[point.Row][point.Col]
		p[point.Row][point.Col] = 0
		removedClues++
		revert := func(p *puzzle) {
			p[point.Row][point.Col] = digit
			removedClues--
		}
		candidates := p.findSimpleCandidates()
		var wg sync.WaitGroup
		wg.Add(1)
		var err error
		chanSteps := make(chan app.PuzzleStep)
		solution := p.clone()
		go func() {
			defer wg.Done()
			_, _, err = solution.solve(candidates, chanSteps, strategies)
		}()
		for step := range chanSteps {
			//log.Printf("%010b %+v", oneRemoveStrategies, step)
			oneRemoveStrategies |= step.Strategy()
		}
		wg.Wait()
		if err != nil {
			revert(p)
			continue
		}
		if !solution.isSolved() {
			revert(p)
			continue
		}
		givenStrategies = oneRemoveStrategies
	}
	return givenStrategies, nil
}

func (p puzzle) GenerateRandom(seed int64) error {
	return errors.Errorf("not implemented")
}

func (p *puzzle) MakeUserStep(candidatesIn string, step app.PuzzleUserStep) (candidatesOut string, wrongCandidates string, err error) {
	var c puzzleCandidates
	c, err = decodeCandidates(candidatesIn)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	switch step.Type {
	case app.UserStepSetDigit:
		p[step.Point.Row][step.Point.Col] = uint8(step.Digit)
	case app.UserStepDeleteDigit:
		p[step.Point.Row][step.Point.Col] = 0
	case app.UserStepSetCandidate:
		c[step.Point.Row][step.Point.Col].addInt8(step.Digit)
	case app.UserStepDeleteCandidate:
		c[step.Point.Row][step.Point.Col].delete(uint8(step.Digit))
	default:
		err = errors.WithStack(err)
		return
	}

	wrongs := p.getWrongCandidates(c)
	wrongCandidates = wrongs.encode()

	candidatesOut = c.encode()
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

func (p puzzle) isSolved() bool {
	out := true
	p.forEach(func(point1 app.Point, val1 uint8, stop1 *bool) {
		if val1 == 0 {
			*stop1, out = true, false
			return
		}
		fnCheck := func(point2 app.Point, val2 uint8, stop2 *bool) {
			if val1 == val2 {
				*stop1, *stop2, out = true, true, false
				return
			}
		}
		p.forEachInRow(point1.Row, fnCheck, point1.Col)
		p.forEachInCol(point1.Col, fnCheck, point1.Row)
		p.forEachInBox(point1, fnCheck, point1)
	})
	return out
}

func (p puzzle) GetWrongPoints() (points []app.Point) {
	pointsUnique := make(map[app.Point]struct{})
	p.forEach(func(point1 app.Point, val1 uint8, _ *bool) {
		fnCheck := func(point2 app.Point, val2 uint8, _ *bool) {
			if val1 != 0 && val1 == val2 {
				pointsUnique[point2] = struct{}{}
			}
		}
		p.forEachInRow(point1.Row, fnCheck, point1.Col)
		p.forEachInCol(point1.Col, fnCheck, point1.Row)
		p.forEachInBox(point1, fnCheck, point1)
	})
	for point := range pointsUnique {
		points = append(points, point)
	}
	sort.Slice(points, func(i, j int) bool {
		if points[i].Row == points[j].Row {
			return points[i].Col < points[j].Col
		}
		return points[i].Row < points[j].Row
	})
	return
}

// Get all puzzle points randomly.
func getRandomPoints(rnd *rand.Rand) []app.Point {
	var points []app.Point
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			points = append(points, app.Point{Row: row, Col: col})
		}
	}
	rnd.Shuffle(len(points), func(i, j int) {
		points[i], points[j] = points[j], points[i]
	})
	return points
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
