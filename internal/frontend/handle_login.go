package frontend

import (
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	"github.com/pkg/errors"
	"net/http"
)

func (srv *service) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, session := FromContextLogger(ctx), FromContextSession(ctx)

	var errMsg string
	if r.Method == http.MethodPost {
		var (
			msgUsernamePasswordWrong = "Username or password invalid."
			msgInternalServerError   = "Internal Server Error."
		)

		errMsg = func() string {
			username, password := r.PostFormValue("username"), r.PostFormValue("password")
			if username == "" {
				log.Warn().Msg("username is empty")
				return msgUsernamePasswordWrong
			}
			user, err := srv.userRepository.GetUserByUsername(ctx, username)
			if err != nil {
				if errors.Is(err, app.ErrorUserNotFound) {
					log.Warn().Err(err).Send()
					return msgUsernamePasswordWrong
				}
				log.Error().Err(err).Send()
				return msgInternalServerError
			}
			verified, err := srv.verifyPassword(password, user.Salt, user.Hash)
			if err != nil {
				log.Error().Err(err).Send()
				return msgInternalServerError
			}
			if !verified {
				log.Warn().Msg("password is not valid")
				return msgUsernamePasswordWrong
			}
			session.UserID = user.ID
			if err := srv.userRepository.UpdateSession(ctx, session); err != nil {
				log.Error().Err(err).Send()
				return msgInternalServerError
			}
			if err := srv.setCookieSessionToResponse(w, session); err != nil {
				log.Error().Err(err).Send()
				return msgInternalServerError
			}
			srv.setCookieNotificationToResponse(w, app.NotificationSuccess, "You have successfully logged in.")
			http.Redirect(w, r, app.EndpointHome, http.StatusSeeOther)
			log.Info().Int64("user_id", session.UserID).Msg("logged in")
			return ""
		}()
		if errMsg == "" {
			return
		}
	}

	srv.executeTemplate(ctx, w, templates.PageLogin, func(params *templates.Params) {
		params.Header.Title = "Login"
		params.Data = struct {
			ErrorMessage string
		}{
			ErrorMessage: errMsg,
		}
	})
}
