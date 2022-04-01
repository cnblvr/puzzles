package main

import (
	"github.com/cnblvr/puzzles/app"
	frontend "github.com/cnblvr/puzzles/internal/frontend"
	"github.com/cnblvr/puzzles/internal/frontend/static"
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

	// Router for static files: JS, CSS, images.
	rStatic := r.NewRoute().Subrouter()
	rStatic.Path("/favicon.ico").Methods(http.MethodGet).Handler(http.FileServer(http.FS(static.Favicon)))
	rStatic.PathPrefix("/css").Methods(http.MethodGet).Handler(http.FileServer(http.FS(static.CSS)))
	rStatic.PathPrefix("/js").Methods(http.MethodGet).Handler(http.FileServer(http.FS(static.JS)))

	pages := r.NewRoute().Subrouter()
	pages.Use(
		srv.MiddlewareReqID,
		srv.MiddlewareLogger,
		srv.MiddlewareLogRequest,
		srv.MiddlewareCookieSession,
		srv.MiddlewareCookieNotification,
	)
	logoutPage := r.NewRoute().Subrouter()
	logoutPage.Use(
		srv.MiddlewareReqID,
		srv.MiddlewareLogger,
		srv.MiddlewareLogRequest,
	)
	authPages := pages.NewRoute().Subrouter()
	authPages.Use(
		srv.MiddlewareMustBeLogged,
	)

	pages.Path(app.EndpointHome).Methods(http.MethodGet).HandlerFunc(srv.HandleHome)
	pages.Path(app.EndpointLogin).Methods(http.MethodGet, http.MethodPost).HandlerFunc(srv.HandleLogin)
	pages.Path(app.EndpointSignup).Methods(http.MethodGet, http.MethodPost).HandlerFunc(srv.HandleSignup)
	logoutPage.Path(app.EndpointLogout).Methods(http.MethodGet).HandlerFunc(srv.HandleLogout)
	authPages.Path(app.EndpointSettings).Methods(http.MethodGet, http.MethodPost).HandlerFunc(srv.HandleSettings)

	mwChainError := func(next http.Handler) http.Handler {
		return srv.MiddlewareReqID(srv.MiddlewareLogger(srv.MiddlewareLogRequest(next)))
	}
	r.NotFoundHandler = mwChainError(http.HandlerFunc(srv.HandleNotFound))
	r.MethodNotAllowedHandler = mwChainError(http.HandlerFunc(srv.HandleMethodNotAllowed))
	r.Path(app.EndpointInternalServerError).Methods(http.MethodGet).HandlerFunc(srv.HandleInternalServerError)

	const address = ":8080" // TODO env var
	log.Info().Msgf("service started on %s...", address)
	if err = http.ListenAndServe(address, r); err != nil {
		log.Fatal().Err(err).Msg("failed to listen and serve")
	}
}
