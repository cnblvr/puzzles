package frontend

import (
	"context"
	"fmt"
)

func init() {
	websocketPool.Add((*websocketSetUserPreferencesRequest)(nil), (*websocketSetUserPreferencesResponse)(nil))
}

type websocketSetUserPreferencesRequest struct {
	UseHighlights *bool `json:"use_highlights"`
	UseCandidates *bool `json:"use_candidates"`
}

func (websocketSetUserPreferencesRequest) Method() string {
	return "setUserPreferences"
}

func (r websocketSetUserPreferencesRequest) Validate(ctx context.Context) error {
	return nil
}

func (r websocketSetUserPreferencesRequest) Execute(ctx context.Context) (websocketResponse, error) {
	srv, session := FromContextServiceFrontendOrNil(ctx), FromContextSession(ctx)

	if session.UserID < 0 {
		return websocketSetUserPreferencesResponse{}, nil
	}

	userPreferences, err := srv.userRepository.GetUserPreferences(ctx, session.UserID)
	if err != nil {
		return websocketSetUserPreferencesResponse{}, fmt.Errorf("internal server error")
	}

	if r.UseHighlights != nil {
		userPreferences.UseHighlights = *r.UseHighlights
	}
	if r.UseCandidates != nil {
		userPreferences.UseCandidates = *r.UseCandidates
	}

	if err := srv.userRepository.SetUserPreferences(ctx, userPreferences); err != nil {
		return websocketSetUserPreferencesResponse{}, fmt.Errorf("internal server error")
	}

	return websocketSetUserPreferencesResponse{}, nil
}

// TODO handle and test
type websocketSetUserPreferencesResponse struct{}

func (websocketSetUserPreferencesResponse) Method() string {
	return "setUserPreferences"
}

func (r websocketSetUserPreferencesResponse) Validate(ctx context.Context) error {
	return nil
}

func (r websocketSetUserPreferencesResponse) Execute(ctx context.Context) error {
	return nil
}
