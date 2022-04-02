package puzzle_library

import (
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/puzzle_library/sudoku_classic"
)

func GetGenerator(typ app.PuzzleType) (app.PuzzleGenerator, error) {
	switch typ {
	case app.PuzzleSudokuClassic:
		return sudoku_classic.Generator{}, nil
	default:
		return nil, app.ErrorPuzzleTypeUnknown
	}
}
