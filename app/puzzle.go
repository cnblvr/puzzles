package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"strconv"
)

//go:generate stringer -type=PuzzleStrategy -linecomment -trimprefix Strategy -output puzzle_strategy_string.go

type PuzzleRepository interface {
	// Errors: ErrorPuzzlePoolEmpty, ErrorPuzzleNotFound, unknown.
	CreateRandomPuzzleGame(ctx context.Context, params CreateRandomPuzzleGameParams) (*Puzzle, *PuzzleGame, error)

	// Errors: ErrorPuzzleGameNotFound, unknown.
	GetPuzzleGame(ctx context.Context, id uuid.UUID) (*PuzzleGame, error)

	UpdatePuzzleGame(ctx context.Context, game *PuzzleGame) error

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
	String() string
	Type() PuzzleType
	//GetCandidates(ctx context.Context, clues string) string
	//FindUserErrors(ctx context.Context, userState string) []Point
	//FindUserCandidatesErrors(ctx context.Context, state string, stateCandidates string) string
	//MakeStep(ctx context.Context, state string, stateCandidates string, step PuzzleStep) (string, string, error)
}

type PuzzleGenerator interface {
	String() string
	Type() PuzzleType
	SwapLines(dir DirectionType, a, b int) error
	SwapBigLines(dir DirectionType, a, b int) error
	Rotate(r RotationType) error
	Reflect(r ReflectionType) error
	SwapDigits(a, b uint8) error
	Solve(candidatesIn string, chanSteps chan<- PuzzleStep) (changed bool, candidatesOut string, err error)
	SolveOneStep(candidatesIn string) (candidatesOut string, step PuzzleStep, err error)
	//GenerateSolution(ctx context.Context, seed int64, generatedSolutions chan<- GeneratedPuzzle)
	//GenerateClues(ctx context.Context, seed int64, generatedSolution GeneratedPuzzle, generated chan<- GeneratedPuzzle)
}

type PuzzleStep interface {
	Strategy() PuzzleStrategy
	Description() string
}

type PuzzleStrategy uint64

const (
	StrategyNakedSingle            PuzzleStrategy = 1 << iota // Naked Single
	StrategyNakedPair                                         // Naked Pair
	StrategyNakedTriple                                       // Naked Triple
	StrategyNakedQuad                                         // Naked Quad
	StrategyHiddenSingle                                      // Hidden Single
	StrategyHiddenPair                                        // Hidden Pair
	StrategyHiddenTriple                                      // Hidden Triple
	StrategyHiddenQuad                                        // Hidden Quad
	StrategyPointingPair                                      // Pointing Pair
	StrategyPointingTriple                                    // Pointing Triple
	StrategyBoxLineReductionPair                              // Box/Line Reduction Pair
	StrategyBoxLineReductionTriple                            // Box/Line Reduction Triple
	StrategyXWing                                             // X-Wing
	StrategyUnknown                PuzzleStrategy = 0
)

type RotationType uint8

const (
	RotateTo90 RotationType = iota + 1
	RotateTo180
	RotateTo270
)

type DirectionType uint8

const (
	Horizontal DirectionType = iota
	Vertical
)

type ReflectionType uint8

const (
	ReflectHorizontal ReflectionType = iota + 1
	ReflectVertical
	ReflectMajorDiagonal
	ReflectMinorDiagonal
)

type GeneratedPuzzle struct {
	Seed       int64
	Level      PuzzleLevel
	Meta       string
	Clues      string
	Candidates string
	Solution   string
}

type Puzzle struct {
	ID         int64       `json:"id" redis:"-"`
	Type       PuzzleType  `json:"type" redis:"type"`
	Seed       int64       `json:"seed" redis:"seed"`
	Level      PuzzleLevel `json:"level" redis:"level"`
	Meta       string      `json:"meta" redis:"meta"`
	Clues      string      `json:"clues" redis:"clues"`
	Candidates string      `json:"candidates" redis:"candidates"`
	Solution   string      `json:"solution" redis:"solution"`
}

type PuzzleGame struct {
	ID                uuid.UUID `json:"id" redis:"-"`
	SessionID         int64     `json:"session_id,omitempty" redis:"session_id"`
	UserID            int64     `json:"user_id,omitempty" redis:"user_id"`
	PuzzleID          int64     `json:"puzzle_id" redis:"puzzle_id"`
	IsNew             bool      `json:"is_new" redis:"is_new"`
	State             string    `json:"state" redis:"state"`
	StateCandidates   string    `json:"state_candidates" redis:"state_candidates"`
	IsWin             bool      `json:"is_win" redis:"is_win"`
	CandidatesAtStart bool      `json:"candidates_at_start" redis:"candidates_at_start"`
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

type PuzzleUserStep struct {
	Type  UserStepType `json:"type"`
	Point Point        `json:"point"`
	Digit int8         `json:"digit"`
}

type UserStepType string

const (
	UserStepSetDigit        UserStepType = "set_digit"
	UserStepDeleteDigit     UserStepType = "del_digit"
	UserStepSetCandidate    UserStepType = "set_cand"
	UserStepDeleteCandidate UserStepType = "del_cand"
)

func (t UserStepType) Validate() error {
	switch t {
	case UserStepSetDigit, UserStepDeleteDigit, UserStepSetCandidate, UserStepDeleteCandidate:
	default:
		return errors.Errorf("unknown step type")
	}
	return nil
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
	DefaultPuzzleType              = PuzzleSudokuClassic
)

func (t PuzzleType) String() string {
	return string(t)
}

type PuzzleLevel uint8

const (
	PuzzleLevelEasy PuzzleLevel = iota + 1
	PuzzleLevelNormal
	PuzzleLevelHard
	PuzzleLevelHarder
	PuzzleLevelInsane
	PuzzleLevelDemon
	DefaultPuzzleLevel = PuzzleLevelNormal
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

func (p *Point) UnmarshalJSON(bts []byte) error {
	var str string
	if err := json.Unmarshal(bts, &str); err != nil {
		return errors.WithStack(err)
	}
	decoded, err := PointFromString(str)
	if err != nil {
		return errors.WithStack(err)
	}
	*p = decoded
	return nil
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
