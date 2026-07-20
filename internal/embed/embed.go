package embedfs

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// Dist is the statically exported Next.js frontend.
//
//go:embed all:dist
var Dist embed.FS

// Handler serves the embedded SPA.
// Missing paths fall back to index.html for client-side routes.
func Handler() (http.Handler, error) {
	sub, err := fs.Sub(Dist, "dist")
	if err != nil {
		return nil, err
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			fileServer.ServeHTTP(w, r)
			return
		}

		if f, err := sub.Open(path); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		fileServer.ServeHTTP(w, r2)
	}), nil
}
