package frontend

import (
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	"net/http"
)

func (srv *service) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	srv.executeTemplate(ctx, w, templates.PageLogin, func(params *templates.Params) {
		params.Header.Title = "Login"
	})
}
