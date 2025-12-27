# Tinkerdown Roadmap: Micro Apps & Tooling Platform

## Vision

Transform Tinkerdown from an interactive documentation tool into a **full-featured platform for building micro apps and internal tooling** using markdown as the primary interface.

**Target Users:**
- Developers building internal tools, dashboards, and admin panels
- Technical writers creating interactive documentation
- Teams needing quick data-driven apps without full-stack complexity
- AI systems generating functional apps from natural language

---

## Strategic Priorities

### Priority Framework

| Level | Criteria | Timeline Focus |
|-------|----------|----------------|
| **P0** | Blocks core functionality or causes data loss | Immediate |
| **P1** | High-impact features enabling new use cases | Near-term |
| **P2** | Developer experience and productivity | Medium-term |
| **P3** | Polish, optimization, and edge cases | Ongoing |

---

## Phase 1: Foundation & Core Fixes (P0)

### Value Proposition
> "Make what exists actually work reliably"

Without these fixes, users hit dead ends. These are blockers for adoption.

### 1.1 Fix Interactive Block State Routing
**Files:** `state.go:133-169`, `internal/server/websocket.go`

**Current State:** `routeInteractiveBlock()` returns empty tree - interactive blocks don't actually work.

**Work Required:**
- [ ] Implement LiveTemplate state factory integration
- [ ] Wire action dispatch to compiled Go plugin methods
- [ ] Return proper tree updates via WebSocket

**Impact:** Unlocks all interactive micro apps (counters, todos, forms)

---

### 1.2 Multi-Page WebSocket Support
**Files:** `internal/server/server.go:225`

**Current State:** WebSocket handler only serves first route - multi-page sites broken for interactivity.

**Work Required:**
- [ ] Accept page identifier via query param or path
- [ ] Route WebSocket messages to correct page's state
- [ ] Handle page transitions gracefully

**Impact:** Enables documentation sites with interactive examples on every page

---

### 1.3 WASM Compilation Endpoint
**Files:** `client/src/wasm/tinygo-executor.ts:4-10`

**Current State:** Client expects `/api/compile` but server doesn't implement it.

**Work Required:**
- [ ] Implement `/api/compile` endpoint accepting Go code
- [ ] Sandbox compilation with TinyGo
- [ ] Return compiled WASM binary
- [ ] Handle compilation errors gracefully

**Impact:** Enables code playground use case (tutorials, learning platforms)

---

### 1.4 Data Source Error Handling
**Files:** `internal/source/*.go`

**Current State:** Errors can crash or hang; no retry logic; silent failures possible.

**Work Required:**
- [ ] Unified error types for all sources
- [ ] Retry with exponential backoff for transient failures
- [ ] Circuit breaker for repeatedly failing sources
- [ ] User-friendly error messages in templates

**Impact:** Production reliability for data-driven apps

---

## Phase 2: Developer Experience (P1)

### Value Proposition
> "Reduce time from idea to working app by 10x"

These features dramatically improve the development loop and reduce frustration.

### 2.1 Enhanced CLI Scaffolding
**Files:** `cmd/tinkerdown/commands/new.go`

**Current State:** `new` command creates minimal template only.

**Work Required:**
- [ ] Add `--template` flag with options:
  - `counter` - Basic state management example
  - `todo` - CRUD with persistence
  - `dashboard` - Multi-source data display
  - `form` - Auto-persist contact form
  - `api-explorer` - REST source with refresh
  - `cli-wrapper` - Exec source with arg form
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
- [ ] Verify action names match Go method naming conventions
- [ ] Warn on unused block IDs
- [ ] Type-check template expressions against state struct
- [ ] Validate config against JSON schema

**Impact:** Catch errors at write-time, not runtime

---

### 2.3 Debug Mode & Logging
**Files:** `internal/server/server.go:238`

**Current State:** Debug mode hardcoded; no structured logging.

**Work Required:**
- [ ] `--debug` / `--verbose` CLI flags
- [ ] Structured JSON logging option
- [ ] Request correlation IDs
- [ ] WebSocket message logging (with sensitive data redaction)
- [ ] State change logging
- [ ] Performance timing for data sources

