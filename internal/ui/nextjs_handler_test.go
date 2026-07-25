package embedfs

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func createTestFixtures(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	// 1. index.html
	if err := os.WriteFile(filepath.Join(tempDir, "index.html"), []byte("<h1>Home</h1>"), 0644); err != nil {
		t.Fatalf("failed to write index.html: %v", err)
	}

	// 2. new.html
	if err := os.WriteFile(filepath.Join(tempDir, "new.html"), []byte("<h1>New Page</h1>"), 0644); err != nil {
		t.Fatalf("failed to write new.html: %v", err)
	}

	// 3. Subdirectory new/ to simulate Next.js route metadata directory
	if err := os.MkdirAll(filepath.Join(tempDir, "new"), 0755); err != nil {
		t.Fatalf("failed to create new/ dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "new", "metadata.txt"), []byte("meta"), 0644); err != nil {
		t.Fatalf("failed to write new/metadata.txt: %v", err)
	}

	// 4. 404.html
	if err := os.WriteFile(filepath.Join(tempDir, "404.html"), []byte("<h1>Custom 404</h1>"), 0644); err != nil {
		t.Fatalf("failed to write 404.html: %v", err)
	}

	// 5. Static asset /favicon.ico
	if err := os.WriteFile(filepath.Join(tempDir, "favicon.ico"), []byte("icon-bytes"), 0644); err != nil {
		t.Fatalf("failed to write favicon.ico: %v", err)
	}

	return tempDir
}

func TestNextJsHandlerWithFixtures(t *testing.T) {
	tempDir := createTestFixtures(t)
	distFS := os.DirFS(tempDir)

	handler, err := nextJsHandler(distFS)
	if err != nil {
		t.Fatalf("unexpected error creating handler: %v", err)
	}

	ts := httptest.NewServer(handler)
	defer ts.Close()

	t.Run("Root path / returns index.html", func(t *testing.T) {
		res, err := http.Get(ts.URL + "/")
		if err != nil {
			t.Fatalf("failed to GET /: %v", err)
		}
		body, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for /, got %d", res.StatusCode)
		}
		if string(body) != "<h1>Home</h1>" {
			t.Errorf("expected body <h1>Home</h1>, got %s", string(body))
		}
	})

	t.Run("Exact static asset /favicon.ico", func(t *testing.T) {
		res, err := http.Get(ts.URL + "/favicon.ico")
		if err != nil {
			t.Fatalf("failed to GET /favicon.ico: %v", err)
		}
		body, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for /favicon.ico, got %d", res.StatusCode)
		}
		if string(body) != "icon-bytes" {
			t.Errorf("expected body icon-bytes, got %s", string(body))
		}
	})

	t.Run("Route path /new returns new.html with 200 OK", func(t *testing.T) {
		res, err := http.Get(ts.URL + "/new")
		if err != nil {
			t.Fatalf("failed to GET /new: %v", err)
		}
		body, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for /new, got %d", res.StatusCode)
		}
		if string(body) != "<h1>New Page</h1>" {
			t.Errorf("expected body <h1>New Page</h1>, got %s", string(body))
		}
	})

	t.Run("Route path with trailing slash /new/ returns new.html with 200 OK", func(t *testing.T) {
		res, err := http.Get(ts.URL + "/new/")
		if err != nil {
			t.Fatalf("failed to GET /new/: %v", err)
		}
		body, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for /new/, got %d", res.StatusCode)
		}
		if string(body) != "<h1>New Page</h1>" {
			t.Errorf("expected body <h1>New Page</h1>, got %s", string(body))
		}
	})

	t.Run("Missing path returns 404.html with 404 Not Found status", func(t *testing.T) {
		res, err := http.Get(ts.URL + "/non-existent-route")
		if err != nil {
			t.Fatalf("failed to GET /non-existent-route: %v", err)
		}
		body, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 Not Found for missing path, got %d", res.StatusCode)
		}
		if string(body) != "<h1>Custom 404</h1>" {
			t.Errorf("expected body <h1>Custom 404</h1>, got %s", string(body))
		}
	})

	t.Run("Explicit .html request /new.html redirects (301) to /new", func(t *testing.T) {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		res, err := client.Get(ts.URL + "/new.html")
		if err != nil {
			t.Fatalf("failed to GET /new.html: %v", err)
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusMovedPermanently {
			t.Errorf("expected 301 Moved Permanently for /new.html, got %d", res.StatusCode)
		}
		if loc := res.Header.Get("Location"); loc != "/new" {
			t.Errorf("expected Location header /new, got %s", loc)
		}
	})

	t.Run("Explicit /index.html redirects (301) to /", func(t *testing.T) {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		res, err := client.Get(ts.URL + "/index.html")
		if err != nil {
			t.Fatalf("failed to GET /index.html: %v", err)
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusMovedPermanently {
			t.Errorf("expected 301 Moved Permanently for /index.html, got %d", res.StatusCode)
		}
		if loc := res.Header.Get("Location"); loc != "/" {
			t.Errorf("expected Location header /, got %s", loc)
		}
	})
}
