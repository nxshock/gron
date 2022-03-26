package main

import (
	_ "embed"
	"html/template"

	log "github.com/sirupsen/logrus"
)

//go:embed index.htm
var indexTemplateStr string

var indexTemplate *template.Template

func initTemplate() {
	var err error
	indexTemplate, err = template.New("index").Parse(indexTemplateStr) // TODO: optimize
	if err != nil {
		log.Fatalln("init template error:", err)
	}
}
