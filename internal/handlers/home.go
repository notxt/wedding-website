package handlers

import (
	"html/template"
	"net/http"

	"github.com/notxt/wedding-website/internal/templates"
)

func Home(t *templates.Set) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, r, "home.html", struct{ Stars template.HTML }{Stars: t.HomeStars()})
	}
}
