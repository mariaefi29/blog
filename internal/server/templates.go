package server

import (
	"embed"
	"text/template"
)

//go:embed templates/*.gohtml
var templateFS embed.FS

func parseTemplates() (*template.Template, error) {
	return template.New("").Funcs(fm).ParseFS(templateFS, "templates/*.gohtml")
}
