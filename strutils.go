package main

import (
	"strings"
	"text/template"
)

func format(fmt string, v interface{}) string {
	t := new(template.Template)
	b := new(strings.Builder)
	_ = template.Must(t.Parse(fmt)).Execute(b, v) // TODO: обработать возможные ошибки
	return b.String()
}
