package handlers

import (
	"net/http"

	"github.com/notxt/wedding-website/internal/templates"
)

func NotFound(t *templates.Set) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.RenderStatus(w, r, "404.html", http.StatusNotFound, nil)
	}
}

func RenderInternalError(t *templates.Set, w http.ResponseWriter, r *http.Request) {
	t.RenderStatus(w, r, "500.html", http.StatusInternalServerError, nil)
}
