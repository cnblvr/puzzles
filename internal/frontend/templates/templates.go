package templates

import (
	"embed"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
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
	PageGameID   = "page_game_id"
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
		"add_internal_js": func(f Footer, js ...template.JS) Footer {
			f.JsInternal = append(f.JsInternal, js...)
			return f
		},
		"dict": func(in ...interface{}) ([]keyValue, error) {
			if len(in)%2 == 1 {
				return nil, errors.Errorf("length list of keyvalue invalid")
			}
			kv := make([]keyValue, 0, len(in)/2)
			for i := 0; i < len(in); i += 2 {
				kv = append(kv, keyValue{Key: in[i], Val: in[i+1]})
			}
			return kv, nil
		},
		"kv": func(key interface{}, val interface{}) keyValue {
			return keyValue{Key: key, Val: val}
		},
	}
}

type keyValue struct {
	Key interface{}
	Val interface{}
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
