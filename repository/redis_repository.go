package repository

import (
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"log"
	"os"
	"time"
)

type redisRepository struct {
	pool  *redis.Pool
	debug bool
}

func (r *redisRepository) connect() redis.Conn {
	conn := r.pool.Get()
	if r.debug {
		conn = redis.NewLoggingConn(conn, log.New(os.Stdout, "", 0), "REDIS: ")
	}
	return conn
}

func newRedisRepository(dial func() (redis.Conn, error), debug bool) (*redisRepository, error) {
	pool := &redis.Pool{
		Dial: dial,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Second*10 {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
		MaxIdle:     3,
		MaxActive:   0,
		IdleTimeout: time.Minute,
	}
	conn := pool.Get()
	defer conn.Close()
	if _, err := conn.Do("PING"); err != nil {
		return nil, errors.Wrap(err, "failed to ping redis connection")
	}
	return &redisRepository{
		pool:  pool,
		debug: debug,
	}, nil
}
