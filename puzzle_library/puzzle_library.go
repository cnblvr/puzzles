package puzzle_library

import (
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/puzzle_library/sudoku_classic"
)

func GetGenerator(typ app.PuzzleType) (app.PuzzleGenerator, error) {
	switch typ {
	case app.PuzzleSudokuClassic:
		return sudoku_classic.SudokuClassic{}, nil
	default:
		return nil, app.ErrorPuzzleTypeUnknown
	}
}

func GetAssistant(typ app.PuzzleType) (app.PuzzleAssistant, error) {
	switch typ {
	case app.PuzzleSudokuClassic:
		return sudoku_classic.SudokuClassic{}, nil
	default:
		return nil, app.ErrorPuzzleTypeUnknown
	}
}
