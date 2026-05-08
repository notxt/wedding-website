package middleware

import (
	"net/http"

	"github.com/notxt/wedding-website/internal/session"
)

type Auth struct {
	Signer *session.Signer
}

func (a *Auth) RequireRSVPAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("rsvp_auth")
		if err != nil || !a.Signer.Verify(c.Value) {
			http.Redirect(w, r, "/rsvp", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