**Impact:** 10x faster debugging of production issues

---

### 2.4 Hot Reload for Configuration
**Current State:** Config changes require server restart.

**Work Required:**
- [ ] Watch `tinkerdown.yaml` for changes
- [ ] Reload sources without dropping WebSocket connections
- [ ] Clear compiled plugins on config change
- [ ] Notify connected clients of config reload

**Impact:** Faster iteration on source configuration

---

### 2.5 Live REPL for Actions
**New Feature**

**Work Required:**
- [ ] `tinkerdown repl <file.md>` command
- [ ] Interactive prompt to call actions
- [ ] Display state after each action
- [ ] Show generated HTML output

**Impact:** Test state logic without browser

---

## Phase 3: Data Ecosystem (P1)

### Value Proposition
> "Connect to any data source in minutes, not days"

Data sources are the backbone of micro apps. More sources = more use cases.

### 3.1 SQLite Source
**Priority:** High (embedded database for simple apps)

**Work Required:**
- [ ] New source type: `sqlite`
- [ ] Config: `file: ./data.db`, `query: SELECT ...`
- [ ] Auto-create database file if missing
- [ ] Parameter binding for safe queries
- [ ] Write support with `INSERT`/`UPDATE`/`DELETE`

**Impact:** Full CRUD apps without external database

---

### 3.2 Source Caching Layer
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
- [ ] Manual cache clear via action

**Impact:** 10x faster page loads; reduced API costs

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
  ```
- [ ] Join sources on common fields
- [ ] Aggregation (count, sum, avg)

**Impact:** Complex data apps without custom Go code

---

### 3.4 GraphQL Source
**Work Required:**
- [ ] New source type: `graphql`
- [ ] Config: `url`, `query`, `variables`
- [ ] Authentication headers
- [ ] Auto-flatten nested response

**Impact:** Modern API ecosystem support

---

### 3.5 Webhook/Async Sources
**Work Required:**
- [ ] Source that receives HTTP POST
- [ ] Store latest N events
- [ ] Trigger UI update on new data
- [ ] Optional signature verification

**Impact:** Real-time integrations (Stripe webhooks, GitHub events)

---

## Phase 4: Production Readiness (P1)

### Value Proposition
> "Deploy with confidence to real users"

These features are required before any serious production use.

### 4.1 Authentication Middleware
**Work Required:**
- [ ] Built-in auth strategies:
  - API key (header-based)
  - Basic auth
  - OAuth2 (Google, GitHub)
  - Custom JWT validation
- [ ] Per-page auth requirements in frontmatter
- [ ] Public vs. authenticated sources
- [ ] User context available in templates (`{{.User.Email}}`)

**Impact:** Secure internal tools; multi-user apps

---

### 4.2 Request Rate Limiting
**Work Required:**
- [ ] Per-IP rate limiting
- [ ] Per-user rate limiting (with auth)
- [ ] Configurable limits per endpoint
- [ ] Graceful 429 responses

**Impact:** Protection against abuse; resource management

---

### 4.3 Health & Metrics Endpoints
**Work Required:**
- [ ] `/health` - Basic liveness check
- [ ] `/ready` - Readiness including data sources
- [ ] `/metrics` - Prometheus-compatible metrics:
  - Request count/latency by route
  - WebSocket connection count
  - Data source latency
  - Error rates

**Impact:** Kubernetes-ready deployment; observability

---

### 4.4 Graceful Shutdown
**Work Required:**
- [ ] Track in-flight requests
- [ ] Drain WebSocket connections
- [ ] Complete pending data source operations
- [ ] Configurable shutdown timeout

**Impact:** Zero-downtime deployments

---

### 4.5 Single Binary Distribution
**Work Required:**
- [ ] Embed client assets in Go binary
- [ ] `tinkerdown build` command producing standalone binary
- [ ] Include all dependencies (no Go toolchain required at runtime)
- [ ] Cross-compilation support

**Impact:** Simple deployment; Docker images <50MB

---

## Phase 5: UI & Components (P2)

### Value Proposition
> "Beautiful apps without CSS expertise"

Pre-built components accelerate development and ensure consistency.

### 5.1 Component Library
**Work Required:**
- [ ] Chart component (line, bar, pie via Chart.js)
- [ ] Modal/dialog component
- [ ] Toast notifications
- [ ] File upload with drag-and-drop
- [ ] Tree view for hierarchical data
- [ ] Tabs component
- [ ] Accordion/collapsible
- [ ] Card grid layout

**Impact:** Rich UIs without custom code

---

### 5.2 Built-in Pagination
**Current State:** Must render all data; large datasets slow.

**Work Required:**
- [ ] `lvt-paginate="20"` attribute
- [ ] Auto-generate prev/next controls
- [ ] Server-side pagination for sources
- [ ] URL-based page state

**Impact:** Apps handling 10k+ records

---

### 5.3 Built-in Sorting & Filtering
**Work Required:**
- [ ] `lvt-sortable` attribute on tables
- [ ] `lvt-filter="field"` for search input
- [ ] Client-side for small datasets
- [ ] Server-side for large datasets

**Impact:** Usable data tables out of the box

---

### 5.4 Theme System Expansion
**Current State:** Only "clean" theme fully implemented.

**Work Required:**
- [ ] Complete "dark" and "minimal" themes
- [ ] Custom theme via CSS variables config
- [ ] Dark mode toggle component
- [ ] Per-page theme override

**Impact:** Brand customization; accessibility

---

### 5.5 UX Improvements (from UX_IMPROVEMENTS.md)
**Work Required:**
- [ ] Sticky table of contents (right sidebar)
- [ ] Previous/Next navigation buttons
- [ ] Code copy buttons
- [ ] Sidebar collapse/expand
- [ ] Active page indicator improvements
- [ ] Loading states/skeleton screens

**Impact:** Professional documentation sites

---

## Phase 6: Advanced Features (P2)

### Value Proposition
> "Handle complex real-world scenarios"

These features address edge cases and advanced use cases.

### 6.1 Multi-User State Broadcasting
**Work Required:**
- [ ] Shared state mode for collaborative apps
- [ ] Broadcast state changes to all connected clients
- [ ] Conflict resolution strategies
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
      action: RefreshAll
  ```
