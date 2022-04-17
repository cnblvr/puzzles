package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/puzzle_library"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func init() {
	websocketPool.Add((*websocketMakeStepRequest)(nil), (*websocketMakeStepResponse)(nil))
}

type websocketMakeStepRequest struct {
	GameID uuid.UUID          `json:"game_id"`
	Step   app.PuzzleUserStep `json:"step"`
}

func (websocketMakeStepRequest) Method() string {
	return "makeStep"
}

func (r websocketMakeStepRequest) Validate(ctx context.Context) error {
	if r.GameID == uuid.Nil {
		return errors.Errorf("game_id is empty")
	}
	if err := r.Step.Type.Validate(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r websocketMakeStepRequest) Execute(ctx context.Context) (websocketResponse, error) {
	srv := FromContextServiceFrontendOrNil(ctx)
	resp := websocketMakeStepResponse{}

	puzzle, game, err := srv.puzzleRepository.GetPuzzleAndGame(ctx, r.GameID)
	if err != nil {
		return websocketMakeStepResponse{}, fmt.Errorf("internal server error")
	}

	statePuzzle, err := puzzle_library.GetAssistant(puzzle.Type, game.State)
	if err != nil {
		return websocketMakeStepResponse{}, fmt.Errorf("internal server error")
	}

	newStateCandidates, wrongCandidates, err := statePuzzle.MakeUserStep(game.StateCandidates, r.Step)
	if err != nil {
		return websocketMakeStepResponse{}, errors.Wrap(err, "failed to make step")
	}
	game.IsNew = false
	game.State, game.StateCandidates = statePuzzle.String(), newStateCandidates

	defer func() {
		if err := srv.puzzleRepository.UpdatePuzzleGame(ctx, game); err != nil {
			log.Error().Err(err).Msg("failed to update puzzle game")
		}
	}()

	if game.State == puzzle.Solution {
		// WIN
		game.IsWin = true
		return websocketMakeStepResponse{
			Win: true,
		}, nil
	}

	resp.Errors = statePuzzle.GetWrongPoints()

	resp.ErrorsCandidates = json.RawMessage(wrongCandidates)

	return resp, nil
}

// TODO handle and test
type websocketMakeStepResponse struct {
	Errors           []app.Point     `json:"errors,omitempty"`
	ErrorsCandidates json.RawMessage `json:"errorsCandidates,omitempty"`
	Win              bool            `json:"win,omitempty"`
}

func (websocketMakeStepResponse) Method() string {
	return "makeStep"
}

func (r websocketMakeStepResponse) Validate(ctx context.Context) error {
	return nil
}

func (r websocketMakeStepResponse) Execute(ctx context.Context) error {
	return nil
}
