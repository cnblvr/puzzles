package app

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
)

const (
	EndpointInternalServerError = "/error"
	EndpointHome                = "/"
	EndpointLogin               = "/login"
	EndpointSignup              = "/signup"
	EndpointLogout              = "/logout"
	EndpointSettings            = "/settings"
	endpointGameIDPattern       = "/game/%s"
	EndpointGameWs              = "/game_ws"
)

type EndpointGameID struct{}

func (EndpointGameID) Path(gameID uuid.UUID) string {
	return fmt.Sprintf(endpointGameIDPattern, gameID.String())
}

func (EndpointGameID) MuxPath() string {
	return fmt.Sprintf(endpointGameIDPattern, "{game_id}")
}

func (EndpointGameID) MuxParse(r *http.Request) (uuid.UUID, error) {
	gameIDStr, ok := mux.Vars(r)["game_id"]
	if !ok {
		return uuid.UUID{}, errors.Errorf("game_id not found")
	}
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		return uuid.UUID{}, errors.Wrap(err, "game_id is not uuid")
	}
	return gameID, nil
}
