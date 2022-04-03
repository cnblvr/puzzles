package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"strconv"
)

type PuzzleRepository interface {
	// Errors: ErrorPuzzlePoolEmpty, ErrorPuzzleNotFound, unknown.
	CreateRandomPuzzleGame(ctx context.Context, params CreateRandomPuzzleGameParams) (*Puzzle, *PuzzleGame, error)

	// Errors: ErrorPuzzleGameNotFound, unknown.
	GetPuzzleGame(ctx context.Context, id uuid.UUID) (*PuzzleGame, error)

	CreatePuzzle(ctx context.Context, params CreatePuzzleParams) (*Puzzle, error)

	// Errors: ErrorPuzzleNotFound, unknown.
	GetPuzzle(ctx context.Context, id int64) (*Puzzle, error)

	// Errors: ErrorPuzzleGameNotFound, ErrorPuzzleNotFound, unknown.
	GetPuzzleByGameID(ctx context.Context, gameID uuid.UUID) (*Puzzle, error)

	// Errors: ErrorPuzzleGameNotFound, ErrorPuzzleNotFound, unknown.
	GetPuzzleAndGame(ctx context.Context, id uuid.UUID) (*Puzzle, *PuzzleGame, error)
}

type CreateRandomPuzzleGameParams struct {
	Session *Session
	Type    PuzzleType
	Level   PuzzleLevel
}

type CreatePuzzleParams struct {
	Type PuzzleType
	GeneratedPuzzle
}

type PuzzleAssistant interface {
	Type() PuzzleType
	GetCandidates(ctx context.Context, clues string) PuzzleCandidates
	FindUserErrors(ctx context.Context, userState string) []Point
}

type PuzzleGenerator interface {
	Type() PuzzleType
	GenerateSolution(ctx context.Context, seed int64, generatedSolutions chan<- GeneratedPuzzle)
	GenerateClues(ctx context.Context, seed int64, generatedSolution GeneratedPuzzle, generated chan<- GeneratedPuzzle)
}

type GeneratedPuzzle struct {
	Seed     int64
	Level    PuzzleLevel
	Meta     json.RawMessage
	Clues    string
	Solution string
}

type Puzzle struct {
	ID       int64           `json:"id" redis:"-"`
	Type     PuzzleType      `json:"type" redis:"type"`
	Seed     int64           `json:"seed" redis:"seed"`
	Level    PuzzleLevel     `json:"level" redis:"level"`
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

// Errors: ErrorPuzzleGameNotAllowed.
func (g *PuzzleGame) ValidateSession(session *Session) error {
	if g.UserID > 0 {
		if g.UserID != session.UserID {
			return ErrorPuzzleGameNotAllowed
		}
	} else {
		if g.SessionID != session.SessionID {
			return ErrorPuzzleGameNotAllowed
		}
	}
	return nil
}

type PuzzleCandidates map[Point][]int8

func (cs PuzzleCandidates) MarshalJSON() ([]byte, error) {
	out := make(map[string][]int8)
	for p, c := range cs {
		out[p.String()] = c
	}
	return json.Marshal(out)
}

var (
	ErrorPuzzleTypeUnknown    = fmt.Errorf("puzzle type unknown")
	ErrorPuzzlePoolEmpty      = fmt.Errorf("puzzle pool is empty")
	ErrorPuzzleNotFound       = fmt.Errorf("puzzle not found")
	ErrorPuzzleGameNotFound   = fmt.Errorf("puzzle game not found")
	ErrorPuzzleGameNotAllowed = fmt.Errorf("puzzle game not allowed")
)

type PuzzleType string

const (
	PuzzleSudokuClassic PuzzleType = "sudoku_classic"
	PuzzleJigsaw        PuzzleType = "jigsaw"
	PuzzleWindoku       PuzzleType = "windoku"
	PuzzleSudokuX       PuzzleType = "sudoku_x"
	PuzzleKakuro        PuzzleType = "kakuro"
)

func (t PuzzleType) String() string {
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

type Point struct {
	Row, Col int
}

func (p Point) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p Point) InSameBox(points ...Point) bool {
	boxRow, boxCol := p.Row/3, p.Col/3
	for _, ip := range points {
		if boxRow != ip.Row {
			return false
		}
		if boxCol != ip.Col {
			return false
		}
	}
	return true
}

func (p Point) InSameRow(points ...Point) bool {
	row := p.Row
	for _, ip := range points {
		if row != ip.Row {
			return false
		}
	}
	return true
}

func (p Point) InSameCol(points ...Point) bool {
	col := p.Col
	for _, ip := range points {
		if col != ip.Col {
			return false
		}
	}
	return true
}

func (p Point) String() string {
	return fmt.Sprintf("%s%d", string('a'+byte(p.Row)), p.Col+1)
}

func PointFromString(s string) (Point, error) {
	if len(s) < 2 {
		return Point{}, fmt.Errorf("unknown format Point")
	}
	p := Point{}

	switch ch := s[0]; {
	case 'a' <= ch && ch <= 'z':
		p.Row = int(ch) - 'a'
	case 'A' <= ch && ch <= 'Z':
		p.Row = int(ch) - 'A'
	default:
		return Point{}, fmt.Errorf("unknown format Point")
	}

	var err error
	p.Col, err = strconv.Atoi(s[1:])
	if err != nil {
		return Point{}, err
	}
	p.Col--

	return p, nil
}
