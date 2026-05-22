package static

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"

	wedding "github.com/notxt/wedding-website"
)

func Handler() (http.Handler, error) {
	sub, err := fs.Sub(wedding.Assets, "static")
	if err != nil {
		return nil, fmt.Errorf("sub static fs: %w", err)
	}
	// In the dev container, never cache so edits show on a normal reload;
	// in any other environment keep a modest browser cache.
	cacheControl := "public, max-age=3600"
	if os.Getenv("WEDDING_DEV_CONTAINER") != "" {
		cacheControl = "no-store"
	}
	fileServer := http.FileServer(http.FS(sub))
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", cacheControl)
		fileServer.ServeHTTP(w, r)
	})
	return http.StripPrefix("/static/", h), nil
}
