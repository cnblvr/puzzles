package frontend

import (
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	"net/http"
)

func (srv *service) HandleSignup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	srv.executeTemplate(ctx, w, templates.PageSignup, func(params *templates.Params) {
		params.Header.Title = "Signup"
	})
}
