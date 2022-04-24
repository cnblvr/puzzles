package app

import (
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

func InitHumanLogger(debug bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	zlog.Logger = zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = "2006-01-02 15:04:05.000000Z"
	})).With().Timestamp().Caller().Logger()
	if !debug {
		zlog.Logger = zlog.Logger.Level(zerolog.InfoLevel)
	}
}
