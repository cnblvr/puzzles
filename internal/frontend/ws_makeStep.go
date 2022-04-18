package frontend

import (
	"context"
	"encoding/json"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func init() {
	wsAddIncoming("makeStep", (*wsMakeStepRequest)(nil))
}

type wsMakeStepRequest struct {
	wsGameMiddleware
	Step app.PuzzleUserStep `json:"step"`
}

func (r *wsMakeStepRequest) Validate(ctx context.Context) app.Status {
	if err := r.Step.Type.Validate(); err != nil {
		return app.StatusBadRequest.WithMessage("invalid .step").WithError(errors.WithStack(err))
	}
	return nil
}

func (r *wsMakeStepRequest) Execute(ctx context.Context) (wsIncomingReply, app.Status) {
	rpl := new(wsMakeStepReply)
	srv := FromContextServiceFrontendOrNil(ctx)

	statePuzzle, err := srv.puzzleLibrary.GetAssistant(r.puzzle.Type, r.game.State)
	if err != nil {
		return nil, app.StatusInternalServerError.WithError(errors.WithStack(err))
	}

	newStateCandidates, wrongCandidates, err := statePuzzle.MakeUserStep(r.game.StateCandidates, r.Step)
	if err != nil {
		return nil, app.StatusUnknown.WithMessage("failed to make step").WithError(errors.WithStack(err))
	}
	r.game.IsNew = false
	r.game.State, r.game.StateCandidates = statePuzzle.String(), newStateCandidates

	defer func() {
		if err := srv.puzzleRepository.UpdatePuzzleGame(ctx, r.game); err != nil {
			log.Error().Err(err).Msg("failed to update puzzle game")
		}
	}()

	if r.game.State == r.puzzle.Solution {
		// WIN
		r.game.IsWin = true
		rpl.Win = true
		return rpl, nil
	}

	rpl.Errors = statePuzzle.GetWrongPoints()
	rpl.ErrorsCandidates = json.RawMessage(wrongCandidates)

	return rpl, nil
}

// TODO handle and test
type wsMakeStepReply struct {
	Errors           []app.Point     `json:"errors,omitempty"`
	ErrorsCandidates json.RawMessage `json:"errorsCandidates,omitempty"`
	Win              bool            `json:"win,omitempty"`
}
