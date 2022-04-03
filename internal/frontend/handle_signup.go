package frontend

import (
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	"github.com/pkg/errors"
	"net/http"
)

func (srv *service) HandleSignup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, session := FromContextLogger(ctx), FromContextSession(ctx)

	var errMsg string
	if r.Method == http.MethodPost {
		var (
			msgUsernameWrong = fmt.Sprintf(
				"Username must be between %d and %d characters. Use latin letters, numbers and symbols '-._'.",
				app.MinLengthUsername, app.MaxLengthUsername,
			)
			msgPasswordWrong = fmt.Sprintf(
				"Password must be between %d and %d characters.",
				app.MinLengthPassword, app.MaxLengthPassword,
			)
			msgRepeatPasswordWrong = "Passwords don't match."
			msgUsernameNotVacant   = "Username is not vacant."
			msgInternalServerError = "Internal Server Error."
		)

		errMsg = func() string {
			username, password, repeatPassword :=
				r.PostFormValue("username"), r.PostFormValue("password"), r.PostFormValue("repeat_password")
			if err := app.ValidateUsername(username); err != nil {
				log.Debug().Err(err).Msgf("validate username")
				return msgUsernameWrong
			}
			if err := app.ValidatePassword(username); err != nil {
				log.Debug().Err(err).Msgf("validate password")
				return msgPasswordWrong
			}
			if password != repeatPassword {
				log.Debug().Msgf("passwords don't match")
				return msgRepeatPasswordWrong
			}
			salt := app.GeneratePasswordSalt()
			hash, err := srv.hashPassword(password, salt)
			if err != nil {
				log.Error().Err(err).Send()
				return msgInternalServerError
			}
			user, err := srv.userRepository.CreateUser(ctx, username, salt, hash)
			if err != nil {
				if errors.Is(err, app.ErrorUsernameIsNotVacant) {
					log.Debug().Msg("username is not vacant")
					return msgUsernameNotVacant
				}
				log.Error().Err(err).Send()
				return msgInternalServerError
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
			srv.setCookieNotificationToResponse(w, app.NotificationSuccess, "You have successfully signed up.")
			http.Redirect(w, r, app.EndpointHome, http.StatusSeeOther)
			return ""
		}()
		if errMsg == "" {
			return
		}
	}

	srv.executeTemplate(ctx, w, templates.PageSignup, func(params *templates.Params) {
		params.Header.Title = "Signup"
		params.Data = struct {
			ErrorMessage string
		}{
			ErrorMessage: errMsg,
		}
	})
}
