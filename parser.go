package main

import (
	"strings"
)

func parseCommand(s string) (command string, params []string) {
	quoted := false
	items := strings.FieldsFunc(s, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})
	for i := range items {
		items[i] = strings.Trim(items[i], `"`)
	}

	return items[0], items[1:]
}
