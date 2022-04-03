package sudoku_classic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"math/rand"
	"strconv"
	"time"
)

type SudokuClassic struct{}

func (SudokuClassic) Type() app.PuzzleType {
	return app.PuzzleSudokuClassic
}

func (SudokuClassic) GenerateSolution(ctx context.Context, seed int64, generatedSolutions chan<- app.GeneratedPuzzle) {
	// randomizer for solution generation
	rnd := rand.New(rand.NewSource(seed))

	defer close(generatedSolutions)

	// solution generation without shuffling
	solution := generateSudokuBoard(rnd)

	// swap of horizontal or vertical lines within one "big" line
	// TODO: imperfect randomization
	for i := 0; i < (rnd.Int()%1024)+1024; i++ {
		typ := Horizontal
		if rnd.Int()%2 == 1 {
			typ = Vertical
		}
		line := rnd.Int() % 9
		solution.swapLines(typ, line, neighborLine(line, rnd.Int()%2))
	}

	// TODO: swap "big" lines

	// horizontal reflection
	if rnd.Int()%2 == 1 {
		solution.reflect(Horizontal)
	}
	// vertical reflection
	if rnd.Int()%2 == 1 {
		solution.reflect(Vertical)
	}

	// rotate the puzzle by a random angle
	solution.rotate(RotationType(rnd.Int() % 4))

	generatedSolutions <- app.GeneratedPuzzle{
		Seed:     seed,
		Meta:     []byte("{}"),
		Solution: solution.String(),
	}

	return
}

func (SudokuClassic) GenerateClues(ctx context.Context, seed int64, generatedSolution app.GeneratedPuzzle, generated chan<- app.GeneratedPuzzle) {
	// randomizer for clues generation
	rnd := rand.New(rand.NewSource(seed))

	defer close(generated)

	solution := sudokuPuzzleFromString(generatedSolution.Solution)

	puzzle := make(sudokuPuzzle, 9)
	for row := 0; row < 9; row++ {
		puzzle[row] = make([]int8, 9)
		copy(puzzle[row], solution[row])
	}

	needHints := make(map[int]app.PuzzleLevel)
	for _, level := range []app.PuzzleLevel{app.PuzzleLevelEasy, app.PuzzleLevelMedium} {
		min, max := getMinMaxCluesOfLevel(level)
		hints := (rnd.Int() % (max - min + 1)) + min
		needHints[hints] = level
	}

	removes := 0
	saveHardIfMatched := func() {
		if _, max := getMinMaxCluesOfLevel(app.PuzzleLevelHard); max >= removes-81 {
			generated <- app.GeneratedPuzzle{
				Seed:     seed,
				Level:    app.PuzzleLevelHard,
				Meta:     generatedSolution.Meta,
				Clues:    puzzle.String(),
				Solution: solution.String(),
			}
		}
	}

	rndPoints := sudokuRandomPoints(rnd)
	for _, p := range rndPoints {
		//log.Printf("point #%d: %v; hints %d", idx+1, p, 81-removes)
		select {
		case <-ctx.Done():
			saveHardIfMatched()
			return
		default:
		}
		digit := puzzle[p.Row][p.Col]
		if digit == 0 {
			continue
		}
		puzzle[p.Row][p.Col] = 0
		if func() bool {
			ctxSolve, cancelSolve := context.WithTimeout(context.Background(), 20*time.Minute)
			defer cancelSolve()
			solutions, err := puzzle.solveBruteForce(ctxSolve, 2)
			if len(solutions) != 1 || err != nil {
				puzzle[p.Row][p.Col] = digit
				return true
			}
			removes++
			return false
		}() {
			continue
		}

		if level, ok := needHints[81-removes]; ok {
			generated <- app.GeneratedPuzzle{
				Seed:     seed,
				Level:    level,
				Meta:     generatedSolution.Meta,
				Clues:    puzzle.String(),
				Solution: solution.String(),
			}
		}
	}

	saveHardIfMatched()
	return
}

// DirectionType is a direction of line/"big" line/some kind of field change.
type DirectionType uint8

const (
	Horizontal DirectionType = iota
	Vertical
)

// RotationType is an angle of rotation.
type RotationType uint8

const (
	Rotate0 RotationType = iota
	Rotate90
	Rotate180
	Rotate270
)

func getMinMaxCluesOfLevel(l app.PuzzleLevel) (int, int) {
	switch l {
	case app.PuzzleLevelEasy:
		return 33, 37
	case app.PuzzleLevelMedium:
		return 28, 32
	case app.PuzzleLevelHard:
		return 17, 27
	default:
		return 81, 81
	}
}

func (SudokuClassic) GetCandidates(ctx context.Context, puzzle string) string {
	p := sudokuPuzzleFromString(puzzle)
	out, err := json.Marshal(p.findCandidates())
	if err != nil {
		log.Error().Err(err).Msg("failed to encode candidates")
	}
	return string(out)
}

func (SudokuClassic) FindUserErrors(ctx context.Context, userState string) []app.Point {
	return sudokuPuzzleFromString(userState).FindUserErrors()
}

