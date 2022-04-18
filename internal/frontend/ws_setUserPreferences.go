package frontend

import (
	"context"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
)

func init() {
	wsAddIncoming("setUserPreferences", (*wsSetUserPreferencesRequest)(nil))
}

type wsSetUserPreferencesRequest struct {
	UseHighlights *bool `json:"use_highlights"`
	UseCandidates *bool `json:"use_candidates"`
}

func (r *wsSetUserPreferencesRequest) Validate(ctx context.Context) app.Status {
	return nil
}

func (r *wsSetUserPreferencesRequest) Execute(ctx context.Context) (wsIncomingReply, app.Status) {
	rpl := new(wsSetUserPreferencesReply)
	srv, session := FromContextServiceFrontendOrNil(ctx), FromContextSession(ctx)

	if session.UserID < 0 {
		return nil, app.StatusUnauthorized
	}

	userPreferences, err := srv.userRepository.GetUserPreferences(ctx, session.UserID)
	if err != nil {
		return nil, app.StatusInternalServerError.WithError(errors.WithStack(err))
	}

	if r.UseHighlights != nil {
		userPreferences.UseHighlights = *r.UseHighlights
	}
	if r.UseCandidates != nil {
		userPreferences.UseCandidates = *r.UseCandidates
	}

	if err := srv.userRepository.SetUserPreferences(ctx, userPreferences); err != nil {
		return nil, app.StatusInternalServerError.WithError(errors.WithStack(err))
	}

	return rpl, nil
}

// TODO handle and test
type wsSetUserPreferencesReply struct{}
