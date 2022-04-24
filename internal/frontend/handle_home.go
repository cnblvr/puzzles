package frontend

import (
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

type PostHome struct {
	PuzzleType        app.PuzzleType
	Level             app.PuzzleLevel
	CandidatesAtStart bool
}

func (p PostHome) Parse(r *http.Request) PostHome {
	p.PuzzleType = app.PuzzleType(r.PostFormValue("puzzle_type"))
	p.Level = app.PuzzleLevel(r.PostFormValue("puzzle_level"))
	p.CandidatesAtStart, _ = strconv.ParseBool(r.PostFormValue("candidates_at_start"))
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
	case app.PuzzleLevelEasy, app.PuzzleLevelNormal, app.PuzzleLevelHard, app.PuzzleLevelHarder:
	case app.PuzzleLevelUnknown:
		return "Puzzle level is not chosen."
	default:
		return fmt.Sprintf("The puzzle level '%s' is not supported.", p.Level)
	}

	return ""
}

func (srv *service) HandleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, session := FromContextLogger(ctx), FromContextSession(ctx)

	renderData := RenderDataHome{
		PuzzleTypes: []listItem{
			{ID: string(app.PuzzleSudokuClassic), Name: "Sudoku Classic", Default: true},
			{ID: string(app.PuzzleJigsaw), Name: "Jigsaw", Disabled: true},
			{ID: string(app.PuzzleWindoku), Name: "Windoku", Disabled: true},
			{ID: string(app.PuzzleSudokuX), Name: "Sudoku X", Disabled: true},
			{ID: string(app.PuzzleKakuro), Name: "Kakuro", Disabled: true},
		},
		PuzzleLevels: []listItem{
			{ID: string(app.PuzzleLevelEasy), Name: "Easy"},
			{ID: string(app.PuzzleLevelNormal), Name: "Normal", Default: true},
			{ID: string(app.PuzzleLevelHard), Name: "Hard"},
			{ID: string(app.PuzzleLevelHarder), Name: "Harder"},
			{ID: string(app.PuzzleLevelInsane), Name: "Insane", Disabled: true},
			{ID: string(app.PuzzleLevelDemon), Name: "Demon", Disabled: true},
			{ID: string(app.PuzzleLevelCustom), Name: "Custom", Disabled: true},
		},
		CandidatesAtStart: app.DefaultCandidatesAtStart,
	}

	var up *app.UserPreferences
	if session.UserID > 0 {
		var err error
		up, err = srv.userRepository.GetUserPreferences(ctx, session.UserID)
		if err == nil {
			for i := 0; i < len(renderData.PuzzleTypes); i++ {
				renderData.PuzzleTypes[i].Default = false
				if renderData.PuzzleTypes[i].ID == up.PuzzleType.String() {
					renderData.PuzzleTypes[i].Default = true
				}
			}
			for i := 0; i < len(renderData.PuzzleLevels); i++ {
				renderData.PuzzleLevels[i].Default = false
				if renderData.PuzzleLevels[i].ID == up.PuzzleLevel.String() {
					renderData.PuzzleLevels[i].Default = true
				}
			}
			renderData.CandidatesAtStart = up.CandidatesAtStart
		}
	}

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
			log = log.With().Stringer("puzzle_type", post.PuzzleType).Stringer("puzzle_level", post.Level).Logger()

			puzzle, game, err := srv.puzzleRepository.CreateRandomPuzzleGame(ctx, app.CreateRandomPuzzleGameParams{
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
			game.State = puzzle.Clues
			if post.CandidatesAtStart {
				game.StateCandidates = puzzle.Candidates
			} else {
				game.StateCandidates = "{}"
			}
			if err := srv.puzzleRepository.UpdatePuzzleGame(ctx, game); err != nil {
				log.Error().Err(err).Msg("failed to update puzzle game")
				renderData.ErrorMessage = msgInternalServerError
				return
			}

			if up != nil {
				up.PuzzleType = post.PuzzleType
				up.PuzzleLevel = post.Level
				up.CandidatesAtStart = post.CandidatesAtStart
				if err := srv.userRepository.SetUserPreferences(ctx, up); err != nil {
					log.Warn().Err(err).Msg("failed to set user preferences")
				}
			}

			http.Redirect(w, r, app.EndpointGameID{}.Path(game.ID), http.StatusSeeOther)
			return
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

type listItem struct {
	ID       string
	Name     string
	Default  bool
	Disabled bool
}

type RenderDataHome struct {
	PuzzleTypes       []listItem
	PuzzleLevels      []listItem
	CandidatesAtStart bool
	ErrorMessage      string
}
