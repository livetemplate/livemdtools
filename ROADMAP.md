# Tinkerdown Roadmap: Micro Apps & Tooling Platform

## Vision

Transform Tinkerdown into a **full-featured platform for building micro apps and internal tooling** using markdown as the primary interface.

**Target Users:**
- Developers building internal tools, dashboards, and admin panels
- Technical writers creating interactive documentation
- Teams needing quick data-driven apps without full-stack complexity
- AI systems generating functional apps from natural language

---

## Priority Framework

| Level | Criteria | Focus |
|-------|----------|-------|
| **P0** | Blocks core functionality or causes data loss | Immediate |
| **P1** | High-impact features enabling new use cases | Near-term |
| **P2** | Developer experience and productivity | Medium-term |
| **P3** | Polish, optimization, and edge cases | Ongoing |

---

## Phase 1: Zero-Template Development (P0)

### Value Proposition
> "Build interactive apps without learning Go templates"

The biggest barrier to adoption is requiring users to learn Go's `html/template` syntax. This phase introduces `lvt-*` attributes as **compile-time sugar** that transforms to Go templates during parsing. Server-side rendering remains the core—the client stays simple.

### Implementation Approach

**Where:** Core LiveTemplate library (not Tinkerdown-specific)

This ensures the same `lvt-*` attributes work for both:
- **Full apps** built directly with LiveTemplate
- **Micro apps** built with Tinkerdown

```
lvt-* attributes in HTML
        ↓ (LiveTemplate core: parse time)
Transform to Go templates
        ↓
Server-side rendering (unchanged)
        ↓
HTML to client
```

**Benefits:**
- Single implementation in core library
- Consistent syntax across full and micro apps
- No new rendering engine
- Client remains lightweight
- SSR advantages preserved (security, performance, no JS required)

### Migration: Tinkerdown → Core

These attributes are currently in Tinkerdown but belong in LiveTemplate core:

**Client-side (move from `tinkerdown-client` to `@livetemplate/client`):**
- `lvt-click`, `lvt-submit`, `lvt-change` - Event handlers
- `lvt-data-*` - Data passing
- `lvt-confirm` - Confirmation dialogs
- `lvt-reset-on:{event}` - Lifecycle hooks

**Server-side (implement in LiveTemplate core):**
- `lvt-for`, `lvt-if`, `lvt-text` - Template sugar (new)
- `lvt-value`, `lvt-label` - Select bindings (move from Tinkerdown)
- `lvt-checked`, `lvt-disabled` - Boolean attributes (new)

