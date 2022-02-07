package res

import (
	"embed"

	// maybe use html/template

	"text/template"

	"github.com/sirupsen/logrus"
)

//go:embed tmpl/*
var TMPLFS embed.FS

// //go:embed css/*
// var CSSFS embed.FS

// //go:embed js/*
// var JSFS embed.FS

var TMPL *template.Template

func init() {
	var err error
	TMPL, err = template.ParseFS(TMPLFS, "tmpl/*")
	if err != nil {
		logrus.Errorf("err parse tmpl: %s", err)
	}
}
