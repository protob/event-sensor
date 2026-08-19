// Package spa serves the single-page app: the embedded build output, the fallback
// handler and nothing else. The embed lives here rather than beside main because
// go:embed cannot reach outside its own package directory, so Vite writes into this
// directory (see frontend/vite.config.ts).
package spa

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Mount serves the built SPA. Register it after /api: the fallback is a catch-all and
// would otherwise answer for a mistyped endpoint.
func Mount(r chi.Router) error {
	dist, err := frontendFS()
	if err != nil {
		return err
	}
	mount(r, dist)
	return nil
}

// mount takes the filesystem as an argument so the test can serve a fake one. Unknown
// paths fall back to index.html, which is what keeps a client-side route alive across a
// reload.
func mount(r chi.Router, dist fs.FS) {
	fileServer := http.FileServer(http.FS(dist))
	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		if _, err := fs.Stat(dist, strings.TrimPrefix(req.URL.Path, "/")); err == nil {
			fileServer.ServeHTTP(w, req)
			return
		}
		req.URL.Path = "/"
		fileServer.ServeHTTP(w, req)
	})
}
