package templates

import (
	"embed"
	"html/template"
)

//go:embed common/*.gohtml *.gohtml
var FS embed.FS

const (
	PageHome   = "page_home"
	PageError  = "page_error"
	PageLogin  = "page_login"
	PageSignup = "page_signup"
)

func CommonTemplates() []string {
	return []string{"common/header.gohtml", "common/footer.gohtml"}
}

func Functions() template.FuncMap {
	return template.FuncMap{}
}

type Params struct {
	Header Header
	Data   interface{}
	Footer Footer
}

type Header struct {
	Title       string
	Navigation  []Navigation
	CssExternal []string
	CssInternal []template.CSS
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
