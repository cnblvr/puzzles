package repository

import (
	"context"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

func NewRedisPuzzleRepository(dial func() (redis.Conn, error)) (app.PuzzleRepository, error) {
	return newRedisRepository(dial)
}

func (r *redisRepository) CreateRandomPuzzleGame(ctx context.Context, params app.CreateRandomPuzzleGameParams) (*app.Puzzle, *app.PuzzleGame, error) {
	conn := r.connect()
	defer conn.Close()

	if params.Session == nil {
		return nil, nil, errors.Errorf("params.Session is nil")
	}

	keyForRandom := r.keyPuzzleByTypeAndLevel(params.Type, params.Level)

	if params.Session.UserID > 0 {
		keyTemp := r.keyTemporary()
		// TODO SDIFFSTORE is slow
		if _, err := conn.Do("SDIFFSTORE", keyTemp, keyForRandom, r.keyUserSolvedPuzzles(params.Session.UserID)); err != nil {
			return nil, nil, errors.Wrap(err, "failed to create list of unsolved puzzles for user")
		}
		if _, err := conn.Do("EXPIRE", keyTemp, time.Second*10); err != nil {
			return nil, nil, errors.Wrap(err, "failed to set expiration for list of unsolved puzzles")
		}
		keyForRandom = keyTemp
	}

	puzzleID, err := redis.Int64(conn.Do("SRANDMEMBER", keyForRandom))
	switch err {
	case redis.ErrNil:
		return nil, nil, errors.WithStack(app.ErrorPuzzlePoolEmpty)
	case nil:
	default:
		return nil, nil, errors.Wrap(err, "failed to get random puzzle id")
	}
	puzzle, err := r.getPuzzle(ctx, conn, puzzleID)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	game := &app.PuzzleGame{
		ID:        r.generatePuzzleGameID(params.Session, puzzle),
		SessionID: params.Session.SessionID,
		PuzzleID:  puzzleID,
	}
	if userID := params.Session.UserID; userID > 0 {
		game.UserID = userID
	}
	if err := r.setPuzzleGame(ctx, conn, game); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return puzzle, game, nil
}

func (r *redisRepository) GetPuzzleGame(ctx context.Context, id uuid.UUID) (*app.PuzzleGame, error) {
	conn := r.connect()
	defer conn.Close()

	puzzleGame, err := r.getPuzzleGame(ctx, conn, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return puzzleGame, nil
}

func (r *redisRepository) CreatePuzzle(ctx context.Context, params app.CreatePuzzleParams) (*app.Puzzle, error) {
	conn := r.connect()
	defer conn.Close()

	id, err := redis.Int64(conn.Do("INCR", r.keyLastPuzzleID()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to increment puzzle id")
	}

	// TODO unique app.Puzzle.Solution

	puzzle := &app.Puzzle{
		ID:       id,
		Type:     params.Type,
		Seed:     params.Seed,
		Level:    params.Level,
		Meta:     params.Meta,
		Clues:    params.Clues,
		Solution: params.Solution,
	}

	if err := r.setPuzzle(ctx, conn, puzzle); err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := conn.Do("SADD", r.keyPuzzleByTypeAndLevel(puzzle.Type, puzzle.Level), puzzle.ID); err != nil {
		return nil, errors.Wrap(err, "failed to add puzzle id in list by type and level")
	}

	return puzzle, nil
}

// Errors: app.ErrorPuzzleNotFound, unknown.
func (r *redisRepository) getPuzzle(ctx context.Context, conn redis.Conn, id int64) (*app.Puzzle, error) {
	if ok, err := redis.Bool(conn.Do("EXISTS", r.keyPuzzle(id))); err != nil {
		return nil, errors.Wrap(err, "failed to check existence puzzle")
	} else if !ok {
		return nil, errors.WithStack(app.ErrorPuzzleNotFound)
	}
	puzzleReply, err := redis.Values(conn.Do("HGETALL", r.keyPuzzle(id)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get puzzle")
	}

	puzzle := &app.Puzzle{}
	if err := redis.ScanStruct(puzzleReply, puzzle); err != nil {
		return nil, errors.Wrap(err, "failed to scan puzzle")
	}
	puzzle.ID = id

	return puzzle, nil
}

// Errors: unknown.
func (r *redisRepository) setPuzzle(ctx context.Context, conn redis.Conn, puzzle *app.Puzzle) error {
	if _, err := conn.Do("HSET", redis.Args{}.Add(r.keyPuzzle(puzzle.ID)).AddFlat(puzzle)...); err != nil {
		return errors.Wrap(err, "failed to set puzzle")
	}
	return nil
}

// Errors: app.ErrorPuzzleGameNotFound, unknown.
func (r *redisRepository) getPuzzleGame(ctx context.Context, conn redis.Conn, id uuid.UUID) (*app.PuzzleGame, error) {
	if ok, err := redis.Bool(conn.Do("EXISTS", r.keyPuzzleGame(id))); err != nil {
		return nil, errors.Wrap(err, "failed to check existence puzzle game")
	} else if !ok {
		return nil, errors.WithStack(app.ErrorPuzzleGameNotFound)
	}
	puzzleGameReply, err := redis.Values(conn.Do("HGETALL", r.keyPuzzleGame(id)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get puzzle game")
	}

	puzzleGame := &app.PuzzleGame{}
	if err := redis.ScanStruct(puzzleGameReply, puzzleGame); err != nil {
		return nil, errors.Wrap(err, "failed to scan puzzle game")
	}
	puzzleGame.ID = id

	return puzzleGame, nil
}

// Errors: unknown.
func (r *redisRepository) setPuzzleGame(ctx context.Context, conn redis.Conn, puzzleGame *app.PuzzleGame) error {
	if _, err := conn.Do("HSET", redis.Args{}.Add(r.keyPuzzleGame(puzzleGame.ID)).AddFlat(puzzleGame)...); err != nil {
		return errors.Wrap(err, "failed to set puzzle game")
	}
	return nil
}

var uuidPuzzleGameSpace = uuid.MustParse("87234032-7832-8923-8298-237589207129")

func (r *redisRepository) generatePuzzleGameID(session *app.Session, puzzle *app.Puzzle) uuid.UUID {
	return uuid.NewSHA1(uuidPuzzleGameSpace, []byte(strconv.FormatInt(session.SessionID, 10)+
		strconv.FormatInt(session.UserID, 10)+
		strconv.FormatInt(puzzle.ID, 10)))
}

func (r *redisRepository) keyLastPuzzleID() string {
	return "last_puzzle_id"
}

func (r redisRepository) keyPuzzle(id int64) string {
	return fmt.Sprintf("puzzle:%d", id)
}

func (r redisRepository) keyPuzzleByTypeAndLevel(typ app.PuzzleType, level app.PuzzleLevel) string {
	return fmt.Sprintf("puzzle_by:%s:%s", typ.String(), level.String())
}

func (r redisRepository) keyPuzzleGame(id uuid.UUID) string {
	return fmt.Sprintf("puzzle_game:%s", id.String())
}

func (r redisRepository) keyTemporary() string {
	return fmt.Sprintf("temp:%s", uuid.New().String())
}
