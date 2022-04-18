package frontend

import (
	"context"
	"fmt"
	"github.com/cnblvr/puzzles/puzzle_library"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func init() {
	websocketPool.Add((*websocketGetHintRequest)(nil), (*websocketGetHintResponse)(nil))
}

type websocketGetHintRequest struct {
	GameID uuid.UUID `json:"game_id"`
}

func (websocketGetHintRequest) Method() string {
	return "getHint"
}

func (r websocketGetHintRequest) Validate(ctx context.Context) error {
	if r.GameID == uuid.Nil {
		return errors.Errorf("game_id is empty")
	}
	return nil
}

func (r websocketGetHintRequest) Execute(ctx context.Context) (websocketResponse, error) {
	srv := FromContextServiceFrontendOrNil(ctx)

	puzzle, game, err := srv.puzzleRepository.GetPuzzleAndGame(ctx, r.GameID)
	if err != nil {
		return websocketMakeStepResponse{}, fmt.Errorf("internal server error")
	}

	statePuzzle, err := puzzle_library.GetAssistant(puzzle.Type, game.State)
	if err != nil {
		return websocketMakeStepResponse{}, fmt.Errorf("internal server error")
	}

	_, step, err := statePuzzle.SolveOneStep(game.StateCandidates, puzzle.Level.Strategies())
	if err != nil {
		return websocketMakeStepResponse{}, fmt.Errorf("internal server error")
	}

	return websocketGetHintResponse{
		Strategy: step.Strategy().String(),
	}, nil
}

// TODO handle and test
type websocketGetHintResponse struct {
	Strategy string `json:"strategy,omitempty"`
}

func (websocketGetHintResponse) Method() string {
	return "getHint"
}

func (r websocketGetHintResponse) Validate(ctx context.Context) error {
	return nil
}

func (r websocketGetHintResponse) Execute(ctx context.Context) error {
	return nil
}
