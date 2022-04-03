package frontend

import (
	"context"
	"github.com/cnblvr/puzzles/app"
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

func NewContextSession(ctx context.Context, session *app.Session) context.Context {
	return context.WithValue(ctx, "session", session)
}

func FromContextSession(ctx context.Context) *app.Session {
	session, ok := ctx.Value("session").(*app.Session)
	if !ok {
		return &app.Session{}
	}
	return session
}

func NewContextServiceFrontend(ctx context.Context, srv *service) context.Context {
	return context.WithValue(ctx, "service_frontend", srv)
}

func FromContextServiceFrontendOrNil(ctx context.Context) *service {
	srv, ok := ctx.Value("service_frontend").(*service)
	if !ok {
		return nil
	}
	return srv
}

func NewContextNotification(ctx context.Context, notification *app.CookieNotification) context.Context {
	return context.WithValue(ctx, "notification", notification)
}

func FromContextNotification(ctx context.Context) *app.CookieNotification {
	n := FromContextNotificationOrNil(ctx)
	if n == nil {
		return &app.CookieNotification{}
	}
	return n
}

func FromContextNotificationOrNil(ctx context.Context) *app.CookieNotification {
	notification, ok := ctx.Value("notification").(*app.CookieNotification)
	if !ok {
		return nil
	}
	return notification
}
