package frontend

import (
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	"github.com/pkg/errors"
	"net/http"
)

type PostHome struct {
	PuzzleType app.PuzzleType
	Level      app.PuzzleLevel
}

func (p PostHome) Parse(r *http.Request) PostHome {
	p.PuzzleType = app.PuzzleType(r.PostFormValue("puzzle_type"))
	p.Level = app.PuzzleLevel(r.PostFormValue("puzzle_level"))
	return p
}

func (p *PostHome) Validate() string {
	switch p.PuzzleType {
	case app.PuzzleSudokuClassic:
	case app.PuzzleJigsaw, app.PuzzleWindoku, app.PuzzleSudokuX, app.PuzzleKakuro:
		return fmt.Sprintf("The puzzle type '%s' is not yet supported.", p.PuzzleType)
	case "":
		return "Puzzle type is not chosen."
	default:
		return fmt.Sprintf("The puzzle type '%s' is not supported.", p.PuzzleType)
	}

	switch p.Level {
	case app.PuzzleLevelEasy, app.PuzzleLevelMedium, app.PuzzleLevelHard:
	case "":
		return "Puzzle level is not chosen."
	default:
		return fmt.Sprintf("The puzzle level '%s' is not supported.", p.Level)
	}

	return ""
}

func (srv *service) HandleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, session := FromContextLogger(ctx), FromContextSession(ctx)

	renderData := RenderDataHome{}
	if r.Method == http.MethodPost {
		var (
			msgInternalServerError = "Internal Server Error."
			msgYourPuzzlePoolEmpty = "Your puzzle pool is empty."
		)
		func() {
			post := PostHome{}.Parse(r)
			renderData.ErrorMessage = post.Validate()
			if renderData.ErrorMessage != "" {
				return
			}
			log = log.With().Stringer("puzzle_type", post.PuzzleType).
				Stringer("puzzle_level", post.Level).Logger()

			_, game, err := srv.puzzleRepository.CreateRandomPuzzleGame(ctx, app.CreateRandomPuzzleGameParams{
				Session: session,
				Type:    post.PuzzleType,
				Level:   post.Level,
			})
			switch {
			case errors.Is(err, app.ErrorPuzzlePoolEmpty):
				log.Error().Err(err).Send()
				renderData.ErrorMessage = msgYourPuzzlePoolEmpty
				return
			case err == nil:
			default:
				log.Error().Err(err).Msg("failed to create random puzzle game by params")
				renderData.ErrorMessage = msgInternalServerError
				return
			}

			http.Redirect(w, r, app.EndpointGameID(game.ID), http.StatusSeeOther)
		}()
		if renderData.ErrorMessage == "" {
			return
		}
	}

	srv.executeTemplate(r.Context(), w, templates.PageHome, func(params *templates.Params) {
		params.Header.Title = "Home"
		params.Data = renderData
	})
}

type RenderDataHome struct {
	ErrorMessage string
}
