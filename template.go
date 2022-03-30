package main

import (
	"embed"
	"html/template"

	log "github.com/sirupsen/logrus"
)

//go:embed webui/*
var siteFS embed.FS

var templates *template.Template

func initTemplate() {
	var err error
	templates, err = template.ParseFS(siteFS, "webui/*.htm")
	if err != nil {
		log.Fatalln("init template error:", err)
	}
}
