package frontend

import (
	"github.com/cnblvr/puzzles/app"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (srv *service) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	success := true
	cookieSession, err := srv.getCookieSessionFromRequest(r)
	if err != nil {
		log.Debug().Err(err).Msg("failed to get cookie session from request")
		success = false
	} else {
		if err := srv.userRepository.DeleteSession(ctx, cookieSession.SessionID); err != nil {
			log.Debug().Err(err).Msg("failed to delete session")
			success = false
		}
	}
	srv.deleteCookieSession(w)
	if success {
		if err := srv.setCookieNotificationToResponse(w, &app.CookieNotification{
			Type:    app.NotificationSuccess,
			Message: "You have successfully logged out.",
		}); err != nil {
			log.Error().Err(err).Msg("failed to set cookie notification")
		}
	}
	http.Redirect(w, r, app.EndpointHome, http.StatusSeeOther)
}
