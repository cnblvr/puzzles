package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)

type PuzzleGameRepository interface {
	// Errors: ErrorPuzzlePoolEmpty, ErrorPuzzleNotFound, unknown.
	CreateRandomPuzzleGame(ctx context.Context, params CreateRandomPuzzleGameParams) (*Puzzle, *PuzzleGame, error)
	// Errors: ErrorPuzzleGameNotFound, unknown.
	GetPuzzleGame(ctx context.Context, id uuid.UUID) (*PuzzleGame, error)
}

type CreateRandomPuzzleGameParams struct {
	Session    *Session
	Type       PuzzleType
	SudokuType PuzzleSudokuType
	Level      PuzzleLevel
}

type Puzzle struct {
	ID       int64           `json:"id" redis:"-"`
	Meta     json.RawMessage `json:"meta" redis:"meta"`
	Clues    string          `json:"clues" redis:"clues"`
	Solution string          `json:"solution" redis:"solution"`
}

type PuzzleGame struct {
	ID        uuid.UUID `json:"id" redis:"-"`
	SessionID int64     `json:"session_id,omitempty" redis:"session_id"`
	UserID    int64     `json:"user_id,omitempty" redis:"user_id"`
	PuzzleID  int64     `json:"puzzle_id" redis:"puzzle_id"`
}

var (
	ErrorPuzzlePoolEmpty    = fmt.Errorf("puzzle pool is empty")
	ErrorPuzzleNotFound     = fmt.Errorf("puzzle not found")
	ErrorPuzzleGameNotFound = fmt.Errorf("puzzle game not found")
)

type PuzzleType string

const (
	PuzzleTypeSudoku PuzzleType = "sudoku"
	PuzzleTypeKakuru PuzzleType = "kakuro"
)

func (t PuzzleType) String() string {
	return string(t)
}

type PuzzleSudokuType string

const (
	PuzzleSudokuTypeClassic PuzzleSudokuType = "classic"
	PuzzleSudokuTypeJigsaw  PuzzleSudokuType = "jigsaw"
	PuzzleSudokuTypeWindoku PuzzleSudokuType = "windoku"
	PuzzleSudokuTypeSudokuX PuzzleSudokuType = "sudoku_x"
)

func (t PuzzleSudokuType) String() string {
	return string(t)
}

type PuzzleLevel string

const (
	PuzzleLevelEasy   PuzzleLevel = "easy"
	PuzzleLevelMedium PuzzleLevel = "medium"
	PuzzleLevelHard   PuzzleLevel = "hard"
)

func (l PuzzleLevel) String() string {
	return string(l)
}
