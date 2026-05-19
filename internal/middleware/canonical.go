package middleware

import (
	"net/http"
	"strings"
)

// CanonicalPath redirects /foo/ -> /foo (308 Permanent Redirect, preserving the
// method) for everything except the root path and /static/ assets, which keep
// their trailing slash.
func CanonicalPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/" || !strings.HasSuffix(p, "/") || strings.HasPrefix(p, "/static/") {
			next.ServeHTTP(w, r)
			return
		}
		target := strings.TrimRight(p, "/")
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusPermanentRedirect)
	})
}
