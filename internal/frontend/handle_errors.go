package frontend

import (
	"fmt"
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	"net/http"
)

func (srv *service) HandleInternalServerError(w http.ResponseWriter, r *http.Request) {
	srv.executeTemplate(r.Context(), w, templates.PageError, func(params *templates.Params) {
		params.Header.Title = "Error"
		params.Data = struct {
			ErrorMessage string
		}{
			ErrorMessage: "Internal server error.",
		}
	})
}

func (srv *service) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	srv.executeTemplate(r.Context(), w, templates.PageError, func(params *templates.Params) {
		params.Header.Title = "Error"
		params.Data = struct {
			ErrorMessage string
		}{
			ErrorMessage: fmt.Sprintf("Page %s not found.", r.URL.Path),
		}
	})
}

func (srv *service) HandleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	srv.executeTemplate(r.Context(), w, templates.PageError, func(params *templates.Params) {
		params.Header.Title = "Error"
		params.Data = struct {
			ErrorMessage string
		}{
			ErrorMessage: fmt.Sprintf("Method %s not allowed for page %s.", r.Method, r.URL.Path),
		}
	})
}
