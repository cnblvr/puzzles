package app

import (
	"fmt"
	"github.com/google/uuid"
)

const (
	EndpointInternalServerError = "/error"
	EndpointHome                = "/"
	EndpointLogin               = "/login"
	EndpointSignup              = "/signup"
	EndpointLogout              = "/logout"
	EndpointSettings            = "/settings"
	endpointGameIDPattern       = "/game/%s"
)

func EndpointGameID(gameID uuid.UUID) string {
	return fmt.Sprintf(endpointGameIDPattern, gameID.String())
}
