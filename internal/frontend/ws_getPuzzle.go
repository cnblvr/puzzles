package frontend

import (
	"context"
	"encoding/json"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
)

func init() {
	wsAddIncoming("getPuzzle", (*wsGetPuzzleRequest)(nil))
}

type wsGetPuzzleRequest struct {
	wsGameMiddleware
}

func (r *wsGetPuzzleRequest) Validate(ctx context.Context) app.Status {
	return nil
}

func (r *wsGetPuzzleRequest) Execute(ctx context.Context) (wsIncomingReply, app.Status) {
	rpl := new(wsGetPuzzleReply)
	srv, log := FromContextServiceFrontendOrNil(ctx), FromContextLogger(ctx)

	rpl.Puzzle = r.puzzle.Clues
	rpl.Candidates = json.RawMessage("{}")
	rpl.IsNew = r.game.IsNew
	rpl.IsWin = r.game.IsWin
	if r.game.CandidatesAtStart {
		rpl.Candidates = json.RawMessage(r.puzzle.Candidates)
	}

	statePuzzle, err := srv.puzzleLibrary.GetAssistant(r.puzzle.Type, rpl.Puzzle)
	if err != nil {
		return nil, app.StatusBadRequest.WithError(errors.WithStack(err))
	}

	if r.game.IsNew {
		rpl.StateCandidates = rpl.Candidates
		r.game.State = r.puzzle.Clues
		r.game.StateCandidates = string(rpl.StateCandidates)
		if err := srv.puzzleRepository.UpdatePuzzleGame(ctx, r.game); err != nil {
			return nil, app.StatusInternalServerError.WithError(errors.WithStack(err))
		}
	} else {
		rpl.StatePuzzle = r.game.State
		rpl.StateCandidates = json.RawMessage(r.game.StateCandidates)
		rpl.Errors = statePuzzle.GetWrongPoints()
		wrongCandidates, err := statePuzzle.GetWrongCandidates(r.game.StateCandidates)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, app.StatusBadRequest.WithMessage("wrong format candidates").WithError(errors.WithStack(err))
		}
		rpl.ErrorsCandidates = json.RawMessage(wrongCandidates)
	}

	return rpl, nil
}

// TODO handle and test
type wsGetPuzzleReply struct {
	Puzzle     string          `json:"puzzle"`
	Candidates json.RawMessage `json:"candidates,omitempty"`
	IsNew      bool            `json:"is_new,omitempty"`
	IsWin      bool            `json:"is_win,omitempty"`

	// if IsNew is false
	StatePuzzle      string          `json:"state_puzzle,omitempty"`
	StateCandidates  json.RawMessage `json:"state_candidates,omitempty"`
	Errors           []app.Point     `json:"errors,omitempty"`
	ErrorsCandidates json.RawMessage `json:"errorsCandidates,omitempty"`
}
