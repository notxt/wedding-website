package handlers

import (
	"net/http"

	"github.com/notxt/wedding-website/internal/templates"
)

func Home(t *templates.Set) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, r, "home.html", nil)
	}
}
