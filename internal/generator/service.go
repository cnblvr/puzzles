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
}

func NewService() (app.ServiceGenerator, error) {
	srv := &service{
		config: app.NewConfig(),
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
		if err := srv.GeneratePuzzle(app.PuzzleSudokuClassic, rand.Int63()); err != nil {
			log.Error().Err(err).Msg("GeneratePuzzle failed")
		}
	}
}

func (srv *service) GeneratePuzzle(typ app.PuzzleType, seed int64) error {
	generator, err := puzzle_library.GetGenerator(typ)
	if err != nil {
		return errors.WithStack(err)
	}

	generatedSolutions := make([]app.GeneratedPuzzle, 0, 3)
	func() {
		ctx := context.Background()
		ctxGen, cancelGen := context.WithTimeout(ctx, time.Hour)
		defer cancelGen()
		generatedSolutionsChan := make(chan app.GeneratedPuzzle, 3)
		go generator.GenerateSolution(ctxGen, seed, generatedSolutionsChan)
		for solution := range generatedSolutionsChan {
			generatedSolutions = append(generatedSolutions, solution)
		}
	}()

	for _, solution := range generatedSolutions {
		err := func(solution app.GeneratedPuzzle) error {
			ctx := context.Background()

			ctxGen, cancelGen := context.WithTimeout(ctx, time.Hour)
			defer cancelGen()
			generatedCluesChan := make(chan app.GeneratedPuzzle, 3)
			go generator.GenerateClues(ctxGen, seed, solution, generatedCluesChan)
			for clues := range generatedCluesChan {
				// Save sudoku
				sudoku, err := srv.puzzleRepository.CreatePuzzle(ctx, app.CreatePuzzleParams{
					Type:            generator.Type(),
					GeneratedPuzzle: clues,
				})
				if err != nil {
					log.Error().Err(err).Msg("failed to create new puzzle in db")
					return err
				}
				log.Info().Int64("id", sudoku.ID).
					Stringer("puzzle_type", generator.Type()).
					Stringer("puzzle_level", clues.Level).
					Msg("new puzzle created and saved")
			}
			return nil
		}(solution)
		if err != nil {
			return err
		}
	}

	return nil
}
