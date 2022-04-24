package generator

import (
	"context"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/puzzle_library"
	"github.com/cnblvr/puzzles/repository"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"math/rand"
	"sort"
	"time"
)

type service struct {
	config           app.Config
	puzzleRepository app.PuzzleRepository
	puzzleLibrary    app.PuzzleLibrary
	rnd              *rand.Rand
}

func NewService() (app.ServiceGenerator, error) {
	srv := &service{
		config: app.NewConfig(),
		rnd:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	app.InitHumanLogger(srv.config.Debug())

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
	}, srv.config.Debug())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create puzzle generator repository")
	}

	srv.puzzleLibrary = &puzzle_library.PuzzleLibrary{}

	return srv, nil
}

// Amount of unsolved puzzles each type and each level in the pool for all users:
//  size ( complement (
//    all puzzles X type and Y level,
//    union (
//      solvedPuzzles(for 1st user),
//      solvedPuzzles(for 2nd user),
//      ...
//    )
//  ) ) >= needPuzzlesUnsolved => it's cool
const needPuzzlesUnsolved = 5

func (srv *service) Run() error {
	for {
		type needPuzzle struct {
			typ   app.PuzzleType
			level app.PuzzleLevel
			need  int
		}
		var needPuzzles []needPuzzle
		for _, typ := range []app.PuzzleType{app.PuzzleSudokuClassic} {
			for _, level := range []app.PuzzleLevel{
				app.PuzzleLevelEasy, app.PuzzleLevelNormal, app.PuzzleLevelHard, app.PuzzleLevelHarder,
			} {
				currentNum, err := srv.puzzleRepository.GetAmountUnsolvedPuzzlesForAllUsers(context.TODO(), app.PuzzleSudokuClassic, level)
				if err != nil {
					log.Error().Err(err).Msg("PuzzleRepository.GetAmountUnsolvedPuzzlesForAllUsers() failed")
					time.Sleep(time.Second)
				}
				if currentNum < needPuzzlesUnsolved {
					needPuzzles = append(needPuzzles, needPuzzle{
						typ:   typ,
						level: level,
						need:  needPuzzlesUnsolved - currentNum,
					})
				}
			}
		}
		sort.Slice(needPuzzles, func(i, j int) bool {
			if needPuzzles[i].typ == needPuzzles[j].typ {
				return app.PuzzleLevelLess(needPuzzles[i].level, needPuzzles[j].level)
			}
			return app.PuzzleTypeLess(needPuzzles[i].typ, needPuzzles[j].typ)
		})
		log.Debug().Msgf("%+v", needPuzzles)
		for _, need := range needPuzzles {
			for idx := 1; idx <= need.need; {
				if gotLevel, err := srv.GeneratePuzzle(need.typ, srv.rnd.Int63(), need.level); err != nil {
					log.Error().Err(err).Msg("GeneratePuzzle() failed")
				} else if gotLevel != need.level {
					log.Debug().Stringer("want_level", need.level).Stringer("got_level", gotLevel).
						Str("progress", fmt.Sprintf("%d/%d", idx, need.need)).
						Msg("regenerate want level")
				} else {
					idx++
				}
			}
		}
		time.Sleep(time.Hour)
	}
}

func (srv *service) GeneratePuzzle(typ app.PuzzleType, seed int64, level app.PuzzleLevel) (app.PuzzleLevel, error) {
	creator, err := srv.puzzleLibrary.GetCreator(typ)
	if err != nil {
		return app.PuzzleLevelUnknown, errors.WithStack(err)
	}

	puzzle := creator.NewSolutionBySeed(seed)
	solution := puzzle.String()

	strategies, err := puzzle.GenerateLogic(seed, level.Strategies())
	if err != nil {
		return app.PuzzleLevelUnknown, errors.Wrap(err, "failed to generate logic")
	}

	gotLevel := strategies.Level()
	if gotLevel == level {
		sudoku, err := srv.puzzleRepository.CreatePuzzle(context.TODO(), app.CreatePuzzleParams{
			Type: creator.Type(),
			GeneratedPuzzle: app.GeneratedPuzzle{
				Seed:       seed,
				Level:      gotLevel,
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
			Stringer("puzzle_level", gotLevel).
			Msg("new puzzle created and saved")
	}

	return gotLevel, nil
}
