package frontend

import (
	"context"
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/internal/frontend/templates"
	zlog "github.com/rs/zerolog/log"
	"html/template"
	"net/http"
	"sort"
)

type service struct {
	templates *template.Template
}

func NewService() (app.ServiceFrontend, error) {
	srv := &service{}

	var err error
	srv.templates, err = template.New("frontend").Funcs(templates.Functions()).
		ParseFS(templates.FS, append(templates.CommonTemplates(), "*.gohtml")...)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to parse FS of templates")
		return nil, err
	}

	return srv, nil
}

func (srv *service) executeTemplate(ctx context.Context, w http.ResponseWriter, name string, fn func(params *templates.Params)) {
	log := FromContextLogger(ctx)
	params := &templates.Params{
		Header: templates.Header{
			Title: "unknown page",
			Navigation: []templates.Navigation{
				{Label: "Home", Path: app.EndpointIndex, Weight: 0},
			},
		},
		Footer: templates.Footer{},
	}
	if true { // TODO session
		params.Header.Navigation = append(params.Header.Navigation,
			templates.Navigation{Label: "Log in", Path: app.EndpointLogin, Weight: 990},
			templates.Navigation{Label: "Sign up", Path: app.EndpointSignup, Weight: 991},
			templates.Navigation{Label: "Log out", Path: app.EndpointLogout, Weight: 992},
		)
	}
	fn(params)
	sort.Slice(params.Header.Navigation, func(i, j int) bool {
		return params.Header.Navigation[i].Weight < params.Header.Navigation[j].Weight
	})
	if err := srv.templates.ExecuteTemplate(w, name, params); err != nil {
		log.Error().Err(err).Msg("failed to execute template")
	}
}
