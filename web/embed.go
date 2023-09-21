package web

import (
	"embed"
	_ "embed"
)

// content holds our static web server content
//
//go:embed *.html *.js
var Content embed.FS
