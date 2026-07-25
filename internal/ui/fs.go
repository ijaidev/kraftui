package embedfs

import (
	"embed"
	"io/fs"
	"net/http"
)

// Dist is the statically exported Next.js frontend.
//
//go:embed all:dist
var uiFS embed.FS

func UiHandler() (http.Handler, error) {
	dist, err := fs.Sub(uiFS, "dist")
	if err != nil {
		return nil, err
	}
	return nextJsHandler(dist)
}
