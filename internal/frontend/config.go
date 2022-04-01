package frontend

import (
	"encoding/base64"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

type Config interface {
	SecCookieSecrets() ([]byte, []byte, error)
	RedisUserConn() (string, string, int, error)
	PasswordPepper() ([]byte, error)
}

type config struct{}

func NewConfig() Config {
	return &config{}
}

const (
	envvarSecCookieHashKey  = "SEC_COOKIE_HASH_KEY"
	envvarSecCookieBlockKey = "SEC_COOKIE_BLOCK_KEY"
	envvarRedisAddress      = "REDIS_ADDRESS"
	envvarRedisPassword     = "REDIS_PASSWORD"
	envvarRedisUserDB       = "REDIS_USER_DB"
	envvarPasswordPepper    = "PASSWORD_PEPPER"
)

func (c config) SecCookieSecrets() ([]byte, []byte, error) {
	hashKey, err := c.getBytes(envvarSecCookieHashKey)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	blockKey, err := c.getBytes(envvarSecCookieBlockKey)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return hashKey, blockKey, nil
}

func (c config) RedisUserConn() (string, string, int, error) {
	return c.redisConn(envvarRedisUserDB)
}

func (c config) PasswordPepper() ([]byte, error) {
	return c.getBytes(envvarPasswordPepper)
}

func (config) redisConn(envvarRedisDB string) (string, string, int, error) {
	address, ok := os.LookupEnv(envvarRedisAddress)
	if !ok {
		return "", "", 0, errors.Errorf("envvar '%s' not set", envvarRedisAddress)
	}

	password := os.Getenv(envvarRedisPassword)

	if s, ok := os.LookupEnv(envvarRedisDB); !ok {
		return "", "", 0, errors.Errorf("envvar '%s' not set", envvarRedisDB)
	} else if db, err := strconv.Atoi(s); err != nil {
		return "", "", 0, errors.Wrapf(err, "failed to parse '%s' as int", envvarRedisDB)
	} else {
		return address, password, db, nil
	}
}

func (config) getBytes(envName string) ([]byte, error) {
	str, ok := os.LookupEnv(envName)
	if !ok {
		return nil, errors.Errorf("envvar '%s' not set", envName)
	}
	val, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode envvar '%s'", envName)
	}
	return val, nil
}
