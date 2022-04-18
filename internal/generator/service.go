package generator

import (
	"context"
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/puzzle_library"
	"github.com/cnblvr/puzzles/repository"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"math/rand"
	"time"
)

type service struct {
	config           app.Config
	puzzleRepository app.PuzzleRepository
	rnd              *rand.Rand
}

func NewService() (app.ServiceGenerator, error) {
	srv := &service{
		config: app.NewConfig(),
		rnd:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	var err error
	srv.puzzleRepository, err = repository.NewRedisPuzzleRepository(func() (redis.Conn, error) {
		address, password, db, err := srv.config.RedisPuzzleConn()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return redis.Dial(
			"tcp", address,
			redis.DialPassword(password),
			redis.DialDatabase(db),
		)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create puzzle generator repository")
	}

	return srv, nil
}

func (srv *service) Run() error {
	for {
		for _, level := range []app.PuzzleLevel{
			app.PuzzleLevelEasy, app.PuzzleLevelNormal, app.PuzzleLevelHard, app.PuzzleLevelHarder,
		} {
			for {
				if gotLevel, err := srv.GeneratePuzzle(app.PuzzleSudokuClassic, srv.rnd.Int63(), level); err != nil {
					log.Error().Err(err).Msg("GeneratePuzzle failed")
					time.Sleep(time.Second)
				} else if gotLevel != level {
					log.Info().Stringer("want_level", level).Stringer("got_level", gotLevel).Msg("regenerate want level")
				} else {
					break
				}
			}
			//if level == app.PuzzleLevelHarder {
			//	log.Fatal().Send()
			//}
		}
	}
}

func (srv *service) GeneratePuzzle(typ app.PuzzleType, seed int64, level app.PuzzleLevel) (app.PuzzleLevel, error) {
	creator, err := puzzle_library.GetCreator(typ)
	if err != nil {
		return app.PuzzleLevelUnknown, errors.WithStack(err)
	}

	puzzle := creator.NewSolutionBySeed(seed)

	solution := puzzle.String()

	strategies, err := puzzle.GenerateLogic(seed, level.Strategies())
	if err != nil {
		return app.PuzzleLevelUnknown, errors.Wrap(err, "failed to generate logic")
	}
	level = strategies.Level()

	sudoku, err := srv.puzzleRepository.CreatePuzzle(context.TODO(), app.CreatePuzzleParams{
		Type: creator.Type(),
		GeneratedPuzzle: app.GeneratedPuzzle{
			Seed:       seed,
			Level:      level,
			Meta:       "{}",
			Clues:      puzzle.String(),
			Candidates: puzzle.GetCandidates(),
			Solution:   solution,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create new puzzle in db")
		return app.PuzzleLevelUnknown, err
	}
	log.Info().Int64("id", sudoku.ID).
		Stringer("puzzle_type", creator.Type()).
		Stringer("puzzle_level", level).
		Msg("new puzzle created and saved")

	return level, nil
}
