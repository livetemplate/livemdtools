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
        /* Theme Variables */
        :root {
            --bg-primary: #ffffff;
            --bg-secondary: linear-gradient(135deg, #f5f7fa 0%%, #e8ecf1 100%%);
            --text-primary: #333;
            --text-secondary: #555;
            --text-heading: #2c3e50;
            --border-color: #e1e4e8;
            --code-bg: #f4f4f4;
            --code-border: #e1e4e8;
            --pre-bg: #282c34;
            --pre-text: #abb2bf;
            --button-bg: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            --button-shadow: rgba(102, 126, 234, 0.3);
            --card-bg: #ffffff;
            --card-border: rgba(0,0,0,0.06);
            --card-shadow: rgba(0,0,0,0.08);
            --accent: #0066cc;
        }

        [data-theme="dark"] {
            --bg-primary: #1a1a1a;
            --bg-secondary: linear-gradient(135deg, #1a1a1a 0%%, #2d2d2d 100%%);
            --text-primary: #e0e0e0;
            --text-secondary: #b0b0b0;
            --text-heading: #f0f0f0;
            --border-color: #404040;
            --code-bg: #2d2d2d;
            --code-border: #404040;
            --pre-bg: #1e1e1e;
            --pre-text: #d4d4d4;
            --button-bg: linear-gradient(135deg, #4da6ff 0%%, #357abd 100%%);
            --button-shadow: rgba(77, 166, 255, 0.3);
            --card-bg: #242424;
            --card-border: rgba(255,255,255,0.1);
            --card-shadow: rgba(0,0,0,0.3);
            --accent: #4da6ff;
        }

        /* Theme transition */
        * {
            transition: background-color 0.3s ease, color 0.3s ease, border-color 0.3s ease;
        }

        /* Base styles */
        * {
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            max-width: 900px;
            margin: 0 auto;
            padding: 2rem;
            color: var(--text-primary);
            background: var(--bg-secondary);
            min-height: 100vh;
        }

        /* Typography */
        h1, h2, h3 {
            color: var(--text-heading);
            font-weight: 600;
            letter-spacing: -0.02em;
        }

        h1 {
            font-size: 2.5rem;
            margin-bottom: 1.5rem;
        }

        h2 {
            font-size: 1.75rem;
            margin-top: 2.5rem;
            margin-bottom: 1rem;
        }

        p {
            margin-bottom: 1.25rem;
            color: var(--text-secondary);
        }

        /* Code blocks */
        code {
            background: var(--code-bg);
            padding: 0.2rem 0.4rem;
            border-radius: 4px;
            font-size: 0.9em;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            border: 1px solid var(--code-border);
            color: var(--text-primary);
        }

        pre {
            background: var(--pre-bg);
            color: var(--pre-text);
            padding: 1.5rem;
            border-radius: 12px;
            overflow-x: auto;
            margin: 1.5rem 0;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            border: 1px solid var(--border-color);
        }

        pre code {
            background: none;
            border: none;
            padding: 0;
            color: inherit;
        }

        /* Interactive blocks */
        .livepage-wasm-block,
        .livepage-interactive-block {
            margin: 2rem 0;
            padding: 2rem;
            background: var(--card-bg);
            border-radius: 16px;
            box-shadow: 0 4px 16px var(--card-shadow);
            border: 1px solid var(--card-border);
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }

        .livepage-wasm-block:hover,
        .livepage-interactive-block:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 24px var(--card-shadow);
        }

        /* Buttons */
        button {
            background: var(--button-bg);
            color: white;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s ease;
            box-shadow: 0 2px 8px var(--button-shadow);
            margin: 0.25rem;
            font-family: inherit;
        }

        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px var(--button-shadow);
        }

        button:active {
            transform: translateY(0);
            box-shadow: 0 1px 4px var(--button-shadow);
        }

        /* Counter display */
        .counter-display {
            font-size: 3rem;
            font-weight: 700;
            text-align: center;
            margin: 2rem 0;
            padding: 2rem;
            background: linear-gradient(135deg, #f5f7fa 0%%, #ffffff 100%%);
            border-radius: 16px;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            border: 2px solid #e1e4e8;
        }

        .counter-display.positive {
            color: #10b981;
            border-color: #10b981;
            box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
        }

        .counter-display.negative {
            color: #ef4444;
            border-color: #ef4444;
            box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
        }

        .counter-display.zero {
            color: #6b7280;
            border-color: #d1d5db;
        }

        /* Number transition animation */
        @keyframes numberPulse {
            0%%, 100%% { transform: scale(1); }
            50%% { transform: scale(1.1); }
        }

        .counter-display.changed {
            animation: numberPulse 0.3s ease;
        }

        /* Button groups */
        .button-group {
            display: flex;
            justify-content: center;
            flex-wrap: wrap;
            gap: 0.5rem;
            margin: 1rem 0;
        }

        /* Responsive design */
        @media (max-width: 768px) {
            body {
                padding: 1rem;
            }

            h1 {
                font-size: 2rem;
            }

            h2 {
                font-size: 1.5rem;
            }

            .livepage-wasm-block,
            .livepage-interactive-block {
                padding: 1.5rem;
                border-radius: 12px;
            }

            .counter-display {
                font-size: 2.5rem;
                padding: 1.5rem;
            }

            button {
                padding: 0.625rem 1.25rem;
                font-size: 0.9rem;
            }
        }

        @media (max-width: 480px) {
            body {
                padding: 0.75rem;
            }

            h1 {
                font-size: 1.75rem;
            }

            .livepage-wasm-block,
            .livepage-interactive-block {
                padding: 1rem;
                margin: 1rem 0;
            }

            .counter-display {
                font-size: 2rem;
                padding: 1rem;
            }

            button {
                width: 100%%;
                margin: 0.25rem 0;
            }

            .button-group {
                flex-direction: column;
            }
        }

        /* Theme Toggle */
        .theme-toggle {
            position: fixed;
            top: 1rem;
            right: 1rem;
            z-index: 1000;
            display: flex;
            gap: 0.5rem;
            background: var(--card-bg);
            padding: 0.5rem;
            border-radius: 8px;
            box-shadow: 0 2px 8px var(--card-shadow);
            border: 1px solid var(--card-border);
        }

        .theme-toggle button {
            background: transparent;
            border: 1px solid var(--border-color);
            color: var(--text-primary);
            padding: 0.5rem;
            margin: 0;
            border-radius: 6px;
            font-size: 1.2rem;
            min-width: 2.5rem;
            box-shadow: none;
        }

        .theme-toggle button:hover {
            background: var(--code-bg);
            transform: none;
            box-shadow: none;
        }

        .theme-toggle button.active {
            background: var(--accent);
            color: white;
            border-color: var(--accent);
        }

        .theme-toggle button:active {
            transform: scale(0.95);
        }

        /* State Inspector */
        .state-inspector {
            position: fixed;
            bottom: 1rem;
            right: 1rem;
            width: 350px;
            max-height: 500px;
            background: var(--card-bg);
            border: 1px solid var(--card-border);
            border-radius: 12px;
            box-shadow: 0 4px 20px var(--card-shadow);
            z-index: 999;
            display: flex;
            flex-direction: column;
            font-size: 0.9rem;
        }

        .state-inspector.collapsed {
            max-height: 45px;
        }

        .state-inspector.hidden {
            display: none;
        }

        .inspector-header {
            padding: 0.75rem 1rem;
            background: var(--code-bg);
            border-bottom: 1px solid var(--border-color);
            border-radius: 12px 12px 0 0;
            display: flex;
            justify-content: space-between;
            align-items: center;
            cursor: pointer;
            user-select: none;
        }

        .inspector-header h4 {
            margin: 0;
            font-size: 0.9rem;
            color: var(--text-heading);
            font-weight: 600;
        }

        .inspector-header button {
            background: transparent;
            border: none;
            color: var(--text-primary);
            cursor: pointer;
            padding: 0.25rem;
            font-size: 1rem;
            line-height: 1;
            margin: 0;
            box-shadow: none;
        }

        .inspector-header button:hover {
            transform: none;
            box-shadow: none;
            opacity: 0.7;
        }

        .inspector-content {
            padding: 1rem;
            overflow-y: auto;
            flex: 1;
        }

        .state-section, .ws-section {
            margin-bottom: 1rem;
        }

        .state-section h5, .ws-section h5 {
            margin: 0 0 0.5rem 0;
            font-size: 0.85rem;
            color: var(--text-secondary);
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .state-json {
            background: var(--pre-bg);
            color: var(--pre-text);
            padding: 0.75rem;
            border-radius: 6px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.85rem;
            overflow-x: auto;
            white-space: pre;
            line-height: 1.4;
        }

        .ws-log {
            max-height: 150px;
            overflow-y: auto;
        }

        .ws-message {
            padding: 0.5rem;
            margin-bottom: 0.25rem;
            background: var(--code-bg);
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.8rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .ws-message.sent {
            border-left: 3px solid #4da6ff;
        }

        .ws-message.received {
            border-left: 3px solid #10b981;
        }

        .ws-message .action {
            flex: 1;
            color: var(--text-primary);
        }

        .ws-message .latency {
            color: var(--text-secondary);
            font-size: 0.75rem;
        }

        .inspector-toggle {
            position: fixed;
            bottom: 1rem;
            right: 1rem;
            background: var(--accent);
            color: white;
            border: none;
            padding: 0.75rem 1rem;
            border-radius: 8px;
            cursor: pointer;
            box-shadow: 0 2px 8px var(--card-shadow);
            font-size: 0.85rem;
            font-weight: 500;
            z-index: 998;
        }

        .inspector-toggle:hover {
            opacity: 0.9;
            transform: translateY(-2px);
        }

        @media (max-width: 768px) {
            .state-inspector {
                width: calc(100%% - 2rem);
                max-height: 400px;
                bottom: 0.5rem;
                right: 0.5rem;
                left: 0.5rem;
            }
        }
    </style>
</head>
<body>
    <!-- Theme Toggle -->
    <div class="theme-toggle">
        <button id="theme-light" title="Light theme" aria-label="Light theme">☀️</button>
        <button id="theme-dark" title="Dark theme" aria-label="Dark theme">🌙</button>
        <button id="theme-auto" title="Auto theme (system preference)" aria-label="Auto theme">🌓</button>
    </div>

    <!-- State Inspector Toggle -->
    <button class="inspector-toggle hidden" id="inspector-toggle" aria-label="Toggle state inspector">
        🔍 Inspector
    </button>

    <!-- State Inspector -->
    <div class="state-inspector hidden" id="state-inspector">
        <div class="inspector-header" id="inspector-header">
            <h4>🔍 State Inspector</h4>
            <button id="inspector-close" aria-label="Close inspector">✕</button>
        </div>
        <div class="inspector-content">
            <div class="state-section">
                <h5>Current State</h5>
                <div class="state-json" id="state-json">{}</div>
            </div>
            <div class="ws-section">
                <h5>WebSocket Log</h5>
                <div class="ws-log" id="ws-log">
                    <div style="color: var(--text-secondary); font-size: 0.8rem; text-align: center; padding: 1rem;">
                        No messages yet
                    </div>
                </div>
            </div>
        </div>
    </div>

    %s

    <script>
        // Theme management
        (function() {
            const STORAGE_KEY = 'livepage-theme';
            const html = document.documentElement;

            // Get current theme from localStorage or default to 'auto'
            function getStoredTheme() {
                return localStorage.getItem(STORAGE_KEY) || 'auto';
            }

            // Get system preference
            function getSystemTheme() {
                return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
            }

            // Apply theme to HTML element
            function applyTheme(theme) {
                const effectiveTheme = theme === 'auto' ? getSystemTheme() : theme;
                html.setAttribute('data-theme', effectiveTheme);

                // Update button states
                document.querySelectorAll('.theme-toggle button').forEach(btn => {
                    btn.classList.remove('active');
                });
                const activeBtn = document.getElementById('theme-' + theme);
                if (activeBtn) {
                    activeBtn.classList.add('active');
                }
            }

            // Set and save theme
            function setTheme(theme) {
                localStorage.setItem(STORAGE_KEY, theme);
                applyTheme(theme);
            }

            // Initialize theme on page load (before paint to avoid flash)
            const storedTheme = getStoredTheme();
            applyTheme(storedTheme);

            // Listen for system theme changes when in auto mode
            window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
                if (getStoredTheme() === 'auto') {
                    applyTheme('auto');
                }
            });

            // Add click handlers after DOM is ready
            window.addEventListener('DOMContentLoaded', () => {
                document.getElementById('theme-light').addEventListener('click', () => setTheme('light'));
                document.getElementById('theme-dark').addEventListener('click', () => setTheme('dark'));
                document.getElementById('theme-auto').addEventListener('click', () => setTheme('auto'));

                // Keyboard shortcut: Ctrl+Shift+D
                document.addEventListener('keydown', (e) => {
                    if (e.ctrlKey && e.shiftKey && e.key === 'D') {
                        e.preventDefault();
                        const current = getStoredTheme();
                        const next = current === 'light' ? 'dark' : current === 'dark' ? 'auto' : 'light';
                        setTheme(next);
                    }
                });
            });
        })();
    </script>

    <script>
        // State Inspector
        (function() {
            const inspector = document.getElementById('state-inspector');
            const toggle = document.getElementById('inspector-toggle');
            const closeBtn = document.getElementById('inspector-close');
            const header = document.getElementById('inspector-header');
            const stateJson = document.getElementById('state-json');
            const wsLog = document.getElementById('ws-log');

            let isCollapsed = false;
            let messageLog = [];
            const MAX_MESSAGES = 10;

            // Show inspector toggle when page has interactive blocks
            window.addEventListener('DOMContentLoaded', () => {
                const hasInteractive = document.querySelector('.livepage-interactive-block') !== null;
                if (hasInteractive) {
                    toggle.classList.remove('hidden');
                }
            });

            // Toggle inspector visibility
            toggle.addEventListener('click', () => {
                inspector.classList.toggle('hidden');
                if (!inspector.classList.contains('hidden')) {
                    toggle.style.display = 'none';
                }
            });

            closeBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                inspector.classList.add('hidden');
                toggle.style.display = 'block';
            });

            // Collapse/expand on header click
            header.addEventListener('click', () => {
                isCollapsed = !isCollapsed;
                inspector.classList.toggle('collapsed', isCollapsed);
            });

            // Update state display
            function updateState(state) {
                try {
                    const formatted = JSON.stringify(state, null, 2);
                    stateJson.textContent = formatted;
                } catch (e) {
                    stateJson.textContent = '{}';
                }
            }

            // Add WebSocket message to log
            function logMessage(direction, action, latency) {
                const msg = {
                    direction, // 'sent' or 'received'
                    action,
                    latency: latency || null,
                    timestamp: Date.now()
                };

                messageLog.unshift(msg);
                if (messageLog.length > MAX_MESSAGES) {
                    messageLog = messageLog.slice(0, MAX_MESSAGES);
                }

                renderMessageLog();
            }

            // Render message log
            function renderMessageLog() {
                if (messageLog.length === 0) {
                    wsLog.innerHTML = '<div style="color: var(--text-secondary); font-size: 0.8rem; text-align: center; padding: 1rem;">No messages yet</div>';
                    return;
                }

                wsLog.innerHTML = messageLog.map(msg => {
                    const arrow = msg.direction === 'sent' ? '→' : '←';
                    const latencyStr = msg.latency ? '(' + msg.latency + 'ms)' : '';
                    return '<div class="ws-message ' + msg.direction + '">' +
                        '<span class="action">' + arrow + ' ' + msg.action + '</span>' +
                        '<span class="latency">' + latencyStr + '</span>' +
                        '</div>';
                }).join('');
            }

            // Hook into WebSocket messages
            // We'll intercept messages by extending the WebSocket prototype
            const originalSend = WebSocket.prototype.send;
            const originalAddEventListener = WebSocket.prototype.addEventListener;

            const pendingMessages = new Map();

            WebSocket.prototype.send = function(data) {
                try {
                    const parsed = JSON.parse(data);
                    if (parsed.action) {
                        const msgId = parsed.action + '-' + Date.now();
                        pendingMessages.set(parsed.action, Date.now());
                        logMessage('sent', parsed.action);
                    }
                } catch (e) {
                    // Not JSON or no action
                }
                return originalSend.call(this, data);
            };

            WebSocket.prototype.addEventListener = function(event, listener, ...args) {
                if (event === 'message') {
                    const wrappedListener = function(e) {
                        try {
                            const data = JSON.parse(e.data);
                            if (data.state) {
                                updateState(data.state);
                            }
                            if (data.action) {
                                const sentTime = pendingMessages.get(data.action);
                                const latency = sentTime ? Date.now() - sentTime : null;
                                logMessage('received', 'update', latency);
                                if (sentTime) {
                                    pendingMessages.delete(data.action);
                                }
                            }
                        } catch (err) {
                            // Not JSON or missing fields
                        }
                        return listener.call(this, e);
                    };
                    return originalAddEventListener.call(this, event, wrappedListener, ...args);
                }
                return originalAddEventListener.call(this, event, listener, ...args);
            };
        })();
    </script>

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
