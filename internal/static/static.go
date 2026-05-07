package static

import (
	"fmt"
	"io/fs"
	"net/http"

	wedding "github.com/notxt/wedding-website"
)

func Handler() (http.Handler, error) {
	sub, err := fs.Sub(wedding.Assets, "static")
	if err != nil {
		return nil, fmt.Errorf("sub static fs: %w", err)
	}
	fileServer := http.FileServer(http.FS(sub))
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		fileServer.ServeHTTP(w, r)
	})
	return http.StripPrefix("/static/", h), nil
}
