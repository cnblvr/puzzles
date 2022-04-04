package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"time"
)

// UserRepository represents the create, read, update, and delete functions for the Session and User structures in the
// data store.
// TODO create method UpdateUser(ctx, *User)
type UserRepository interface {
	// CreateSession creates and returns a session with a limited lifetime.
	// userID can be empty if the session is created for an anonymous user.
	//
	// Errors: unknown.
	CreateSession(ctx context.Context, userID int64, expiration time.Duration) (*Session, error)

	// GetSession searches up and returns the session by sessionID.
	// The lifetime for a record is redefined in the database from the Session.ExpirationSeconds value.
	//
	// Errors: ErrorSessionNotFound, unknown.
	GetSession(ctx context.Context, sessionID int64) (*Session, error)

	// UpdateSession updates session.
	// The lifetime of the record is also updated.
	//
	// Errors: unknown.
	UpdateSession(ctx context.Context, session *Session) error

	// DeleteSession deletes the session by sessionID.
	//
	// Errors: unknown.
	DeleteSession(ctx context.Context, sessionID int64) error

	// CreateUser creates a user.
	// It is forbidden to create a user with the same username in the future.
	//
	// Errors: ErrorUsernameIsNotVacant, unknown.
	//
	// TODO: create struct ParamsCreateUser
	CreateUser(ctx context.Context, username string, salt, hash string) (*User, error)

	// GetUser returns a user by id.
	//
	// Errors: ErrorUserNotFound, unknown.
	GetUser(ctx context.Context, id int64) (*User, error)

	// GetUserByUsername returns a user by username.
	//
	// Errors: ErrorUserNotFound, unknown.
	GetUserByUsername(ctx context.Context, username string) (*User, error)

	// GetUserActiveSessions returns all active sessions for userID.
	//
	// Errors: ErrorUserNotFound, unknown.
	GetUserActiveSessions(ctx context.Context, userID int64) ([]*Session, error)
}

// User presents a user in this system.
type User struct {
	// ID is the user's system identifier in the database.
	// Auto-increments in UserRepository.CreateUser.
	ID int64 `json:"id" redis:"-"`

	// Username is the user identifier for identification.
	Username string `json:"username" redis:"username"`

	// Salt is random bytes to make it harder to guess the user's original password in case of an unauthorized database
	// dump.
	// Base64 encoded.
	Salt string `json:"salt" redis:"salt"`

	// Hash is the hash of the password, salt and pepper.
	// Base64 encoded.
	Hash string `json:"hash" redis:"hash"`

	// HashTimestamp is the date and time the password was modified.
	HashTimestamp DateTime `json:"hash_ts" redis:"hash_ts"`
}

type CookieSession struct {
	SessionID int64  `json:"session_id" redis:"-"`
	UserID    int64  `json:"user_id,omitempty" redis:"user_id"`
	Secret    string `json:"secret" redis:"secret"`
}

type Session struct {
	CookieSession
	CreatedAt         DateTime `json:"created_at" redis:"created_at"`
	RecentActivity    DateTime `json:"recent_activity" redis:"recent_activity"`
	ExpirationSeconds int64    `json:"expiration" redis:"expiration"`
}

func (s *Session) ValidateWith(cookieSession *CookieSession) error {
	switch {
	case s.SessionID != cookieSession.SessionID:
	case s.UserID != cookieSession.UserID:
	case s.Secret != cookieSession.Secret:
	default:
		return nil
	}
	return errors.WithStack(ErrorSessionInvalid)
}

const DefaultCookieSessionExpiration = time.Hour * 24

var (
	ErrorSessionNotFound     = fmt.Errorf("session not found")
	ErrorSessionInvalid      = fmt.Errorf("session invalid")
	ErrorUsernameIsNotVacant = fmt.Errorf("username is not vacant")
	ErrorUserNotFound        = fmt.Errorf("user not found")
)

var regexpUsername = regexp.MustCompile(`^[A-Za-z0-9-._]+$`)

const (
	MinLengthUsername = 3
	MaxLengthUsername = 32
)

func ValidateUsername(username string) error {
	switch {
	case !regexpUsername.MatchString(username):
		return errors.Errorf("username contains unsupported characters")
	case MinLengthUsername > len(username) || len(username) > MaxLengthUsername:
		return errors.Errorf("username wrong length")
	default:
		return nil
	}
}

const (
	MinLengthPassword = 3
	MaxLengthPassword = 64
)

func ValidatePassword(password string) error {
	switch {
	case MinLengthPassword > len(password) || len(password) > MaxLengthUsername:
		return errors.Errorf("password wrong length")
	default:
		return nil
	}
}

func GeneratePasswordSalt() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	return base64.StdEncoding.EncodeToString(buf)
}
