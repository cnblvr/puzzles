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

	// it's sloooooowly (maybe)
	GetAmountUnsolvedPuzzlesForAllUsers(ctx context.Context, typ PuzzleType, level PuzzleLevel) (int, error)
}

type PuzzleLibrary interface {
	// Errors: ErrorPuzzleTypeUnknown, unknown.
	GetCreator(typ PuzzleType) (PuzzleCreator, error)
	// Errors: ErrorPuzzleTypeUnknown, unknown.
	GetGenerator(typ PuzzleType, puzzle string) (PuzzleGenerator, error)
	// Errors: ErrorPuzzleTypeUnknown, unknown.
	GetAssistant(typ PuzzleType, puzzle string) (PuzzleAssistant, error)
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

type PuzzleCreator interface {
	Type() PuzzleType
	NewRandomSolution() (s PuzzleGenerator, seed int64)
	NewSolutionBySeed(seed int64) PuzzleGenerator
}

type PuzzleAssistant interface {
	String() string
	Type() PuzzleType
	GetWrongPoints() []Point
	GetWrongCandidates(candidates string) (string, error)
	MakeUserStep(candidatesIn string, step PuzzleUserStep) (candidatesOut string, wrongCandidates string, err error)
	SolveOneStep(candidatesIn string, strategies PuzzleStrategy) (candidatesChanges string, step PuzzleStep, err error)
	//GetCandidates(ctx context.Context, clues string) string
	//FindUserErrors(ctx context.Context, userState string) []Point
	//FindUserCandidatesErrors(ctx context.Context, state string, stateCandidates string) string
	//MakeStep(ctx context.Context, state string, stateCandidates string, step PuzzleStep) (string, string, error)
}

type PuzzleGenerator interface {
	String() string
	Type() PuzzleType
	GetCandidates() string
	GetWrongPoints() []Point
	SwapLines(dir DirectionType, a, b int) error
	SwapBigLines(dir DirectionType, a, b int) error
	Rotate(r RotationType) error
	Reflect(r ReflectionType) error
	SwapDigits(a, b uint8) error
	Solve(candidatesIn string, chanSteps chan<- PuzzleStep, strategies PuzzleStrategy) (changed bool, candidatesOut string, err error)
	SolveOneStep(candidatesIn string, strategies PuzzleStrategy) (candidatesChanges string, step PuzzleStep, err error)
	GenerateLogic(seed int64, strategies PuzzleStrategy) (PuzzleStrategy, error)
	GenerateRandom(seed int64) error
	//GenerateSolution(ctx context.Context, seed int64, generatedSolutions chan<- GeneratedPuzzle)
	//GenerateClues(ctx context.Context, seed int64, generatedSolution GeneratedPuzzle, generated chan<- GeneratedPuzzle)
}

type PuzzleStep interface {
	Strategy() PuzzleStrategy
	CandidateChanges() string
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

	levelEasyStrategies   = StrategyNakedSingle
	levelNormalStrategies = StrategyNakedPair | StrategyNakedTriple | StrategyHiddenSingle | StrategyHiddenPair | StrategyHiddenTriple
	levelHardStrategies   = StrategyNakedQuad | StrategyHiddenQuad | StrategyPointingPair | StrategyPointingTriple |
		StrategyBoxLineReductionPair | StrategyBoxLineReductionTriple
	levelHarderStrategies = StrategyXWing // TODO
	levelInsaneStrategies = 0             // TODO
	levelDemonStrategies  = 0             // TODO
)

func (i PuzzleStrategy) Has(s PuzzleStrategy) bool {
	return i&s > 0
}

func (i PuzzleStrategy) Level() PuzzleLevel {
	switch {
	case i&levelDemonStrategies > 0:
		return PuzzleLevelDemon
	case i&levelInsaneStrategies > 0:
		return PuzzleLevelInsane
	case i&levelHarderStrategies > 0:
		return PuzzleLevelHarder
	case i&levelHardStrategies > 0:
		return PuzzleLevelHard
	case i&levelNormalStrategies > 0:
		return PuzzleLevelNormal
	case i&levelEasyStrategies > 0:
		return PuzzleLevelEasy
	default:
		return ""
	}
}

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
	ID              uuid.UUID `json:"id" redis:"-"`
	SessionID       int64     `json:"session_id,omitempty" redis:"session_id"`
	UserID          int64     `json:"user_id,omitempty" redis:"user_id"`
	PuzzleID        int64     `json:"puzzle_id" redis:"puzzle_id"`
	IsNew           bool      `json:"is_new" redis:"is_new"`
	State           string    `json:"state" redis:"state"`
	StateCandidates string    `json:"state_candidates" redis:"state_candidates"`
	IsWin           bool      `json:"is_win" redis:"is_win"`
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
	PuzzleSudokuClassic PuzzleType = "sudoku_classic" // Sudoku Classic
	PuzzleJigsaw        PuzzleType = "jigsaw"         // Jigsaw
	PuzzleWindoku       PuzzleType = "windoku"        // Windoku
	PuzzleSudokuX       PuzzleType = "sudoku_x"       // Sudoku X
	PuzzleKakuro        PuzzleType = "kakuro"         // Kakuro
)

