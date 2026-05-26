package server

import (
	"embed"
	"html/template"
	"strconv"
	"time"
)

//go:embed templates/*.gohtml
var templateFS embed.FS

func parseTemplates(googleAnalyticsMeasurementID string) (*template.Template, error) {
	fm := template.FuncMap{
		"truncate":      truncate,
		"incline":       commentIncline,
		"categoryTitle": categoryTitle,
		"postDate":      postDate,
		"safeHTML":      safeHTML,
		"googleAnalyticsMeasurementID": func() string {
			return googleAnalyticsMeasurementID
		},
	}

	return template.New("").Funcs(fm).ParseFS(templateFS, "templates/*.gohtml")
}

func truncate(s string) string {
	var numRunes = 0
	for index := range s {
		numRunes++
		k := rune(s[index])
		if (numRunes > 150) && (k == 32) {
			return s[:index]
		}
	}
	return s
}

func commentIncline(cnt int) string {
	if cnt == 1 {
		return "1 Comment"
	}

	return strconv.Itoa(cnt) + " Comments"
}

func categoryTitle(categoryEng, fallback string) string {
	switch categoryEng {
	case "family":
		return "Family"
	case "travel":
		return "Adventures"
	case "english":
		return "Languages"
	case "webdev":
		return "Tech"
	default:
		return fallback
	}
}

func postDate(createdAt time.Time, fallback string) string {
	if createdAt.IsZero() {
		return fallback
	}

	return createdAt.Format("January 2, 2006")
}

func safeHTML(s string) template.HTML {
	return template.HTML(s)
}
