package handlers

import (
	"net/http"

	"github.com/notxt/wedding-website/internal/templates"
)

func Page(t *templates.Set, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, r, name, nil)
	}
}
