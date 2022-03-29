package main

import (
	"github.com/cnblvr/puzzles/app"
	frontend "github.com/cnblvr/puzzles/internal/frontend"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
)

func main() {
	app.InitHumanLogger()
	srv, err := frontend.NewService()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create frontend service")
	}

	r := mux.NewRouter()

	pages := r.NewRoute().Subrouter()
	pages.Use(
		srv.MiddlewareReqID,
		srv.MiddlewareLogger,
		srv.MiddlewareLogRequest,
	)

	pages.Path(app.EndpointIndex).Methods(http.MethodGet).HandlerFunc(srv.HandleHome)
	pages.Path(app.EndpointLogin).Methods(http.MethodGet, http.MethodPost).HandlerFunc(srv.HandleLogin)
	pages.Path(app.EndpointSignup).Methods(http.MethodGet, http.MethodPost).HandlerFunc(srv.HandleSignup)
	pages.Path(app.EndpointLogout).Methods(http.MethodGet).HandlerFunc(srv.HandleLogout)

	mwChainError := func(next http.Handler) http.Handler {
		return srv.MiddlewareReqID(srv.MiddlewareLogger(srv.MiddlewareLogRequest(next)))
	}
	pages.NotFoundHandler = mwChainError(http.HandlerFunc(srv.HandleNotFound))
	pages.MethodNotAllowedHandler = mwChainError(http.HandlerFunc(srv.HandleMethodNotAllowed))

	const address = "localhost:8080" // TODO env var
	log.Info().Msgf("service started on %s...", address)
	if err = http.ListenAndServe(address, r); err != nil {
		log.Fatal().Err(err).Msg("failed to listen and serve")
	}
}
