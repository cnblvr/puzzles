package frontend

import (
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
	"net/http"
	"time"
)

func (srv *service) MiddlewareReqID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = NewContextReqID(ctx, app.GenerateReqID())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (srv *service) MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		loggerBuilder := zlog.Logger.With()
		if reqID := FromContextReqID(ctx); reqID != "" {
			loggerBuilder = loggerBuilder.Str("req_id", reqID)
		}
		ctx = NewContextLogger(ctx, loggerBuilder.Logger())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (srv *service) MiddlewareLogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := FromContextLogger(ctx)
		log.Info().
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Send()
		next.ServeHTTP(w, r)
	})
}

func (srv *service) MiddlewareCookieSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		internalServerError := func() {
			http.Redirect(w, r, app.EndpointInternalServerError, http.StatusSeeOther)
		}
		logout := func() {
			http.Redirect(w, r, app.EndpointLogout, http.StatusSeeOther)
		}
		ctx := r.Context()
		log := FromContextLogger(ctx)

		var session *app.Session
		cookieSession, err := srv.getCookieSessionFromRequest(r)
		if err != nil {
			log.Debug().Err(err).Msg("failed to get cookie session from request")
			session, err = srv.userRepository.CreateSession(ctx, 0, app.DefaultCookieSessionExpiration)
			if err != nil {
				log.Error().Err(err).Msg("failed to create new session")
				internalServerError()
				return
			}
		} else {
			session, err = srv.userRepository.GetSession(ctx, cookieSession.SessionID)
			if err != nil {
				if errors.Is(err, app.ErrorSessionNotFound) {
					log.Error().Err(err).Msg("session has expired")
					srv.setCookieNotificationToResponse(w, &app.CookieNotification{
						Type:    app.NotificationWarning,
						Message: "Your session has expired.",
					})
					logout()
					return
				}
				log.Error().Err(err).Msg("failed to get session")
				internalServerError()
				return
			}
			if err := session.ValidateWith(cookieSession); err != nil {
				log.Error().Err(err).Msg("session is hacked")
				logout()
				return
			}
		}
		if err := srv.setCookieSessionToResponse(w, session); err != nil {
			log.Error().Err(err).Msg("failed to set cookie session to response")
			internalServerError()
			return
		}
		log.Debug().Int64("user_id", session.UserID).Int64("session_id", session.SessionID).Msg("session verified")

		ctx = NewContextSession(ctx, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (srv *service) MiddlewareCookieNotification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := FromContextLogger(ctx)

		notification, err := srv.getCookieNotificationFromRequest(r)
		if err != nil {
			if !errors.Is(err, http.ErrNoCookie) {
				log.Error().Err(err).Msg("failed to get cookie notification from request")
				srv.deleteCookieNotification(w)
			}
		} else {
			srv.deleteCookieNotification(w)
		}

		ctx = NewContextNotification(ctx, notification)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (srv *service) MiddlewareMustBeLogged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session, log := FromContextSession(ctx), FromContextLogger(ctx)

		if session.UserID <= 0 {
			log.Debug().Msgf("user is not logged")
			srv.setCookieNotificationToResponse(w, &app.CookieNotification{
				Type:    app.NotificationError,
				Message: fmt.Sprintf("The page %s is not allowed for you.", r.URL.Path),
			})
			http.Redirect(w, r, app.EndpointHome, http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// session cookie

const cookieSessionName = "session"

func (srv *service) getCookieSessionFromRequest(r *http.Request) (*app.CookieSession, error) {
	cookie, err := r.Cookie(cookieSessionName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get '%s' cookie", cookieSessionName)
	}
	cookieSession := &app.CookieSession{}
	if err := srv.secCookie.Decode(cookieSessionName, cookie.Value, cookieSession); err != nil {
		return nil, errors.Wrapf(err, "failed to decode '%s' cookie", cookieSessionName)
	}
	return cookieSession, nil
}

func (srv *service) setCookieSessionToResponse(w http.ResponseWriter, session *app.Session) error {
	cookieValue, err := srv.secCookie.Encode(cookieSessionName, session.CookieSession)
	if err != nil {
		return errors.Wrapf(err, "failed to encode '%s' cookie", cookieSessionName)
	}
	srv.setCookie(w, cookieSessionName, cookieValue, time.Duration(session.ExpirationSeconds)*time.Second)
	return nil
}

func (srv *service) deleteCookieSession(w http.ResponseWriter) {
	srv.deleteCookie(w, cookieSessionName)
}

// notification cookie

const cookieNotificationName = "notification"

func (srv *service) getCookieNotificationFromRequest(r *http.Request) (*app.CookieNotification, error) {
	cookie, err := r.Cookie(cookieNotificationName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get '%s' cookie", cookieNotificationName)
	}
	notification := &app.CookieNotification{}
	if err := srv.secCookie.Decode(cookieNotificationName, cookie.Value, notification); err != nil {
		return nil, errors.Wrapf(err, "failed to decode '%s' cookie", cookieNotificationName)
	}
	return notification, nil
}

func (srv *service) setCookieNotificationToResponse(w http.ResponseWriter, notification *app.CookieNotification) error {
	cookieValue, err := srv.secCookie.Encode(cookieNotificationName, notification)
	if err != nil {
		return errors.Wrapf(err, "failed to encode '%s' cookie", cookieNotificationName)
	}
	srv.setCookie(w, cookieNotificationName, cookieValue, time.Minute)
	return nil
}

func (srv *service) deleteCookieNotification(w http.ResponseWriter) {
	srv.deleteCookie(w, cookieNotificationName)
}

// common cookie

func (srv *service) setCookie(w http.ResponseWriter, name string, value string, expiration time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(expiration),
		Secure:   false,
		HttpOnly: true,
	})
}

func (srv *service) deleteCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		Secure:   false,
		HttpOnly: true,
	})
}
