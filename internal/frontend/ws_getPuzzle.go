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
	rpl.StatePuzzle = r.game.State
	rpl.StateCandidates = json.RawMessage(r.game.StateCandidates)
	rpl.IsNew = r.game.IsNew
	rpl.IsWin = r.game.IsWin

	statePuzzle, err := srv.puzzleLibrary.GetAssistant(r.puzzle.Type, rpl.StatePuzzle)
	if err != nil {
		return nil, app.StatusBadRequest.WithError(errors.WithStack(err))
	}

	if !r.game.IsNew {
		rpl.Wrongs = statePuzzle.GetWrongPoints()
		wrongCandidates, err := statePuzzle.GetWrongCandidates(r.game.StateCandidates)
		if err != nil {
			log.Error().Err(err).Send()
			return nil, app.StatusBadRequest.WithMessage("wrong format candidates").WithError(errors.WithStack(err))
		}
		rpl.WrongsCandidates = json.RawMessage(wrongCandidates)
	}

	return rpl, nil
}

// TODO handle and test
type wsGetPuzzleReply struct {
	Puzzle string `json:"puzzle"`
	IsNew  bool   `json:"is_new,omitempty"`
	IsWin  bool   `json:"is_win,omitempty"`

	// if IsNew is false
	StatePuzzle      string          `json:"state_puzzle,omitempty"`
	StateCandidates  json.RawMessage `json:"state_candidates,omitempty"`
	Wrongs           []app.Point     `json:"wrongs,omitempty"`
	WrongsCandidates json.RawMessage `json:"wrongsCandidates,omitempty"`
}