- [ ] Background execution
- [ ] Error notifications

**Impact:** Data refresh without user interaction

---

### 6.3 API Endpoint Mode
**Work Required:**
- [ ] Expose actions as REST endpoints
- [ ] `tinkerdown.yaml` API configuration:
  ```yaml
  api:
    enabled: true
    prefix: /api/v1
    actions:
      - name: CreateUser
        method: POST
        path: /users
  ```
- [ ] OpenAPI spec generation

**Impact:** Backend services from markdown definitions

---

### 6.4 Offline Support
**Work Required:**
- [ ] Service worker for asset caching
- [ ] Offline indicator component
- [ ] Queue actions when offline
- [ ] Sync when reconnected

**Impact:** Reliable mobile/field usage

---

### 6.5 Plugin System for Custom Sources
**Work Required:**
- [ ] Source plugin interface
- [ ] Plugin discovery from `~/.tinkerdown/plugins`
- [ ] Example plugin template
- [ ] Plugin marketplace/registry

**Impact:** Community-extensible data ecosystem

---

## Phase 7: Polish & Optimization (P3)

### Value Proposition
> "Production-grade performance and reliability"

### 7.1 Bundle Size Optimization
**Current State:** Monaco editor adds ~3.6MB.

**Work Required:**
- [ ] Lazy load Monaco only for WASM blocks
- [ ] Tree-shake unused Monaco features
- [ ] Alternative lightweight editor option

---

### 7.2 Accessibility Audit
**Work Required:**
- [ ] WCAG 2.1 AA compliance
- [ ] Full keyboard navigation
- [ ] Screen reader testing
- [ ] ARIA attributes for all components
- [ ] Focus management

