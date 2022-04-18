package frontend

import (
	"context"
	"github.com/cnblvr/puzzles/app"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"reflect"
	"testing"
)

func TestWsGetHint(t *testing.T) {
	tests := []struct {
		name             string
		req              wsGetHintRequest
		getPuzzleAndGame func(ctx context.Context, id uuid.UUID) (*app.Puzzle, *app.PuzzleGame, error)
		getAssistant     func(typ app.PuzzleType, puzzle string) (app.PuzzleAssistant, error)
		wantRpl          wsIncomingReply
		wantSts          app.Status
	}{
		{
			name:             "unknown puzzle type",
			req:              wsGetHintRequest{mockWsGameMiddleware()},
			getPuzzleAndGame: mockGetPuzzleAndGame(),
			getAssistant: func(typ app.PuzzleType, puzzle string) (app.PuzzleAssistant, error) {
				return nil, app.ErrorPuzzleTypeUnknown
			},
			wantSts: app.StatusBadRequest,
		},
		{
			name:             "solve one step error",
			req:              wsGetHintRequest{mockWsGameMiddleware()},
			getPuzzleAndGame: mockGetPuzzleAndGame(),
			getAssistant: func(typ app.PuzzleType, puzzle string) (app.PuzzleAssistant, error) {
				return mockPuzzleAssistant{
					solveOneStep: func(candidatesIn string, strategies app.PuzzleStrategy) (candidatesChanges string, step app.PuzzleStep, err error) {
						return "", nil, errors.Errorf("any error")
					},
				}, nil
			},
			wantSts: app.StatusUnknown,
		},
		{
			name:             "success",
			req:              wsGetHintRequest{mockWsGameMiddleware()},
			getPuzzleAndGame: mockGetPuzzleAndGame(),
			getAssistant: func(typ app.PuzzleType, puzzle string) (app.PuzzleAssistant, error) {
				return mockPuzzleAssistant{
					solveOneStep: func(candidatesIn string, strategies app.PuzzleStrategy) (candidatesChanges string, step app.PuzzleStep, err error) {
						return "", mockPuzzleStep{
							strategy: func() app.PuzzleStrategy {
								return app.StrategyNakedSingle
							},
						}, nil
					},
				}, nil
			},
			wantRpl: &wsGetHintReply{
				Strategy: app.StrategyNakedSingle.String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := mockService(mockPuzzleRepository{
				getPuzzleAndGame: tt.getPuzzleAndGame,
			}, mockPuzzleLibrary{
				getAssistant: tt.getAssistant,
			})
			if !checkStatus(t, "wsGetHint.GameMiddleware", tt.req.GameMiddleware(ctx), nil) {
				return
			}
			if !checkStatus(t, "wsGetHint.Validate", tt.req.Validate(ctx), nil) {
				return
			}
			rpl, status := tt.req.Execute(ctx)
			if !checkStatus(t, "wsGetHint.Execute", status, tt.wantSts) {
				return
			}
			if !reflect.DeepEqual(rpl, tt.wantRpl) {
				t.Errorf("wsGetHint.Execute() got = %+v, want = %+v",
					rpl, tt.wantRpl)
				return
			}
		})
	}
}
