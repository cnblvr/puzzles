package frontend

import (
	"context"
	"github.com/cnblvr/puzzles/app"
	"github.com/google/uuid"
	"reflect"
	"testing"
)

func TestWsGameMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		req              wsGameMiddleware
		getPuzzleAndGame func(ctx context.Context, id uuid.UUID) (*app.Puzzle, *app.PuzzleGame, error)
		want             wsGameMiddleware
		wantSts          app.Status
	}{
		{
			name:    "empty",
			req:     wsGameMiddleware{GameID: uuid.UUID{}},
			wantSts: app.StatusBadRequest,
		},
		{
			name: "not found",
			req:  wsGameMiddleware{GameID: uuid.MustParse("11112222-3333-4444-5555-666677778888")},
			getPuzzleAndGame: func(ctx context.Context, id uuid.UUID) (*app.Puzzle, *app.PuzzleGame, error) {
				return nil, nil, app.ErrorPuzzleGameNotFound
			},
			wantSts: app.StatusBadRequest,
		},
		{
			name: "found",
			req:  wsGameMiddleware{GameID: uuid.MustParse("11112222-3333-4444-5555-666677778888")},
			getPuzzleAndGame: func(ctx context.Context, id uuid.UUID) (*app.Puzzle, *app.PuzzleGame, error) {
				return &app.Puzzle{ID: 1}, &app.PuzzleGame{ID: uuid.MustParse("11112222-3333-4444-5555-666677778888")}, nil
			},
			want: wsGameMiddleware{
				puzzle: &app.Puzzle{ID: 1},
				game:   &app.PuzzleGame{ID: uuid.MustParse("11112222-3333-4444-5555-666677778888")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := mockService(mockPuzzleRepository{
				getPuzzleAndGame: tt.getPuzzleAndGame,
			}, mockPuzzleLibrary{})
			if !checkStatus(t, "GameMiddleware", tt.req.GameMiddleware(ctx), tt.wantSts) {
				return
			}
			if !reflect.DeepEqual(tt.req.puzzle, tt.want.puzzle) {
				t.Errorf("GameMiddleware() got puzzle = %+v, want = %+v",
					tt.req.puzzle, tt.want.puzzle)
				return
			}
			if !reflect.DeepEqual(tt.req.game, tt.want.game) {
				t.Errorf("GameMiddleware() got game = %+v, want = %+v",
					tt.req.game, tt.want.game)
				return
			}
		})
	}
}

func checkStatus(t *testing.T, fn string, got app.Status, want app.Status) bool {
	if got != nil {
		if got.GetCode() != want.GetCode() {
			t.Errorf("%s() got status code = %d (%v), want = %d", fn,
				got.GetCode(), got.GetError(), want.GetCode())
			return false
		}
	} else {
		if want != nil {
			t.Errorf("%s() got status = %v, want = %d", fn,
				got, want.GetCode())
			return false
		}
	}
	return true
}

func mockWsGameMiddleware() wsGameMiddleware {
	return wsGameMiddleware{GameID: uuid.MustParse("11112222-3333-4444-5555-666677778888")}
}

func mockGetPuzzleAndGame() func(ctx context.Context, id uuid.UUID) (*app.Puzzle, *app.PuzzleGame, error) {
	return func(ctx context.Context, id uuid.UUID) (*app.Puzzle, *app.PuzzleGame, error) {
		return &app.Puzzle{ID: 1}, &app.PuzzleGame{ID: uuid.MustParse("11112222-3333-4444-5555-666677778888")}, nil
	}
}

func mockService(puzzleRepository mockPuzzleRepository, puzzleLibrary mockPuzzleLibrary) context.Context {
	return context.WithValue(context.Background(), "service_frontend", &service{
		puzzleRepository: puzzleRepository,
		puzzleLibrary:    puzzleLibrary,
	})
}

type mockPuzzleRepository struct {
	createRandomPuzzleGame func(ctx context.Context, params app.CreateRandomPuzzleGameParams) (*app.Puzzle, *app.PuzzleGame, error)
	getPuzzleGame          func(ctx context.Context, id uuid.UUID) (*app.PuzzleGame, error)
	updatePuzzleGame       func(ctx context.Context, game *app.PuzzleGame) error
	createPuzzle           func(ctx context.Context, params app.CreatePuzzleParams) (*app.Puzzle, error)
	getPuzzle              func(ctx context.Context, id int64) (*app.Puzzle, error)
	getPuzzleByGameID      func(ctx context.Context, gameID uuid.UUID) (*app.Puzzle, error)
	getPuzzleAndGame       func(ctx context.Context, id uuid.UUID) (*app.Puzzle, *app.PuzzleGame, error)
}

func (m mockPuzzleRepository) CreateRandomPuzzleGame(ctx context.Context, params app.CreateRandomPuzzleGameParams) (*app.Puzzle, *app.PuzzleGame, error) {
	if m.createRandomPuzzleGame != nil {
		return m.createRandomPuzzleGame(ctx, params)
	}
	panic("not implemented")
}

