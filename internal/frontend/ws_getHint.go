package frontend

import (
	"context"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
)

func init() {
	wsAddIncoming("getHint", (*wsGetHintRequest)(nil))
}

type wsGetHintRequest struct {
	wsGameMiddleware
}

func (r *wsGetHintRequest) Validate(ctx context.Context) app.Status {
	return nil
}

func (r *wsGetHintRequest) Execute(ctx context.Context) (wsIncomingReply, app.Status) {
	rpl := new(wsGetHintReply)
	srv := FromContextServiceFrontendOrNil(ctx)

	statePuzzle, err := srv.puzzleLibrary.GetAssistant(r.puzzle.Type, r.game.State)
	if err != nil {
		return nil, app.StatusInternalServerError.WithError(errors.WithStack(err))
	}

	_, step, err := statePuzzle.SolveOneStep(r.game.StateCandidates, r.puzzle.Level.Strategies())
	if err != nil {
		return nil, app.StatusUnknown.WithMessage("TODO").WithError(errors.New("TODO failed to solve"))
	}
	rpl.Strategy = step.Strategy().String()

	return rpl, nil
}

// TODO handle and test
type wsGetHintReply struct {
	Strategy string `json:"strategy,omitempty"`
}