---

### 7.3 Performance Profiling
**Work Required:**
- [ ] Built-in performance tracing
- [ ] Slow source warnings
- [ ] Memory usage monitoring
- [ ] WebSocket message size optimization

---

### 7.4 Comprehensive Test Suite
**Work Required:**
- [ ] Unit tests for all sources
- [ ] Integration tests for WebSocket flows
- [ ] Cross-browser E2E tests
- [ ] Performance regression tests

---

## Implementation Timeline

```
┌─────────────────────────────────────────────────────────────────────┐
│ Phase 1: Foundation (P0)                                            │
│ ═══════════════════════                                             │
│ • Interactive block routing                                         │
│ • Multi-page WebSocket                                              │
│ • WASM compilation                                                  │
│ • Error handling                                                    │
├─────────────────────────────────────────────────────────────────────┤
│ Phase 2: DX (P1) ──────────────────────────────────────┐            │
│ • CLI templates                                        │            │
│ • Validation                                           │            │
│ • Debug mode                                           │            │
│ • Hot reload                                           │            │
├────────────────────────────────────────────────────────┼────────────┤
│ Phase 3: Data (P1) ────────────────────────────────────┤            │
│ • SQLite source                                        │            │
│ • Caching                                              │            │
│ • Composition                                          │            │
├────────────────────────────────────────────────────────┼────────────┤
│ Phase 4: Production (P1) ──────────────────────────────┘            │
│ • Auth                                                              │
│ • Rate limiting                                                     │
│ • Health checks                                                     │
│ • Binary distribution                                               │
├─────────────────────────────────────────────────────────────────────┤
│ Phase 5: UI (P2) ───────────────────────────────────────────────┐   │
│ • Components                                                    │   │
│ • Pagination                                                    │   │
│ • Themes                                                        │   │
├─────────────────────────────────────────────────────────────────┼───┤
│ Phase 6: Advanced (P2) ─────────────────────────────────────────┤   │
│ • Broadcasting                                                  │   │
│ • Scheduled tasks                                               │   │
│ • API mode                                                      │   │
├─────────────────────────────────────────────────────────────────┼───┤
│ Phase 7: Polish (P3) ───────────────────────────────────────────┘   │
│ • Bundle size                                                       │
│ • Accessibility                                                     │
│ • Performance                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Success Metrics

### Phase 1 Complete
- [ ] All 10 example apps work end-to-end
- [ ] Zero crashes on source errors
- [ ] Multi-page interactive sites functional

### Phase 2 Complete
- [ ] New app scaffolded in <1 minute
- [ ] 90% of errors caught at validation time
- [ ] Debug logs enable issue resolution without code changes

### Phase 3 Complete
- [ ] Apps load <500ms with caching
- [ ] 5+ data source types available
- [ ] Complex queries without custom Go code

### Phase 4 Complete
- [ ] Apps deployable to Kubernetes
- [ ] Auth-protected internal tools
- [ ] Single binary <50MB

### Phase 5 Complete
- [ ] 10+ reusable components
- [ ] 10k row tables render smoothly
- [ ] Professional-looking default styling

### Phase 6 Complete
- [ ] Real-time collaborative editing demo
- [ ] Scheduled data refresh without user action
- [ ] REST API from markdown definition

---

## Quick Wins (Can Start Immediately)

These high-impact, low-effort items can be tackled immediately:

1. **Fix config file naming** - Rename `livemdtools.yaml.example` to `tinkerdown.yaml.example`
2. **Add code copy buttons** - Simple client-side enhancement
3. **Add `--debug` flag** - Expose existing debug mode via CLI
4. **Expand validation** - Add source reference checking
5. **SQLite source** - Follow existing source patterns
6. **Document existing features** - Many features undocumented

---

## Contributing

See `CONTRIBUTING.md` for development setup and guidelines.

Each phase should have:
1. Design document before implementation
2. Tests covering new functionality
3. Documentation updates
4. Example apps demonstrating features
