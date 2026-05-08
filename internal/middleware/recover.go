package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/notxt/wedding-website/internal/handlers"
	"github.com/notxt/wedding-website/internal/templates"
)

// Recover catches panics, logs them with a stack trace, and renders the 500
// page. The user never sees the panic message.
func Recover(t *templates.Set) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rv := recover(); rv != nil {
					slog.Error("panic recovered",
						"err", rv,
						"path", r.URL.Path,
						"stack", string(debug.Stack()),
					)
					handlers.RenderInternalError(t, w, r)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
