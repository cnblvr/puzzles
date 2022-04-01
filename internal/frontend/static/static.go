package static

import "embed"

// CSS is a repository of CSS files.
//go:embed css/*.css
var CSS embed.FS

const (
	CssSudoku = "sudoku"
)

// JS is a repository of JS files.
//go:embed js/*.js
var JS embed.FS

const (
	JsSudoku = "sudoku"
	JsWs     = "ws"
)

// Favicon is a website icon file.
//go:embed favicon.ico
var Favicon embed.FS
