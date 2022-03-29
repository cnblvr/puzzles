package frontend

import (
	"github.com/cnblvr/puzzles/app"
	zlog "github.com/rs/zerolog/log"
	"net/http"
)

func (srv *service) MiddlewareReqID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = NewContextReqID(ctx, app.GenerateReqID())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (srv *service) MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		loggerBuilder := zlog.Logger.With()
		if reqID := FromContextReqID(ctx); reqID != "" {
			loggerBuilder = loggerBuilder.Str("req_id", reqID)
		}
		ctx = NewContextLogger(ctx, loggerBuilder.Logger())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (srv *service) MiddlewareLogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := FromContextLogger(ctx)
		log.Info().
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Send()
		next.ServeHTTP(w, r)
	})
}
