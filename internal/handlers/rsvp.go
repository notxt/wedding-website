package handlers

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/notxt/wedding-website/internal/session"
	"github.com/notxt/wedding-website/internal/store"
	"github.com/notxt/wedding-website/internal/templates"
)

type AuthForm struct {
	Error string
}

type RSVPFormView struct {
	Attending      string
	PartyNames     string
	MealChoice     string
	DairyAllergy   string
	GlutenAllergy  string
	OtherAllergies string
	Errors         map[string]string
}

func RSVP(t *templates.Set, signer *session.Signer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rsvpAuthed(r, signer) {
			t.Render(w, r, "rsvp.html", RSVPFormView{})
			return
		}
		t.Render(w, r, "rsvp-auth.html", AuthForm{})
	}
}

func RSVPSubmit(t *templates.Set, st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		rsvp, errs := validateRSVP(r.PostForm)
		if len(errs) > 0 {
			view := RSVPFormView{
				Attending:      r.PostFormValue("attending"),
				PartyNames:     r.PostFormValue("party_names"),
				MealChoice:     r.PostFormValue("meal_choice"),
				DairyAllergy:   r.PostFormValue("dairy_allergy"),
				GlutenAllergy:  r.PostFormValue("gluten_allergy"),
				OtherAllergies: r.PostFormValue("other_allergies"),
				Errors:         errs,
			}
			t.RenderStatus(w, r, "rsvp.html", http.StatusBadRequest, view)
			return
		}
		if ip := clientIP(r); ip != "" {
			rsvp.RemoteAddr = &ip
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if _, err := st.InsertRSVP(ctx, rsvp); err != nil {
			slog.Error("insert rsvp failed", "err", err, "path", r.URL.Path)
			RenderInternalError(t, w, r)
			return
		}
		http.Redirect(w, r, "/rsvp/thanks", http.StatusSeeOther)
	}
}

func RSVPThanks(t *templates.Set) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Render(w, r, "rsvp-thanks.html", nil)
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
			t.RenderStatus(w, r, "rsvp-auth.html", http.StatusUnauthorized, AuthForm{Error: "Incorrect code. Please try again."})
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

func validateRSVP(form url.Values) (store.RSVP, map[string]string) {
	errs := map[string]string{}
	out := store.RSVP{}

	switch form.Get("attending") {
	case "yes":
		out.Attending = true
	case "no":
		out.Attending = false
	default:
		errs["attending"] = "Please tell us if you'll be attending."
	}

	out.PartyNames = strings.TrimSpace(form.Get("party_names"))
	if out.PartyNames == "" {
		errs["party_names"] = "Please enter the names of everyone in your party."
	}

	if out.Attending {
		switch meal := form.Get("meal_choice"); meal {
		case "chicken", "vegetarian":
			m := meal
			out.MealChoice = &m
		default:
			errs["meal_choice"] = "Please choose a meal preference."
		}
		out.DairyAllergy = form.Get("dairy_allergy") == "yes"
		out.GlutenAllergy = form.Get("gluten_allergy") == "yes"
		if other := strings.TrimSpace(form.Get("other_allergies")); other != "" {
			out.OtherAllergies = &other
		}
	}

	return out, errs
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	return r.RemoteAddr
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
