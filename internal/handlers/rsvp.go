package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
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
	GuestName         string
	PlusOneAllowed    bool
	Attending         string
	MealChoice        string
	DairyAllergy      string
	GlutenAllergy     string
	OtherAllergies    string
	PlusOneAttending  string
	PlusOneName       string
	PlusOneMealChoice string
	Errors            map[string]string
}

type RSVPConfirmationView struct {
	GuestName         string
	Attending         bool
	MealChoice        string
	DairyAllergy      bool
	GlutenAllergy     bool
	OtherAllergies    string
	PlusOneAttending  bool
	PlusOneName       string
	PlusOneMealChoice string
}

func RSVP(t *templates.Set, signer *session.Signer, st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guestID, ok := guestIDFromCookie(r, signer)
		if !ok {
			t.Render(w, r, "rsvp-auth.html", AuthForm{})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		guest, found, err := st.GetGuestByID(ctx, guestID)
		if err != nil {
			slog.Error("load guest failed", "err", err, "path", r.URL.Path)
			RenderInternalError(t, w, r)
			return
		}
		if !found {
			clearAuthCookie(w, r)
			t.Render(w, r, "rsvp-auth.html", AuthForm{})
			return
		}

		_, hasRSVP, err := st.GetRSVPByGuestID(ctx, guestID)
		if err != nil {
			slog.Error("load rsvp failed", "err", err, "path", r.URL.Path)
			RenderInternalError(t, w, r)
			return
		}
		if hasRSVP {
			http.Redirect(w, r, "/rsvp/thanks", http.StatusSeeOther)
			return
		}
		view := RSVPFormView{GuestName: guest.FirstName, PlusOneAllowed: guest.PlusOneAllowed}
		if guest.PlusOneName != nil {
			view.PlusOneName = *guest.PlusOneName
		}
		t.Render(w, r, "rsvp.html", view)
	}
}

func RSVPSubmit(t *templates.Set, signer *session.Signer, st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guestID, ok := guestIDFromCookie(r, signer)
		if !ok {
			http.Redirect(w, r, "/rsvp", http.StatusSeeOther)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		guest, found, err := st.GetGuestByID(ctx, guestID)
		if err != nil {
			slog.Error("load guest failed", "err", err, "path", r.URL.Path)
			RenderInternalError(t, w, r)
			return
		}
		if !found {
			http.Redirect(w, r, "/rsvp", http.StatusSeeOther)
			return
		}

		// RSVPs are submit-once: a guest who has already responded can't change it.
		_, hasRSVP, err := st.GetRSVPByGuestID(ctx, guestID)
		if err != nil {
			slog.Error("load rsvp failed", "err", err, "path", r.URL.Path)
			RenderInternalError(t, w, r)
			return
		}
		if hasRSVP {
			http.Redirect(w, r, "/rsvp/thanks", http.StatusSeeOther)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		rsvp, view, errs := validateRSVP(r.PostForm, guest)
		if len(errs) > 0 {
			view.Errors = errs
			t.RenderStatus(w, r, "rsvp.html", http.StatusBadRequest, view)
			return
		}
		rsvp.GuestID = guestID
		if ip := clientIP(r); ip != "" {
			rsvp.RemoteAddr = &ip
		}
		// inserted == false means a concurrent submit beat us in; either way the
		// guest's response is recorded, so send them to the confirmation.
		if _, _, err := st.InsertRSVP(ctx, rsvp); err != nil {
			slog.Error("insert rsvp failed", "err", err, "path", r.URL.Path)
			RenderInternalError(t, w, r)
			return
		}
		http.Redirect(w, r, "/rsvp/thanks", http.StatusSeeOther)
	}
}

func RSVPThanks(t *templates.Set, signer *session.Signer, st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guestID, ok := guestIDFromCookie(r, signer)
		if !ok {
			http.Redirect(w, r, "/rsvp", http.StatusSeeOther)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		rsvp, hasRSVP, err := st.GetRSVPByGuestID(ctx, guestID)
		if err != nil {
			slog.Error("load rsvp failed", "err", err, "path", r.URL.Path)
			RenderInternalError(t, w, r)
			return
		}
		if !hasRSVP {
			http.Redirect(w, r, "/rsvp", http.StatusSeeOther)
			return
		}
		guest, found, err := st.GetGuestByID(ctx, guestID)
		if err != nil {
			slog.Error("load guest failed", "err", err, "path", r.URL.Path)
			RenderInternalError(t, w, r)
			return
		}
		if !found {
			clearAuthCookie(w, r)
			http.Redirect(w, r, "/rsvp", http.StatusSeeOther)
			return
		}
		t.Render(w, r, "rsvp-thanks.html", confirmationView(guest, rsvp))
	}
}

func RSVPAuthSubmit(t *templates.Set, signer *session.Signer, st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		email := normalizeEmail(r.PostFormValue("email"))
		if email == "" {
			t.RenderStatus(w, r, "rsvp-auth.html", http.StatusUnauthorized,
				AuthForm{Error: "Please enter the email address from your invitation."})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		guest, found, err := st.GetGuestByEmail(ctx, email)
		if err != nil {
			slog.Error("guest lookup failed", "err", err, "path", r.URL.Path)
			RenderInternalError(t, w, r)
			return
		}
		if !found {
			t.RenderStatus(w, r, "rsvp-auth.html", http.StatusUnauthorized,
				AuthForm{Error: "We couldn't find that email on our guest list. Double-check the address, or get in touch and we'll sort it out."})
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "rsvp_auth",
			Value:    signer.Sign(strconv.FormatInt(guest.ID, 10)),
			Path:     "/rsvp",
			MaxAge:   30 * 24 * 60 * 60,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   requestIsHTTPS(r),
		})
		http.Redirect(w, r, "/rsvp", http.StatusSeeOther)
	}
}

