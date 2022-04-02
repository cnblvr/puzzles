package frontend

import (
	"context"
	"encoding/base64"
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	"github.com/cnblvr/puzzles/repository"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/securecookie"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"sort"
)

type service struct {
	config               Config
	templates            *template.Template
	userRepository       app.UserRepository
	puzzleGameRepository app.PuzzleGameRepository
	secCookie            *securecookie.SecureCookie
	passwordPepper       []byte
}

func NewService() (app.ServiceFrontend, error) {
	srv := &service{
		config: NewConfig(),
	}

	var err error
	srv.templates, err = template.New("frontend").Funcs(templates.Functions()).
		ParseFS(templates.FS, append(templates.CommonTemplates(), "*.gohtml")...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse FS of templates")
	}

	srv.userRepository, err = repository.NewRedisUserRepository(func() (redis.Conn, error) {
		address, password, db, err := srv.config.RedisUserConn()
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
		return nil, errors.Wrap(err, "failed to create user repository")
	}
	srv.puzzleGameRepository, err = repository.NewRedisPuzzleGameRepository(func() (redis.Conn, error) {
		address, password, db, err := srv.config.RedisPuzzleGameConn()
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
		return nil, errors.Wrap(err, "failed to create puzzle game repository")
	}

	hashKey, blockKey, err := srv.config.SecCookieSecrets()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	srv.secCookie = securecookie.New(hashKey, blockKey)

	srv.passwordPepper, err = srv.config.PasswordPepper()
	if err != nil {
		log.Error().Err(err).Msg("failed to decode 'PASSWORD_PEPPER' env variable")
		return nil, err
	}

	return srv, nil
}

func (srv *service) executeTemplate(ctx context.Context, w http.ResponseWriter, name string, fn func(params *templates.Params)) {
	log := FromContextLogger(ctx)
	params := &templates.Params{
		Header: templates.Header{
			Title: "unknown page",
			Navigation: []templates.Navigation{
				{Label: "Home", Path: app.EndpointHome, Weight: 0},
			},
			Notification: FromContextNotificationOrNil(ctx),
		},
		Footer: templates.Footer{},
	}
	session := FromContextSession(ctx)
	if session.UserID <= 0 {
		params.Header.Navigation = append(params.Header.Navigation,
			templates.Navigation{Label: "Log in", Path: app.EndpointLogin, Weight: 991},
			templates.Navigation{Label: "Sign up", Path: app.EndpointSignup, Weight: 992},
		)
	} else {
		params.Header.Navigation = append(params.Header.Navigation,
			templates.Navigation{Label: "Settings", Path: app.EndpointSettings, Weight: 981},
			templates.Navigation{Label: "Log out", Path: app.EndpointLogout, Weight: 993},
		)
	}
	fn(params)
	sort.Slice(params.Header.Navigation, func(i, j int) bool {
		return params.Header.Navigation[i].Weight < params.Header.Navigation[j].Weight
	})
	if err := srv.templates.ExecuteTemplate(w, name, params); err != nil {
		log.Error().Err(err).Msg("failed to execute template")
	}
}

func (srv *service) hashPassword(password string, saltStr string) (string, error) {
	salt, err := base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode salt from base64")
	}
	passwordBts := append(append([]byte(password), salt...), srv.passwordPepper...)
	hash, err := bcrypt.GenerateFromPassword(passwordBts, bcrypt.DefaultCost)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return base64.StdEncoding.EncodeToString(hash), nil
}

func (srv *service) verifyPassword(password string, saltStr, hashStr string) (bool, error) {
	salt, err := base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		return false, errors.Wrap(err, "failed to decode salt from base64")
	}
	hash, err := base64.StdEncoding.DecodeString(hashStr)
	if err != nil {
		return false, errors.Wrap(err, "failed to decode hash from base64")
	}
	passwordBts := append(append([]byte(password), salt...), srv.passwordPepper...)
	return bcrypt.CompareHashAndPassword(hash, passwordBts) == nil, nil
}