func (SudokuClassic) FindUserCandidatesErrors(ctx context.Context, stateStr string, stateCandidatesJson string) string {
	state := sudokuPuzzleFromString(stateStr)
	stateCandidates := sudokuCandidates{}
	err := json.Unmarshal([]byte(stateCandidatesJson), &stateCandidates)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode candidates")
	}

	listErrors := newSudokuCandidates()
	stateCandidates.forEach(func(p1 app.Point, candidates1 []int8) {
		if state.In(p1) > 0 {
			return
		}
		for _, c := range candidates1 {
			findErrs := func(_ app.Point, value2 int8, break2 *bool) {
				if c == value2 {
					listErrors[p1.Row][p1.Col][c] = struct{}{}
					*break2 = true
				}
			}
			state.forEachInRow(p1.Row, findErrs, p1.Col)
			state.forEachInCol(p1.Col, findErrs, p1.Row)
			state.forEachInBox(p1, findErrs, p1)
		}
	})
	out, err := json.Marshal(listErrors)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode candidates")
	}
	return string(out)
}

func (SudokuClassic) MakeStep(ctx context.Context, stateStr string, stateCandidatesJson string, step app.PuzzleStep) (string, string, error) {
	state := sudokuPuzzleFromString(stateStr)
	stateCandidates := sudokuCandidates{}
	if err := json.Unmarshal([]byte(stateCandidatesJson), &stateCandidates); err != nil {
		return "", "", errors.Wrap(err, "failed to decode state candidates")
	}

	switch step.Type {
	case app.StepSetDigit:
		state[step.Point.Row][step.Point.Col] = step.Digit
	case app.StepDeleteDigit:
		state[step.Point.Row][step.Point.Col] = 0
	case app.StepSetCandidate:
		stateCandidates[step.Point.Row][step.Point.Col][step.Digit] = struct{}{}
	case app.StepDeleteCandidate:
		delete(stateCandidates[step.Point.Row][step.Point.Col], step.Digit)
	}

	newStateCandidates, err := json.Marshal(stateCandidates)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to encode new state candidates")
	}

	return state.String(), string(newStateCandidates), nil
}

// Puzzle generation without shuffling.
func generateSudokuBoard(rnd *rand.Rand) sudokuPuzzle {
	b := make(sudokuPuzzle, 9)
	for i := 0; i < 9; i++ {
		b[i] = make([]int8, 9)
	}
	// Generate first line randomly
	digits := []int8{1, 2, 3, 4, 5, 6, 7, 8, 9}
	line := make([]int8, 0, 9)
	for len(digits) > 0 {
		idx := rnd.Int() % len(digits)
		line = append(line, digits[idx])
		digits = append(digits[:idx], digits[idx+1:]...)
	}
	copy(b[0], line)

	// The second line is the offset of the first line to the left by 3
	line = append(line[3:9], line[:3]...)
	copy(b[1], line)

	// The third line is the offset of the second line to the left by 3
	line = append(line[3:9], line[:3]...)
	copy(b[2], line)

	// First "big" horizontally line is completed. Next lines generate by this algorithm:
	//  line n:   is offset of the previous line to the left by 1
	//  line n+1: is offset of the previous line to the left by 3
	//  line n+2: is offset of the previous line to the left by 3
	line = append(line[1:9], line[0]) // n
	copy(b[3], line)
	line = append(line[3:9], line[:3]...) // n+1
	copy(b[4], line)
	line = append(line[3:9], line[:3]...) // n+2
	copy(b[5], line)

	// Generation of third "big" horizontally line.
	line = append(line[1:9], line[0]) // n
	copy(b[6], line)
	line = append(line[3:9], line[:3]...) // n+1
	copy(b[7], line)
	line = append(line[3:9], line[:3]...) // n+2
	copy(b[8], line)

	return b
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

func sudokuString(s [][]int8) (out string) {
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			val := strconv.Itoa(int(s[row][col]))
			if val == "0" {
				val = "."
			}
			out += val
		}
	}
	return
}

// ASCII representation of the puzzle when debugging.
func sudokuDebug(s [][]int8) (out string) {
	out += "╔═══════╤═══════╤═══════╗\n"
	for i := 0; i < 9; i++ {
		out += "║ "
		for j := 0; j < 9; j++ {
			space := " "
			if j%3 == 2 && j != 8 {
				space = " │ "
			}
			value := strconv.Itoa(int(s[i][j]))
			if value == "0" {
				value = " "
			}
			out += fmt.Sprintf("%s%s", value, space)
		}
		out += fmt.Sprintf("║ %s\n", string('a'+byte(i)))
		if i%3 == 2 && i != 8 {
			out += "╟───────┼───────┼───────╢\n"
		}
	}
	out += "╚═══════╧═══════╧═══════╝\n"
	out += "  1 2 3   4 5 6   7 8 9  "
	return out
}

// Get all puzzle points randomly.
func sudokuRandomPoints(rnd *rand.Rand) []app.Point {
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
