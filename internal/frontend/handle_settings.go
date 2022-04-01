package frontend

import (
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	"github.com/google/uuid"
	"net/http"
	"sort"
	"strconv"
	"time"
)

func (srv *service) HandleSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, session := FromContextLogger(ctx), FromContextSession(ctx)

	type preparedSession struct {
		Name           string
		CreatedAt      time.Time
		RecentActivity time.Time
		sessionID      int64
		Active         bool
	}
	var renderData struct {
		Sessions     []preparedSession
		ErrorMessage string
	}

	sessions, err := srv.userRepository.GetUserActiveSessions(ctx, session.UserID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get active user sessions")
		goto renderPage
	}
	for _, s := range sessions {
		prepared := preparedSession{
			Name:           generateUUIDSession(s).String(),
			CreatedAt:      s.CreatedAt.Time,
			RecentActivity: s.RecentActivity.Time,
			sessionID:      s.SessionID,
		}
		if session.SessionID == prepared.sessionID {
			prepared.Active = true
		}
		renderData.Sessions = append(renderData.Sessions, prepared)
	}
	sort.Slice(renderData.Sessions, func(i, j int) bool {
		return renderData.Sessions[i].RecentActivity.After(renderData.Sessions[j].RecentActivity)
	})

	if r.Method == http.MethodPost {
		var (
		//msgInternalServerError = "Internal Server Error."
		)

		renderData.ErrorMessage = func() string {
			switch action := r.PostFormValue("action_terminate"); action {
			case "selected":
				for _, s := range renderData.Sessions {
					if r.PostFormValue(s.Name) == "terminate" {
						if err := srv.userRepository.DeleteSession(ctx, s.sessionID); err != nil {
							log.Warn().Err(err).Int64("session_id", s.sessionID).Msg("failed to terminate session")
						} else {
							log.Debug().Int64("session_id", s.sessionID).Msg("session ended")
						}
					}
				}
			case "expect_current":
				for _, s := range renderData.Sessions {
					if s.Active {
						continue
					}
					if err := srv.userRepository.DeleteSession(ctx, s.sessionID); err != nil {
						log.Warn().Err(err).Int64("session_id", s.sessionID).Msg("failed to terminate session")
					} else {
						log.Debug().Int64("session_id", s.sessionID).Msg("session ended")
					}
				}
			default:
				log.Warn().Str("action", action).Msgf("unknown action")
			}

			http.Redirect(w, r, app.EndpointSettings, http.StatusSeeOther)
			return ""
		}()
		if renderData.ErrorMessage == "" {
			return
		}
	}

renderPage:
	srv.executeTemplate(ctx, w, templates.PageSettings, func(params *templates.Params) {
		params.Header.Title = "Settings"
		params.Data = renderData
	})
}

var uuidSessionSpace = uuid.MustParse("20937502-7289-0597-7952-629358123986")

func generateUUIDSession(session *app.Session) uuid.UUID {
	return uuid.NewSHA1(uuidSessionSpace, []byte(strconv.FormatInt(session.SessionID, 10)+session.Secret))
}