**Stay in Tinkerdown (micro-app specific):**
- `lvt-source` - Data source binding (ties to Tinkerdown's source system)
- `lvt-columns`, `lvt-actions` - Auto-table generation

---

### 1.1 Auto-Rendering Components

**Problem:** Users must write `{{range .Data}}...{{end}}` loops manually.

**Solution:** HTML elements that auto-render from data sources:

```html
<!-- Instead of writing Go template loops -->
<table lvt-source="tasks" lvt-columns="done,text,priority">
  <!-- Auto-generates thead, tbody, and rows -->
</table>

<ul lvt-source="items" lvt-template="card">
  <!-- Auto-generates list items -->
</ul>

<select lvt-source="categories" lvt-value="id" lvt-label="name">
  <!-- Auto-populates options -->
</select>
```

**Work Required:**
- [ ] `<table lvt-source>` auto-generates full table from data
- [ ] `<ul/ol lvt-source>` auto-generates list items
- [ ] `<select lvt-source>` already works - document it
- [ ] `lvt-template` attribute for card/row/custom layouts
- [ ] `lvt-columns` for table column selection and ordering
- [ ] `lvt-empty` for empty state message

**Impact:** 80% of apps need no template knowledge

---

### 1.2 Declarative Attributes (Alpine.js-style)

**Problem:** Go template syntax `{{if .Done}}checked{{end}}` is unfamiliar.

**Solution:** HTML attributes that express logic declaratively:

```html
<!-- Instead of Go templates -->
<tr lvt-for="item in tasks">
  <td lvt-text="item.text"></td>
  <td lvt-if="item.done">✓</td>
  <td lvt-class="{'completed': item.done}"></td>
  <input type="checkbox" lvt-checked="item.done">
  <button lvt-click="Delete" lvt-data-id="item.id">Delete</button>
</tr>

<!-- Conditionals -->
<div lvt-if="error" class="error" lvt-text="error"></div>
<div lvt-if="!data.length">No items yet</div>
```

**Transformations:**
```
lvt-for="item in tasks"     → {{range .tasks}}...{{end}}
lvt-text="item.text"        → {{.text}}
lvt-if="item.done"          → {{if .done}}...{{end}}
lvt-checked="item.done"     → {{if .done}}checked{{end}}
lvt-class="done: item.done" → class="{{if .done}}done{{end}}"
```

**Work Required (in LiveTemplate core):**
- [ ] `lvt-for="item in source"` - Loop over data
- [ ] `lvt-text="field"` - Set text content
- [ ] `lvt-if="condition"` - Conditional rendering
- [ ] `lvt-checked`, `lvt-disabled`, `lvt-selected` - Boolean attributes
- [ ] `lvt-class="name: condition"` - Dynamic classes
- [ ] Attribute transformer in LiveTemplate's template processing

**Impact:** Familiar syntax for frontend developers, zero runtime overhead, works in both full and micro apps

---

### 1.3 Form Auto-Binding

**Problem:** Forms require manual wiring of inputs to actions.

**Solution:** Smart form components that infer behavior:

```html
<!-- Auto-generates Add form from source schema -->
<form lvt-source="tasks" lvt-action="Add">
  <!-- Auto-creates inputs based on table columns -->
</form>

<!-- Or explicit but simple -->
<form lvt-submit="Add" lvt-source="tasks">
  <input lvt-field="text" placeholder="Task...">
  <select lvt-field="priority" lvt-options="low,medium,high">
  <button type="submit">Add</button>
</form>
```

**Work Required:**
- [ ] `lvt-field` auto-binds input name and validation
- [ ] `lvt-options` for simple select options
- [ ] Form schema inference from SQLite tables
- [ ] Built-in validation display

**Impact:** CRUD apps in minutes without template code

---

### 1.4 Markdown-Native Data Binding

**Problem:** Users want to stay in markdown, not write HTML.

**Solution:** Extend markdown syntax for data display:

```markdown
## Tasks

<!-- Markdown table that binds to source -->
| Done | Task | Priority |
|------|------|----------|
{tasks}

<!-- Or a simple list -->
- {tasks: text} ({tasks: priority})
```

**Work Required:**
- [ ] `{source}` syntax in markdown tables
- [ ] `{source: field}` for inline field access
- [ ] Automatic table generation from source
- [ ] List binding syntax

**Impact:** True markdown-first development

---

## Phase 2: Stability & Performance (P0)

### Value Proposition
> "Make what exists work reliably in production"

### 2.1 Data Source Error Handling
**Files:** `internal/source/*.go`, `internal/runtime/state.go`

**Current State:** Errors can crash or hang; no retry logic; silent failures possible.

**Work Required:**
- [ ] Unified error types for all sources
- [ ] Retry with exponential backoff for transient failures
- [ ] Circuit breaker for repeatedly failing sources
- [ ] User-friendly error messages in templates (`.Error` field rendering)
- [ ] Timeout configuration per source

**Impact:** Production reliability for data-driven apps

---

### 1.2 Source Caching Layer
**Current State:** Every page view refetches all data.

**Work Required:**
- [ ] Cache configuration per source:
  ```yaml
  sources:
    users:
      type: rest
      url: https://api.example.com/users
      cache:
        ttl: 5m
        strategy: stale-while-revalidate
  ```
- [ ] In-memory cache with TTL
- [ ] Cache invalidation on write operations
- [ ] Manual cache clear via `Refresh` action

**Impact:** 10x faster page loads; reduced API costs; better UX

---

### 1.3 Multi-Page WebSocket Support
**Files:** `internal/server/server.go`

**Current State:** WebSocket handler only serves first route - multi-page sites limited.

**Work Required:**
- [ ] Accept page identifier via query param or path
- [ ] Route WebSocket messages to correct page's state
- [ ] Handle page transitions gracefully
- [ ] Clean up state on page navigation

**Impact:** Enables documentation sites with interactive examples on every page

---

## Phase 3: Developer Experience (P1)

### Value Proposition
> "Reduce time from idea to working app by 10x"

### 2.1 Enhanced CLI Scaffolding
**Files:** `cmd/tinkerdown/commands/new.go`

**Current State:** `new` command creates minimal template only.

**Work Required:**
- [ ] Add `--template` flag with options:
  - `basic` - Minimal with one source
  - `todo` - SQLite CRUD with toggle/delete
  - `dashboard` - Multi-source data display
  - `form` - Contact form with SQLite persistence
  - `api-explorer` - REST source with refresh
  - `cli-wrapper` - Exec source with argument form
  - `wasm-source` - Template for building custom WASM sources
- [ ] Generate sample data files for each template
- [ ] Include inline documentation comments

**Impact:** 5-minute start to working prototype

---

### 2.2 Expanded Validation
**Files:** `cmd/tinkerdown/commands/validate.go`

**Current State:** Only validates markdown parsing and Mermaid syntax.

**Work Required:**
- [ ] Validate source references exist in config
- [ ] Check `lvt-*` attributes have valid values
- [ ] Verify source types are valid (exec, pg, rest, json, csv, markdown, sqlite, wasm)
- [ ] Warn on unused source definitions
- [ ] Validate WASM module paths exist
- [ ] Type-check common template patterns

**Impact:** Catch errors at write-time, not runtime

---

### 2.3 Debug Mode & Logging
**Files:** `internal/server/server.go`

**Work Required:**
- [ ] `--debug` / `--verbose` CLI flags
- [ ] Structured JSON logging option
- [ ] Request correlation IDs
- [ ] WebSocket message logging (with sensitive data redaction)
- [ ] State change logging
- [ ] Source fetch timing

**Impact:** 10x faster debugging of production issues

---

### 2.4 Hot Reload for Configuration
**Current State:** Config changes require server restart.

**Work Required:**
- [ ] Watch `tinkerdown.yaml` for changes
- [ ] Reload sources without dropping WebSocket connections
- [ ] Notify connected clients of config reload
- [ ] Support frontmatter changes via file watcher

**Impact:** Faster iteration on source configuration

---

### 2.5 WASM Source Development Kit
**New Feature** - Critical for ecosystem growth

**Work Required:**
- [ ] `tinkerdown wasm init <name>` - Scaffold new WASM source
- [ ] `tinkerdown wasm build` - Compile TinyGo source to WASM
- [ ] `tinkerdown wasm test` - Test WASM module locally
- [ ] Documentation for WASM interface contract
- [ ] Example sources: GitHub API, Notion, Airtable

**Impact:** Enable community source contributions

---

## Phase 4: Data Ecosystem (P1)

### Value Proposition
> "Connect to any data source in minutes, not days"

### 3.1 GraphQL Source
**Work Required:**
- [ ] New source type: `graphql`
- [ ] Config: `url`, `query`, `variables`
- [ ] Authentication headers
- [ ] Auto-flatten nested response
- [ ] Support for mutations via write operations

**Impact:** Modern API ecosystem support

---

### 3.2 MongoDB Source
**Work Required:**
- [ ] New source type: `mongodb`
- [ ] Config: `uri`, `database`, `collection`, `filter`
- [ ] Pure Go driver (no cgo)
- [ ] CRUD operations support

**Impact:** NoSQL database support

---

### 3.3 Source Composition
**New Feature**

**Work Required:**
- [ ] Computed sources that transform other sources:
  ```yaml
  sources:
    users:
      type: rest
      url: https://api.example.com/users
    active_users:
      type: computed
      from: users
      filter: "status == 'active'"
      sort: "name asc"
  ```
- [ ] Join sources on common fields
- [ ] Aggregation (count, sum, avg)

**Impact:** Complex data apps without custom code

---

### 3.4 Webhook Source
**Work Required:**
- [ ] Source that receives HTTP POST
- [ ] Store latest N events
- [ ] Trigger UI update on new data
- [ ] Optional signature verification (Stripe, GitHub)

**Impact:** Real-time integrations (webhooks, events)

---

### 3.5 S3/Cloud Storage Source
**Work Required:**
- [ ] New source type: `s3`
- [ ] List objects, read JSON/CSV files
- [ ] Support for GCS, Azure Blob (compatible APIs)
- [ ] Credentials via environment variables

**Impact:** Cloud-native data access

---

## Phase 5: Production Readiness (P1)

### Value Proposition
> "Deploy with confidence to real users"

### 4.1 Authentication Middleware
**Work Required:**
- [ ] Built-in auth strategies:
  - API key (header-based)
  - Basic auth
  - OAuth2 (Google, GitHub)
  - Custom JWT validation
- [ ] Per-page auth requirements in frontmatter:
  ```yaml
  auth: required
  # or
  auth:
    provider: github
    allowed_orgs: [mycompany]
  ```
- [ ] User context available in templates (`{{.User.Email}}`)

**Impact:** Secure internal tools; multi-user apps

---

### 4.2 Request Rate Limiting
**Work Required:**
- [ ] Per-IP rate limiting
- [ ] Per-source rate limiting (protect external APIs)
- [ ] Configurable limits per endpoint
- [ ] Graceful 429 responses with retry-after

**Impact:** Protection against abuse; resource management

---

### 4.3 Health & Metrics Endpoints
**Work Required:**
- [ ] `/health` - Basic liveness check
- [ ] `/ready` - Readiness including source connectivity
- [ ] `/metrics` - Prometheus-compatible metrics:
  - Request count/latency by route
  - WebSocket connection count
  - Source fetch latency/error rates
  - WASM execution time

**Impact:** Kubernetes-ready deployment; observability

---

### 4.4 Graceful Shutdown
**Work Required:**
- [ ] Track in-flight requests
- [ ] Drain WebSocket connections
- [ ] Complete pending source operations
- [ ] Close WASM runtimes cleanly
- [ ] Configurable shutdown timeout

**Impact:** Zero-downtime deployments

---

### 4.5 Single Binary Distribution
**Work Required:**
- [ ] Embed client assets in Go binary
- [ ] `tinkerdown build <dir>` command producing standalone binary
- [ ] Cross-compilation support (linux, darwin, windows)
- [ ] Docker image generation

**Impact:** Simple deployment; Docker images <50MB

---

## Phase 6: UI & Components (P2)

### Value Proposition
> "Beautiful apps without CSS expertise"

### 5.1 Component Library
**Work Required:**
- [ ] Chart component (line, bar, pie via Chart.js or similar)
- [ ] Modal/dialog component with `lvt-modal`
- [ ] Toast notifications for action feedback
- [ ] File upload with drag-and-drop
- [ ] Tree view for hierarchical data
- [ ] Tabs component
- [ ] Accordion/collapsible sections

**Impact:** Rich UIs without custom code

---

### 5.2 Built-in Pagination
**Current State:** Must render all data; large datasets slow.

**Work Required:**
- [ ] `lvt-paginate="20"` attribute on containers
- [ ] Auto-generate prev/next controls
- [ ] Server-side pagination for sources
- [ ] URL-based page state for bookmarkability

**Impact:** Apps handling 10k+ records

---

### 5.3 Built-in Sorting & Filtering
**Work Required:**
- [ ] `lvt-sortable` attribute on tables
- [ ] `lvt-filter="field"` for search input
- [ ] Client-side for small datasets (<1000 rows)
- [ ] Server-side for large datasets

**Impact:** Usable data tables out of the box

---

### 5.4 Theme System Expansion
**Current State:** Only "clean" theme fully implemented.

**Work Required:**
- [ ] Complete "dark" and "minimal" themes
- [ ] Custom theme via CSS variables:
  ```yaml
  styling:
    theme: custom
    primary_color: "#007bff"
    background: "#1a1a2e"
  ```
- [ ] Dark mode toggle component
- [ ] Per-page theme override in frontmatter

**Impact:** Brand customization; accessibility

---

### 5.5 UX Improvements (from UX_IMPROVEMENTS.md)
**Work Required:**
- [ ] Sticky table of contents (right sidebar)
- [ ] Previous/Next navigation buttons
- [ ] Code copy buttons on all code blocks
- [ ] Sidebar collapse/expand toggle
- [ ] Active page indicator improvements
- [ ] Loading states/skeleton screens
- [ ] Better search result previews (120+ chars)

**Impact:** Professional documentation sites

---

## Phase 7: Advanced Features (P2)

### Value Proposition
> "Handle complex real-world scenarios"

### 6.1 Multi-User State Broadcasting
**Work Required:**
- [ ] Shared state mode for collaborative apps:
  ```yaml
  sources:
    tasks:
      type: sqlite
      broadcast: true  # Sync across all connected clients
  ```
- [ ] Broadcast state changes to all connected clients
- [ ] Conflict resolution strategies (last-write-wins, merge)
- [ ] Presence indicators (who's viewing)

**Impact:** Real-time collaborative tools

---

### 6.2 Scheduled Tasks
**Work Required:**
- [ ] Cron-like syntax in config:
  ```yaml
  schedules:
    refresh_data:
      cron: "*/5 * * * *"
      source: external_api
      action: Refresh
  ```
- [ ] Background execution without user connection
- [ ] Error notifications via webhook

**Impact:** Data refresh without user interaction

---

### 6.3 API Endpoint Mode
**Work Required:**
- [ ] Expose sources as REST endpoints:
  ```yaml
  api:
    enabled: true
    prefix: /api/v1
    sources:
      - name: tasks
        methods: [GET, POST, PUT, DELETE]
  ```
- [ ] OpenAPI spec generation
- [ ] API key authentication

**Impact:** Backend services from markdown definitions

---

### 6.4 Offline Support
**Work Required:**
- [ ] Service worker for asset caching
- [ ] Offline indicator component
- [ ] Queue actions when offline
- [ ] Sync when reconnected
- [ ] Conflict resolution UI

**Impact:** Reliable mobile/field usage

---

### 6.5 WASM Source Marketplace
**Work Required:**
- [ ] Registry of community WASM sources
- [ ] `tinkerdown source add <name>` command
- [ ] Versioning and updates
- [ ] Security review process
- [ ] Documentation site for source authors

**Impact:** Rich ecosystem without core development

---

## Phase 8: Polish & Optimization (P3)

### Value Proposition
> "Production-grade performance and reliability"

### 7.1 Bundle Size Optimization
**Current State:** Client bundle includes Monaco for code blocks.

**Work Required:**
- [ ] Lazy load heavy components (Monaco, charts)
- [ ] Tree-shake unused features
- [ ] Alternative lightweight code viewer
- [ ] Target <200KB initial bundle

---

### 7.2 Accessibility Audit
**Work Required:**
- [ ] WCAG 2.1 AA compliance
- [ ] Full keyboard navigation
- [ ] Screen reader testing
- [ ] ARIA attributes for all components
- [ ] Focus management on page transitions
- [ ] Color contrast validation

---

### 7.3 Performance Profiling
**Work Required:**
- [ ] Built-in performance tracing
- [ ] Slow source warnings (>500ms)
- [ ] Memory usage monitoring for WASM
- [ ] WebSocket message size optimization
- [ ] Template rendering performance

---

### 7.4 Comprehensive Test Suite
**Work Required:**
- [ ] Unit tests for all sources (including SQLite, WASM)
- [ ] Integration tests for GenericState action dispatch
- [ ] Cross-browser E2E tests
- [ ] Performance regression tests
- [ ] WASM source contract tests

---

## Implementation Priorities

```
Priority Order (based on user impact and dependencies):

HIGH IMPACT, LOW EFFORT (Quick Wins)
├── 2.3 Debug mode CLI flag
├── 2.2 Source reference validation
├── 2.1 CLI templates (todo, dashboard)
└── 5.5 Code copy buttons

HIGH IMPACT, MEDIUM EFFORT (Core Features)
├── 1.2 Source caching
├── 1.1 Error handling improvements
├── 2.5 WASM source dev kit
├── 4.3 Health endpoints
└── 3.1 GraphQL source

HIGH IMPACT, HIGH EFFORT (Major Features)
├── 4.1 Authentication
├── 6.1 Multi-user broadcasting
├── 4.5 Single binary distribution
└── 6.3 API endpoint mode

MEDIUM IMPACT (Nice to Have)
├── 5.1-5.4 Component library
├── 3.3 Source composition
├── 6.2 Scheduled tasks
└── 7.1-7.4 Polish items
```

---

## Success Metrics

### Phase 1-2 Complete
- [ ] All 8 example apps work reliably
- [ ] Zero crashes on source errors (graceful degradation)
- [ ] New app scaffolded in <1 minute
- [ ] 90% of config errors caught at validation time

### Phase 3-4 Complete
- [ ] Apps load <500ms with caching
- [ ] 8+ data source types available
- [ ] Apps deployable to Kubernetes with health checks
- [ ] Auth-protected internal tools working

### Phase 5-6 Complete
- [ ] 10+ reusable components
- [ ] 10k row tables render smoothly
- [ ] Real-time collaborative demo working
- [ ] REST API generation from sources

### Phase 7 Complete
- [ ] <200KB initial bundle size
- [ ] WCAG 2.1 AA compliant
- [ ] 95%+ test coverage on core

---

## Quick Wins (Start Immediately)

These high-impact, low-effort items can be tackled immediately:

**Zero-Template (Highest Priority):**
1. **Auto-table rendering** - `<table lvt-source="x" lvt-columns="a,b,c">` generates full table
2. **Document existing `<select lvt-source>`** - Already works, just undocumented
3. **`lvt-for` attribute** - Simple loop syntax without Go templates
4. **`lvt-text` attribute** - Set text content from field

**Developer Experience:**
5. **Add `--debug` flag** - Expose debug logging via CLI
6. **CLI templates** - Add todo and dashboard templates to `new` command
7. **Source reference validation** - Check sources exist in validate command

---

## Contributing

See `CONTRIBUTING.md` for development setup and guidelines.

Each feature should have:
1. Design document before implementation
2. Tests covering new functionality
3. Documentation updates
4. Example apps demonstrating usage
