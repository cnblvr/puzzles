package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

func NewRedisUserRepository(dial func() (redis.Conn, error), debug bool) (app.UserRepository, error) {
	return newRedisRepository(dial, debug)
}

func (r *redisRepository) CreateSession(ctx context.Context, userID int64, expiration time.Duration) (*app.Session, error) {
	conn := r.connect()
	defer conn.Close()

	id, err := redis.Int64(conn.Do("INCR", r.keyLastSessionID()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to increment session id")
	}
	if expiration == 0 {
		expiration = app.DefaultCookieSessionExpiration
	}
	secret := make([]byte, 16)
	rand.Read(secret)

	session := &app.Session{
		CookieSession: app.CookieSession{
			SessionID: id,
			UserID:    userID,
			Secret:    hex.EncodeToString(secret),
		},
		CreatedAt:         app.DateTime{Time: time.Now().UTC()},
		RecentActivity:    app.DateTime{Time: time.Now().UTC()},
		ExpirationSeconds: int64(expiration / time.Second),
	}

	if _, err := conn.Do("HSET", redis.Args{}.Add(r.keySession(id)).AddFlat(session)...); err != nil {
		return nil, errors.Wrap(err, "failed to set session")
	}
	if _, err := conn.Do("EXPIRE", r.keySession(id), session.ExpirationSeconds); err != nil {
		return nil, errors.Wrap(err, "failed to set expiration for session")
	}

	if userID > 0 {
		if _, err := conn.Do("SADD", r.keyUserSessions(userID), id); err != nil {
			return nil, errors.Wrap(err, "failed to add session id in user session list")
		}
	}

	return session, nil
}

func (r *redisRepository) GetSession(ctx context.Context, sessionID int64) (*app.Session, error) {
	conn := r.connect()
	defer conn.Close()

	session, err := r.getSession(ctx, conn, sessionID)
	if err != nil {
		return nil, err
	}

	session.RecentActivity = app.DateTime{Time: time.Now().UTC()}
	if _, err := conn.Do("HSET", redis.Args{}.Add(r.keySession(session.SessionID)).AddFlat(session)...); err != nil {
		return nil, errors.Wrap(err, "failed to update session")
	}

	if _, err := conn.Do("EXPIRE", r.keySession(sessionID), session.ExpirationSeconds); err != nil {
		return nil, errors.Wrap(err, "failed to set expiration for session")
	}

	return session, nil
}

// Errors: app.ErrorSessionNotFound, unknown.
func (r *redisRepository) getSession(ctx context.Context, conn redis.Conn, sessionID int64) (*app.Session, error) {
	if ok, err := redis.Bool(conn.Do("EXISTS", r.keySession(sessionID))); err != nil {
		return nil, errors.Wrap(err, "failed to check existence session")
	} else if !ok {
		return nil, errors.WithStack(app.ErrorSessionNotFound)
	}
	sessionReply, err := redis.Values(conn.Do("HGETALL", r.keySession(sessionID)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}

	session := &app.Session{}
	if err := redis.ScanStruct(sessionReply, session); err != nil {
		return nil, errors.Wrap(err, "failed to scan session")
	}
	session.SessionID = sessionID

	return session, nil
}

func (r *redisRepository) UpdateSession(ctx context.Context, session *app.Session) error {
	conn := r.connect()
	defer conn.Close()

	if _, err := conn.Do("HSET", redis.Args{}.Add(r.keySession(session.SessionID)).AddFlat(session)...); err != nil {
		return errors.Wrap(err, "failed to update session")
	}
	if _, err := conn.Do("EXPIRE", r.keySession(session.SessionID), session.ExpirationSeconds); err != nil {
		return errors.Wrap(err, "failed to set expiration for session")
	}

	if session.UserID > 0 {
		if _, err := conn.Do("SADD", r.keyUserSessions(session.UserID), session.SessionID); err != nil {
			return errors.Wrap(err, "failed to add session id in user session list")
		}
	}

	return nil
}

func (r *redisRepository) DeleteSession(ctx context.Context, sessionID int64) error {
	conn := r.connect()
	defer conn.Close()

	if deleted, _ := redis.Int(conn.Do("DEL", r.keySession(sessionID))); deleted != 1 {
		return errors.Errorf("failed to delete session")
	}

	return nil
}

func (r *redisRepository) CreateUser(ctx context.Context, username string, salt, hash string) (*app.User, error) {
	conn := r.pool.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("HEXISTS", r.keyUsernames(), username))
	if err != nil {
		return nil, errors.Wrap(err, "failed to check existence username")
	}
	if ok {
		return nil, errors.WithStack(app.ErrorUsernameIsNotVacant)
	}

	id, err := redis.Int64(conn.Do("INCR", r.keyLastUserID()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to increment last user id")
	}

	user := &app.User{
		ID:            id,
		Username:      username,
		Salt:          salt,
		Hash:          hash,
		HashTimestamp: app.DateTime{Time: time.Now().UTC()},
	}

	if _, err := conn.Do("HSET", redis.Args{}.Add(r.keyUser(id)).AddFlat(user)...); err != nil {
		return nil, errors.Wrap(err, "failed to set user")
	}
	if _, err := conn.Do("HSET", r.keyUsernames(), username, id); err != nil {
		return nil, errors.Wrap(err, "failed to occupy username")
	}

	return user, nil
}

func (r *redisRepository) GetUser(ctx context.Context, id int64) (*app.User, error) {
	conn := r.connect()
	defer conn.Close()

	userReply, err := redis.Values(conn.Do("HGETALL", r.keyUser(id)))
	switch err {
	case redis.ErrNil:
		return nil, errors.WithStack(app.ErrorUserNotFound)
	case nil:
	default:
		return nil, errors.Wrap(err, "failed to get user")
	}

	user := &app.User{}
	if err := redis.ScanStruct(userReply, user); err != nil {
		return nil, errors.Wrap(err, "failed to scan user")
	}
	user.ID = id

	return user, nil
}

func (r *redisRepository) GetUserByUsername(ctx context.Context, username string) (*app.User, error) {
	conn := r.connect()
	defer conn.Close()

	id, err := redis.Int64(conn.Do("HGET", r.keyUsernames(), username))
	switch err {
	case redis.ErrNil:
		return nil, errors.WithStack(app.ErrorUserNotFound)
	case nil:
	default:
		return nil, errors.Wrap(err, "failed to get user id")
	}

	userReply, err := redis.Values(conn.Do("HGETALL", r.keyUser(id)))
	switch err {
	case redis.ErrNil:
		return nil, errors.WithStack(app.ErrorUserNotFound)
	case nil:
	default:
		return nil, errors.Wrap(err, "failed to get user")
	}

	user := &app.User{}
	if err := redis.ScanStruct(userReply, user); err != nil {
		return nil, errors.Wrap(err, "failed to scan user")
	}
	user.ID = id

	return user, nil
}

func (r *redisRepository) GetUserActiveSessions(ctx context.Context, userID int64) ([]*app.Session, error) {
	conn := r.connect()
	defer conn.Close()

	listSessionID, err := redis.Int64s(conn.Do("SMEMBERS", r.keyUserSessions(userID)))
	switch err {
	case redis.ErrNil:
		return nil, errors.WithStack(app.ErrorUserNotFound)
	case nil:
	default:
		return nil, errors.Wrap(err, "failed to get user session list")
	}

	sessions := make([]*app.Session, 0, len(listSessionID))
	for _, sessionID := range listSessionID {
		session, err := r.getSession(ctx, conn, sessionID)
		switch {
		case errors.Is(err, app.ErrorSessionNotFound):
			// Delete sessions that have already expired
			if _, err := conn.Do("SREM", r.keyUserSessions(userID), sessionID); err != nil {
				log.Warn().Err(err).Msg("failed to remove old session id from user session list")
			}
			continue
		case err == nil:
		default:
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *redisRepository) GetUserPreferences(ctx context.Context, userID int64) (*app.UserPreferences, error) {
	conn := r.connect()
	defer conn.Close()

	preferencesReply, err := redis.Values(conn.Do("HGETALL", r.keyUserPreferences(userID)))
	switch err {
	case redis.ErrNil, nil:
	default:
		return nil, errors.Wrap(err, "failed to get user preferences")
	}

	preferences := &app.UserPreferences{}
	preferences.Defaults()
	if err := redis.ScanStruct(preferencesReply, preferences); err != nil {
		return nil, errors.Wrap(err, "failed to scan user preferences")
	}
	preferences.UserID = userID

	return preferences, nil
}

func (r *redisRepository) SetUserPreferences(ctx context.Context, preferences *app.UserPreferences) error {
	conn := r.connect()
	defer conn.Close()

	if _, err := conn.Do("HSET", redis.Args{}.Add(r.keyUserPreferences(preferences.UserID)).AddFlat(preferences)...); err != nil {
		return errors.Wrap(err, "failed to set user preferences")
	}

	return nil
}

// keyLastSessionID returns a key of the last used ID for store app.Session.
// Value type is int64.
func (r *redisRepository) keyLastSessionID() string {
	return "last_session_id"
}

// keySession returns a key to store app.Session by id.
// The value type is a hash for app.Session structure.
func (r *redisRepository) keySession(id int64) string {
	return fmt.Sprintf("session:%d", id)
}

// keyLastUserID returns a key of the last used ID for store app.User.
// Value type is int64.
func (r *redisRepository) keyLastUserID() string {
	return "last_user_id"
}

// keyUser returns a key to store app.User by id.
// The value type is a hash for app.User structure.
func (r *redisRepository) keyUser(id int64) string {
	return fmt.Sprintf("user:%d", id)
}

// keyUserSessions returns a key to a list of user sessions.
// The value type is a set of session identifiers.
func (r *redisRepository) keyUserSessions(id int64) string {
	return fmt.Sprintf("%s:sessions", r.keyUser(id))
}

func (r *redisRepository) keyUserSolvedPuzzles(id int64) string {
	return fmt.Sprintf("%s:solved_puzzles", r.keyUser(id))
}

func (r *redisRepository) keyUserPreferences(id int64) string {
	return fmt.Sprintf("%s:preferences", r.keyUser(id))
}

// keyUsernames returns a key to check for existence of a username.
// The value type is a hash:
//  the key as the username (string)
//  the value as the user id (int64)
func (r *redisRepository) keyUsernames() string {
	return "usernames"
}
