package puzzle_library

import (
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/puzzle_library/sudoku_classic"
)

type PuzzleLibrary struct{}

func (PuzzleLibrary) GetCreator(typ app.PuzzleType) (app.PuzzleCreator, error) {
	switch typ {
	case app.PuzzleSudokuClassic:
		return sudoku_classic.SudokuClassic{}, nil
	default:
		return nil, app.ErrorPuzzleTypeUnknown
	}
}

func (PuzzleLibrary) GetGenerator(typ app.PuzzleType, puzzle string) (app.PuzzleGenerator, error) {
	switch typ {
	case app.PuzzleSudokuClassic:
		return sudoku_classic.ParseGenerator(puzzle)
	default:
		return nil, app.ErrorPuzzleTypeUnknown
	}
}

func (PuzzleLibrary) GetAssistant(typ app.PuzzleType, puzzle string) (app.PuzzleAssistant, error) {
	switch typ {
	case app.PuzzleSudokuClassic:
		return sudoku_classic.ParseAssistant(puzzle)
	default:
		return nil, app.ErrorPuzzleTypeUnknown
	}
}
