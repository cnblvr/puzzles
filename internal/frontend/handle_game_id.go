package frontend

import (
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/internal/frontend/static"
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	"github.com/pkg/errors"
	"net/http"
)

type RenderDataGameID struct {
	GameID         string
	UseHighlights  bool
	ShowCandidates bool
	ShowWrongs     bool
}

func (srv *service) HandleGameID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, session := FromContextLogger(ctx), FromContextSession(ctx)
	renderData := RenderDataGameID{
		UseHighlights:  app.DefaultUseHighlights,
		ShowCandidates: app.DefaultShowCandidates,
		ShowWrongs:     app.DefaultShowWrongs,
	}

	gameID, err := app.EndpointGameID{}.MuxParse(r)
	if err != nil {
		log.Warn().Err(err).Msg("incorrect game_id")
		srv.setCookieNotificationToResponse(w, app.NotificationWarning, "Incorrect game id.")
		http.Redirect(w, r, app.EndpointHome, http.StatusSeeOther)
		return
	}
	log = log.With().Stringer("game_id", gameID).Logger()
	renderData.GameID = gameID.String()

	game, err := srv.puzzleRepository.GetPuzzleGame(ctx, gameID)
	if err != nil {
		msg := "Internal server error."
		if errors.Is(err, app.ErrorPuzzleGameNotFound) {
			log.Error().Msg("puzzle game not found")
			msg = "Game not found."
		} else {
			log.Error().Err(err).Msg("failed to ge puzzle game")
		}
		srv.setCookieNotificationToResponse(w, app.NotificationError, msg)
		http.Redirect(w, r, app.EndpointHome, http.StatusSeeOther)
		return
	}

	if err := game.ValidateSession(session); err != nil {
		log.Info().Err(err).Send()
		srv.setCookieNotificationToResponse(w, app.NotificationError, "This game is not available to you.")
		http.Redirect(w, r, app.EndpointHome, http.StatusSeeOther)
		return
	}

	if session.UserID > 0 {
		up, err := srv.userRepository.GetUserPreferences(ctx, session.UserID)
		if err == nil {
			renderData.UseHighlights = up.UseHighlights
			renderData.ShowCandidates = up.ShowCandidates
			renderData.ShowWrongs = up.ShowWrongs
		}
	}

	srv.executeTemplate(ctx, w, templates.PageGameID, func(params *templates.Params) {
		params.Header.Title = "Puzzle game"
		params.Header.CssExternal = append(params.Header.CssExternal, static.CssSudoku)
		params.Data = renderData
		params.Footer.JsExternal = append(params.Footer.JsExternal, static.JsWs, static.JsSudoku)
	})
}
