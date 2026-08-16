package embedfs

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/ijaidev/kraftui/log"
)

type notFoundWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *notFoundWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.wroteHeader = true
		if code == http.StatusOK {
			w.ResponseWriter.WriteHeader(http.StatusNotFound)
			return
		}
	}
	w.ResponseWriter.WriteHeader(code)
}

// isFile reports whether name exists in fsys and is a regular file.
func isFile(fsys fs.FS, name string) bool {
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()
	stat, err := f.Stat()
	return err == nil && !stat.IsDir()
}

// Handler serves the embedded Next.js static export.
func nextJsHandler(dist fs.FS) (http.Handler, error) {
	fileServer := http.FileServer(http.FS(dist))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := strings.Trim(r.URL.Path, "/")

		// 1. If path ends with .html, redirect to clean URL (/new.html -> /new, /index.html -> /)
		if target, found := strings.CutSuffix(cleanPath, ".html"); found {
			if target == "index" {
				target = ""
			}
			http.Redirect(w, r, "/"+target, http.StatusMovedPermanently)
			return
		}

		// 2. Root path or exact static asset (_next/..., favicon.ico)
		if cleanPath == "" || isFile(dist, cleanPath) {
			log.G().Debug("serving static assets", "path", cleanPath)
			fileServer.ServeHTTP(w, r)
			return
		}

		// 3. Next.js static page route (/new or /new/ -> new.html)
		if isFile(dist, cleanPath+".html") {
			log.G().Debug("serving static page route", "path", cleanPath)
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/" + cleanPath + ".html"
			fileServer.ServeHTTP(w, r2)
			return
		}

		log.G().Debug("serving not-found route", "path", cleanPath)

		// 4. Next.js 404 page fallback (with 404 status)
		if isFile(dist, "404.html") {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/404.html"
			fileServer.ServeHTTP(&notFoundWriter{ResponseWriter: w}, r2)
			return
		}

		http.NotFound(w, r)
	}), nil
}
