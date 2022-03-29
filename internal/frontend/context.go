package frontend

import (
	"context"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

func NewContextReqID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, "req_id", reqID)
}

func FromContextReqID(ctx context.Context) string {
	reqID, ok := ctx.Value("req_id").(string)
	if !ok {
		return ""
	}
	return reqID
}

func NewContextLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, "logger", logger)
}

func FromContextLogger(ctx context.Context) zerolog.Logger {
	logger, ok := ctx.Value("logger").(zerolog.Logger)
	if !ok {
		return zlog.Logger
	}
	return logger
}
