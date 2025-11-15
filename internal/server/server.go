package server

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/livetemplate/livepage"
	"github.com/livetemplate/livepage/internal/assets"
)

// Route represents a discovered page route.
type Route struct {
	Pattern  string         // URL pattern (e.g., "/counter")
	FilePath string         // Relative file path (e.g., "counter.md")
	Page     *livepage.Page // Parsed page
}

// Server is the livepage development server.
type Server struct {
	rootDir string
	routes  []*Route
	mu      sync.RWMutex
}

// New creates a new server for the given root directory.
func New(rootDir string) *Server {
	return &Server{
		rootDir: rootDir,
		routes:  make([]*Route, 0),
	}
}

// Discover scans the directory for .md files and creates routes.
func (s *Server) Discover() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.routes = make([]*Route, 0)

	err := filepath.WalkDir(s.rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			// Skip directories starting with _ or .
			name := d.Name()
			if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .md files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(s.rootDir, path)
		if err != nil {
			return err
		}

		// Skip files in _ directories
		if strings.Contains(relPath, "/_") || strings.HasPrefix(relPath, "_") {
			return nil
		}

		// Generate route pattern
		pattern := mdToPattern(relPath)

		// Parse the page
		page, err := livepage.ParseFile(path)
		if err != nil {
			log.Printf("Warning: Failed to parse %s: %v", relPath, err)
			return nil // Continue with other files
		}

		route := &Route{
			Pattern:  pattern,
			FilePath: relPath,
			Page:     page,
		}

		s.routes = append(s.routes, route)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	// Sort routes (index routes first)
	sortRoutes(s.routes)

	return nil
}

// Routes returns the discovered routes.
func (s *Server) Routes() []*Route {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.routes
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Serve WebSocket endpoint
	if r.URL.Path == "/ws" {
		s.serveWebSocket(w, r)
		return
	}

	// Serve assets
	if strings.HasPrefix(r.URL.Path, "/assets/") {
		s.serveAsset(w, r)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find matching route
	for _, route := range s.routes {
		if route.Pattern == r.URL.Path {
			s.servePage(w, r, route)
			return
		}
	}

	// No route found
	http.NotFound(w, r)
}

// serveWebSocket handles WebSocket connections for interactive blocks.
func (s *Server) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get the page from query parameter (for now, use first route)
	// TODO: Support multiple pages via query param or path
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.routes) == 0 {
		http.Error(w, "No pages available", http.StatusNotFound)
		return
	}

	// Use the first route's page for now
	route := s.routes[0]

	// Create WebSocket handler for this page
	wsHandler := NewWebSocketHandler(route.Page, true) // debug=true
	wsHandler.ServeHTTP(w, r)
}

// serveAsset serves embedded client assets.
func (s *Server) serveAsset(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/assets/")

	// Serve client JS
	if path == "livepage-client.js" {
		js, err := assets.GetClientJS()
		if err != nil {
			http.Error(w, "Asset not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(js)
		return
	}

	// Serve client CSS
	if path == "livepage-client.css" {
		css, err := assets.GetClientCSS()
		if err != nil {
			http.Error(w, "Asset not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/css")
		w.Write(css)
		return
	}

	http.NotFound(w, r)
}

// servePage serves a page.
func (s *Server) servePage(w http.ResponseWriter, r *http.Request, route *Route) {
	// For now, just serve the static HTML
	// TODO: Add WebSocket support for interactivity
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := s.renderPage(route.Page)
	w.Write([]byte(html))
}

// renderPage renders a page to HTML.
func (s *Server) renderPage(page *livepage.Page) string {
	// Render code blocks with metadata for client discovery
	content := s.renderContent(page)

	// Basic HTML wrapper with the static content
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="livepage-ws-url" content="ws://localhost:8080/ws">
    <meta name="livepage-debug" content="true">
    <title>%s</title>
    <link rel="stylesheet" href="/assets/livepage-client.css">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            max-width: 900px;
            margin: 0 auto;
            padding: 2rem;
            color: #333;
            background: #f9f9f9;
        }
        h1, h2, h3 { color: #2c3e50; }
        code {
            background: #f4f4f4;
            padding: 0.2rem 0.4rem;
            border-radius: 3px;
            font-size: 0.9em;
        }
        pre {
            background: #282c34;
            color: #abb2bf;
            padding: 1rem;
            border-radius: 8px;
            overflow-x: auto;
            margin: 1rem 0;
        }
        pre code {
            background: none;
            padding: 0;
            color: inherit;
        }
        .livepage-wasm-block,
        .livepage-interactive-block {
            margin: 2rem 0;
            padding: 1rem;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
    </style>
</head>
<body>
    %s
    <script src="/assets/livepage-client.js"></script>
</body>
</html>`, page.Title, content)

	return html
}

// renderContent renders the page content with code blocks
func (s *Server) renderContent(page *livepage.Page) string {
	content := page.StaticHTML

	// TODO: Enhance markdown parser to add data attributes to code blocks
	// For now, the client will need to discover blocks by parsing the HTML
	// In Phase 4.5, we'll improve this to inject proper data attributes during parsing

	return content
}

// mdToPattern converts a markdown file path to a URL pattern.
// Examples:
//   - "index.md" → "/"
//   - "counter.md" → "/counter"
//   - "tutorials/intro.md" → "/tutorials/intro"
//   - "tutorials/index.md" → "/tutorials/"
func mdToPattern(relPath string) string {
	// Remove .md extension
	path := strings.TrimSuffix(relPath, ".md")

	// Convert to URL path
	path = filepath.ToSlash(path)

	// Handle index files
	if path == "index" {
		return "/"
	}
	if strings.HasSuffix(path, "/index") {
		return "/" + strings.TrimSuffix(path, "index")
	}

	return "/" + path
}

// sortRoutes sorts routes with index routes first.
func sortRoutes(routes []*Route) {
	// Simple sort: / first, then /foo/, then /foo
	// This is a basic implementation; could be more sophisticated
	for i := 0; i < len(routes); i++ {
		for j := i + 1; j < len(routes); j++ {
			if shouldSwap(routes[i], routes[j]) {
				routes[i], routes[j] = routes[j], routes[i]
			}
		}
	}
}

func shouldSwap(a, b *Route) bool {
	// Root path comes first
	if a.Pattern == "/" {
		return false
	}
	if b.Pattern == "/" {
		return true
	}

	// Directory index paths come before other paths
	aIsIndex := strings.HasSuffix(a.Pattern, "/")
	bIsIndex := strings.HasSuffix(b.Pattern, "/")

	if aIsIndex && !bIsIndex {
		return false
	}
	if !aIsIndex && bIsIndex {
		return true
	}

	// Alphabetical otherwise
	return a.Pattern > b.Pattern
}
