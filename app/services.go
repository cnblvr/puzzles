package app

import "net/http"

type ServiceFrontend interface {
	MiddlewareReqID(next http.Handler) http.Handler
	MiddlewareLogger(next http.Handler) http.Handler
	MiddlewareLogRequest(next http.Handler) http.Handler
	MiddlewareCookieSession(next http.Handler) http.Handler
	MiddlewareCookieNotification(next http.Handler) http.Handler
	MiddlewareMustBeLogged(next http.Handler) http.Handler

	HandleHome(w http.ResponseWriter, r *http.Request)
	HandleNotFound(w http.ResponseWriter, r *http.Request)
	HandleMethodNotAllowed(w http.ResponseWriter, r *http.Request)
	HandleInternalServerError(w http.ResponseWriter, r *http.Request)
	HandleLogin(w http.ResponseWriter, r *http.Request)
	HandleSignup(w http.ResponseWriter, r *http.Request)
	HandleLogout(w http.ResponseWriter, r *http.Request)
	HandleSettings(w http.ResponseWriter, r *http.Request)
}

type ServiceGenerator interface {
	Run() error
}
