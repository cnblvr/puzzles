package frontend

import (
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	"net/http"
)

func (srv *service) HandleHome(w http.ResponseWriter, r *http.Request) {
	srv.executeTemplate(r.Context(), w, templates.PageHome, func(params *templates.Params) {
		params.Header.Title = "Home"
	})
}
