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
		log.Warn().Err(err).Msg("failed to get cookie session from request")
		success = false
	} else {
		if err := srv.userRepository.DeleteSession(ctx, cookieSession.SessionID); err != nil {
			log.Warn().Err(err).Msg("failed to delete session")
			success = false
		}
	}
	srv.deleteCookieSession(w)
	if success {
		log.Info().Int64("user_id", cookieSession.UserID).Msg("logged out")
		srv.setCookieNotificationToResponse(w, app.NotificationSuccess, "You have successfully logged out.")
	}
	http.Redirect(w, r, app.EndpointHome, http.StatusSeeOther)
}
