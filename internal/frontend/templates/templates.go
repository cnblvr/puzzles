package templates

import (
	"embed"
	"github.com/cnblvr/puzzles/app"
	"html/template"
)

//go:embed common/*.gohtml *.gohtml
var FS embed.FS

const (
	PageHome     = "page_home"
	PageError    = "page_error"
	PageLogin    = "page_login"
	PageSignup   = "page_signup"
	PageSettings = "page_settings"
)

func CommonTemplates() []string {
	return []string{"common/header.gohtml", "common/footer.gohtml"}
}

func Functions() template.FuncMap {
	return template.FuncMap{
		"add_internal_css": func(h Header, css ...template.CSS) Header {
			h.CssInternal = append(h.CssInternal, css...)
			return h
		},
	}
}

type Params struct {
	Header Header
	Data   interface{}
	Footer Footer
}

type Header struct {
	Title        string
	Navigation   []Navigation
	Notification *app.CookieNotification
	CssExternal  []string
	CssInternal  []template.CSS
}

type Footer struct {
	JsExternal []string
	JsInternal []template.JS
}

type Navigation struct {
	Label  string
	Path   string
	Weight int
}
