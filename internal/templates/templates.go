package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
)

//go:embed *.html
var files embed.FS

type Set struct {
	pages map[string]*template.Template
}

type pageView struct {
	Active string
	Data   any
	Stars  template.HTML
}

var homeStars = renderStars()

func Load() (*Set, error) {
	pageNames := []string{
		"home.html",
		"faqs.html",
		"information.html",
		"information-itinerary.html",
		"information-travel.html",
		"information-things-to-do.html",
		"contact-about-us.html",
		"contact-registry.html",
		"contact-get-in-touch.html",
		"rsvp.html",
		"rsvp-auth.html",
		"rsvp-thanks.html",
		"404.html",
		"500.html",
	}
	partials := []string{
		"layout.html",
		"site-sky.html",
		"site-nav.html",
		"home-hero.html",
		"home-hero-sky.html",
		"home-hero-mountain.html",
		"home-hero-moon.html",
	}
	funcs := template.FuncMap{"hasPrefix": strings.HasPrefix}
	pages := make(map[string]*template.Template, len(pageNames))
	for _, name := range pageNames {
		t, err := template.New(name).Funcs(funcs).ParseFS(files, append(partials, name)...)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		pages[name] = t
	}
	return &Set{pages: pages}, nil
}

func (s *Set) HomeStars() template.HTML {
	return homeStars
}

func (s *Set) Render(w http.ResponseWriter, r *http.Request, name string, data any) {
	s.RenderStatus(w, r, name, http.StatusOK, data)
}

func (s *Set) RenderStatus(w http.ResponseWriter, r *http.Request, name string, status int, data any) {
	t, ok := s.pages[name]
	if !ok {
		slog.Error("template not found", "name", name, "path", r.URL.Path)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	view := pageView{Active: r.URL.Path, Data: data, Stars: homeStars}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout.html", view); err != nil {
		slog.Error("template execute failed", "name", name, "path", r.URL.Path, "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}
