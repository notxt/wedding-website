package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

//go:embed *.html
var files embed.FS

type Set struct {
	pages map[string]*template.Template
}

type pageView struct {
	Active string
	Data   any
}

func Load() (*Set, error) {
	pageNames := []string{"home.html", "faqs.html"}
	pages := make(map[string]*template.Template, len(pageNames))
	for _, name := range pageNames {
		t, err := template.ParseFS(files, "layout.html", name)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		pages[name] = t
	}
	return &Set{pages: pages}, nil
}

func (s *Set) Render(w http.ResponseWriter, r *http.Request, name string, data any) {
	t, ok := s.pages[name]
	if !ok {
		slog.Error("template not found", "name", name, "path", r.URL.Path)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	view := pageView{Active: r.URL.Path, Data: data}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout.html", view); err != nil {
		slog.Error("template execute failed", "name", name, "path", r.URL.Path, "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)
}