func (t PuzzleType) String() string {
	return string(t)
}

func PuzzleTypeLess(i, j PuzzleType) bool {
	list := map[PuzzleType]int{
		PuzzleSudokuClassic: 0,
		PuzzleJigsaw:        1,
		PuzzleWindoku:       2,
		PuzzleSudokuX:       3,
		PuzzleKakuro:        4,
	}
	return list[i] < list[j]
}

type PuzzleLevel string

const (
	PuzzleLevelEasy    = PuzzleLevel("easy")
	PuzzleLevelNormal  = PuzzleLevel("normal")
	PuzzleLevelHard    = PuzzleLevel("hard")
	PuzzleLevelHarder  = PuzzleLevel("harder")
	PuzzleLevelInsane  = PuzzleLevel("insane")
	PuzzleLevelDemon   = PuzzleLevel("demon")
	PuzzleLevelCustom  = PuzzleLevel("custom")
	PuzzleLevelUnknown = PuzzleLevel("")
)

func (l PuzzleLevel) String() string {
	return string(l)
}

func PuzzleLevelLess(i, j PuzzleLevel) bool {
	list := map[PuzzleLevel]int{
		PuzzleLevelEasy:   0,
		PuzzleLevelNormal: 1,
		PuzzleLevelHard:   2,
		PuzzleLevelHarder: 3,
		PuzzleLevelInsane: 4,
		PuzzleLevelDemon:  5,
	}
	return list[i] < list[j]
}

func (l PuzzleLevel) Strategies() PuzzleStrategy {
	switch l {
	case PuzzleLevelEasy:
		return levelEasyStrategies
	case PuzzleLevelNormal:
		return levelEasyStrategies | levelNormalStrategies
	case PuzzleLevelHard:
		return levelEasyStrategies | levelNormalStrategies | levelHardStrategies
	case PuzzleLevelHarder:
		return levelEasyStrategies | levelNormalStrategies | levelHardStrategies | levelHarderStrategies
	case PuzzleLevelInsane:
		return levelEasyStrategies | levelNormalStrategies | levelHardStrategies | levelHarderStrategies |
			levelInsaneStrategies
	case PuzzleLevelDemon:
		return levelEasyStrategies | levelNormalStrategies | levelHardStrategies | levelHarderStrategies |
			levelInsaneStrategies | levelDemonStrategies
	default:
		return 0
	}
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
		return Point{}, fmt.Errorf("wrong format Point")
	}
	p := Point{}

	switch ch := s[0]; {
	case 'a' <= ch && ch <= 'z':
		p.Row = int(ch) - 'a'
	case 'A' <= ch && ch <= 'Z':
		p.Row = int(ch) - 'A'
	default:
		return Point{}, fmt.Errorf("wrong format Point")
	}

	var err error
	p.Col, err = strconv.Atoi(s[1:])
	if err != nil {
		return Point{}, errors.Wrap(err, "wrong format Point")
	}
	p.Col--

	return p, nil
}