func (m mockPuzzleRepository) GetPuzzleGame(ctx context.Context, id uuid.UUID) (*app.PuzzleGame, error) {
	if m.getPuzzleGame != nil {
		return m.getPuzzleGame(ctx, id)
	}
	panic("not implemented")
}

func (m mockPuzzleRepository) UpdatePuzzleGame(ctx context.Context, game *app.PuzzleGame) error {
	if m.updatePuzzleGame != nil {
		return m.updatePuzzleGame(ctx, game)
	}
	panic("not implemented")
}

func (m mockPuzzleRepository) CreatePuzzle(ctx context.Context, params app.CreatePuzzleParams) (*app.Puzzle, error) {
	if m.createPuzzle != nil {
		return m.createPuzzle(ctx, params)
	}
	panic("not implemented")
}

func (m mockPuzzleRepository) GetPuzzle(ctx context.Context, id int64) (*app.Puzzle, error) {
	if m.getPuzzle != nil {
		return m.getPuzzle(ctx, id)
	}
	panic("not implemented")
}

func (m mockPuzzleRepository) GetPuzzleByGameID(ctx context.Context, gameID uuid.UUID) (*app.Puzzle, error) {
	if m.getPuzzleByGameID != nil {
		return m.getPuzzleByGameID(ctx, gameID)
	}
	panic("not implemented")
}

func (m mockPuzzleRepository) GetPuzzleAndGame(ctx context.Context, id uuid.UUID) (*app.Puzzle, *app.PuzzleGame, error) {
	if m.getPuzzleAndGame != nil {
		return m.getPuzzleAndGame(ctx, id)
	}
	panic("not implemented")
}

type mockPuzzleLibrary struct {
	getCreator   func(typ app.PuzzleType) (app.PuzzleCreator, error)
	getGenerator func(typ app.PuzzleType, puzzle string) (app.PuzzleGenerator, error)
	getAssistant func(typ app.PuzzleType, puzzle string) (app.PuzzleAssistant, error)
}

func (m mockPuzzleLibrary) GetCreator(typ app.PuzzleType) (app.PuzzleCreator, error) {
	if m.getCreator != nil {
		return m.getCreator(typ)
	}
	panic("not implemented")
}

func (m mockPuzzleLibrary) GetGenerator(typ app.PuzzleType, puzzle string) (app.PuzzleGenerator, error) {
	if m.getGenerator != nil {
		return m.getGenerator(typ, puzzle)
	}
	panic("not implemented")
}

func (m mockPuzzleLibrary) GetAssistant(typ app.PuzzleType, puzzle string) (app.PuzzleAssistant, error) {
	if m.getAssistant != nil {
		return m.getAssistant(typ, puzzle)
	}
	panic("not implemented")
}

type mockPuzzleAssistant struct {
	string             func() string
	typeFunc           func() app.PuzzleType
	getWrongPoints     func() []app.Point
	getWrongCandidates func(candidates string) (string, error)
	makeUserStep       func(candidatesIn string, step app.PuzzleUserStep) (candidatesOut string, wrongCandidates string, err error)
	solveOneStep       func(candidatesIn string, strategies app.PuzzleStrategy) (candidatesChanges string, step app.PuzzleStep, err error)
}

func (m mockPuzzleAssistant) String() string {
	if m.string != nil {
		return m.string()
	}
	panic("not implemented")
}

func (m mockPuzzleAssistant) Type() app.PuzzleType {
	if m.typeFunc != nil {
		return m.typeFunc()
	}
	panic("not implemented")
}

func (m mockPuzzleAssistant) GetWrongPoints() []app.Point {
	if m.getWrongPoints != nil {
		return m.getWrongPoints()
	}
	panic("not implemented")
}

func (m mockPuzzleAssistant) GetWrongCandidates(candidates string) (string, error) {
	if m.getWrongCandidates != nil {
		return m.getWrongCandidates(candidates)
	}
	panic("not implemented")
}
func (m mockPuzzleAssistant) MakeUserStep(candidatesIn string, step app.PuzzleUserStep) (candidatesOut string, wrongCandidates string, err error) {
	if m.makeUserStep != nil {
		return m.makeUserStep(candidatesIn, step)
	}
	panic("not implemented")
}
func (m mockPuzzleAssistant) SolveOneStep(candidatesIn string, strategies app.PuzzleStrategy) (candidatesChanges string, step app.PuzzleStep, err error) {
	if m.solveOneStep != nil {
		return m.solveOneStep(candidatesIn, strategies)
	}
	panic("not implemented")
}

type mockPuzzleStep struct {
	strategy         func() app.PuzzleStrategy
	candidateChanges func() string
	description      func() string
}

func (m mockPuzzleStep) Strategy() app.PuzzleStrategy {
	if m.strategy != nil {
		return m.strategy()
	}
	panic("not implemented")
}

func (m mockPuzzleStep) CandidateChanges() string {
	if m.candidateChanges != nil {
		return m.candidateChanges()
	}
	panic("not implemented")
}
func (m mockPuzzleStep) Description() string {
	if m.description != nil {
		return m.description()
	}
	panic("not implemented")
}