// validateRSVP builds the store row and the re-render view from the posted form.
// Plus-one fields are only honored when the guest is allowed one and is attending.
func validateRSVP(form url.Values, guest store.Guest) (store.RSVP, RSVPFormView, map[string]string) {
	errs := map[string]string{}
	out := store.RSVP{}
	view := RSVPFormView{
		GuestName:         guest.FirstName,
		PlusOneAllowed:    guest.PlusOneAllowed,
		Attending:         form.Get("attending"),
		MealChoice:        form.Get("meal_choice"),
		DairyAllergy:      form.Get("dairy_allergy"),
		GlutenAllergy:     form.Get("gluten_allergy"),
		OtherAllergies:    form.Get("other_allergies"),
		PlusOneAttending:  form.Get("plus_one_attending"),
		PlusOneName:       form.Get("plus_one_name"),
		PlusOneMealChoice: form.Get("plus_one_meal_choice"),
	}

	switch form.Get("attending") {
	case "yes":
		out.Attending = true
	case "no":
		out.Attending = false
	default:
		errs["attending"] = "Please tell us if you'll be attending."
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

		if guest.PlusOneAllowed && form.Get("plus_one_attending") == "yes" {
			out.PlusOneAttending = true
			if name := strings.TrimSpace(form.Get("plus_one_name")); name != "" {
				out.PlusOneName = &name
			} else {
				errs["plus_one_name"] = "Please enter your guest's name."
			}
			switch meal := form.Get("plus_one_meal_choice"); meal {
			case "chicken", "vegetarian":
				m := meal
				out.PlusOneMealChoice = &m
			default:
				errs["plus_one_meal_choice"] = "Please choose a meal for your guest."
			}
		}
	}

	out.PartyNames = strings.TrimSpace(guest.FirstName + " " + guest.LastName)
	if out.PlusOneAttending && out.PlusOneName != nil {
		out.PartyNames += " + " + *out.PlusOneName
	}

	return out, view, errs
}

// confirmationView resolves a stored RSVP into display-ready fields for the
// read-only recap on the thanks page (templates can't safely print *string).
func confirmationView(guest store.Guest, r store.RSVP) RSVPConfirmationView {
	v := RSVPConfirmationView{
		GuestName:        guest.FirstName,
		Attending:        r.Attending,
		DairyAllergy:     r.DairyAllergy,
		GlutenAllergy:    r.GlutenAllergy,
		PlusOneAttending: r.PlusOneAttending,
	}
	if r.MealChoice != nil {
		v.MealChoice = mealLabel(*r.MealChoice)
	}
	if r.OtherAllergies != nil {
		v.OtherAllergies = *r.OtherAllergies
	}
	if r.PlusOneName != nil {
		v.PlusOneName = *r.PlusOneName
	}
	if r.PlusOneMealChoice != nil {
		v.PlusOneMealChoice = mealLabel(*r.PlusOneMealChoice)
	}
	return v
}

// mealLabel renders a stored meal value ("chicken") for display ("Chicken").
func mealLabel(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
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

func guestIDFromCookie(r *http.Request, s *session.Signer) (int64, bool) {
	c, err := r.Cookie("rsvp_auth")
	if err != nil || !s.Verify(c.Value) {
		return 0, false
	}
	marker := c.Value[:strings.LastIndex(c.Value, ".")]
	id, err := strconv.ParseInt(marker, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func clearAuthCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "rsvp_auth",
		Value:    "",
		Path:     "/rsvp",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   requestIsHTTPS(r),
	})
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
