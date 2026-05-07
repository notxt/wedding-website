package handlers

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/notxt/wedding-website/internal/session"
	"github.com/notxt/wedding-website/internal/templates"
)

type AuthForm struct {
	Error string
}

func RSVP(t *templates.Set, signer *session.Signer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rsvpAuthed(r, signer) {
			t.Render(w, r, "rsvp.html", nil)
			return
		}
		t.Render(w, r, "rsvp-auth.html", AuthForm{})
	}
}

func RSVPAuthSubmit(t *templates.Set, signer *session.Signer, accessCode string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		code := r.PostFormValue("code")
		if subtle.ConstantTimeCompare([]byte(code), []byte(accessCode)) != 1 {
			w.WriteHeader(http.StatusUnauthorized)
			t.Render(w, r, "rsvp-auth.html", AuthForm{Error: "Incorrect code. Please try again."})
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "rsvp_auth",
			Value:    signer.Sign("ok"),
			Path:     "/rsvp",
			MaxAge:   30 * 24 * 60 * 60,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   requestIsHTTPS(r),
		})
		http.Redirect(w, r, "/rsvp", http.StatusSeeOther)
	}
}

func rsvpAuthed(r *http.Request, s *session.Signer) bool {
	c, err := r.Cookie("rsvp_auth")
	if err != nil {
		return false
	}
	return s.Verify(c.Value)
}

func requestIsHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	for _, part := range strings.Split(r.Header.Get("Forwarded"), ";") {
		if strings.EqualFold(strings.TrimSpace(part), "proto=https") {
			return true
		}
	}
	return false
}
