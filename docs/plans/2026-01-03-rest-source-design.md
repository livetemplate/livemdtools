# REST Source Enhancement Design (Issue #37)

## Overview

Enhance REST API source with cleaner YAML config, query parameters, and configurable JSON path extraction.

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| URL field naming | Replace `url:` with `from:` | Consistency, matches issue spec |
| Headers config | Top-level `headers:` map | Cleaner than comma-separated in options |
| Result path default | Expect array at root | Simpler mental model, no auto-detect magic |
| Query params merge | Merge with URL params | Most flexible, matches HTTP library behavior |
| Pass-through queries | Skip for now | Complex, static query_params covers most cases |
| Auth masking | String() method only | Sufficient for debugging, don't over-engineer |

## Config Schema

```yaml
sources:
  api:
    type: rest
    from: https://api.example.com/users  # Replaces 'url:'
    headers:                              # New: proper YAML map
      Authorization: Bearer ${API_TOKEN}
      Accept: application/json
    query_params:                         # New: URL query parameters
      limit: "100"
      status: active
    result_path: data.items              # New: dot-path for nested JSON
    timeout: 30s
    retry:
      max_retries: 3
```

## Implementation

### Config Changes

**`internal/config/config.go`:**
```go
type SourceConfig struct {
    // ... existing fields ...
    From        string            `yaml:"from,omitempty"`         // For rest: API endpoint
    Headers     map[string]string `yaml:"headers,omitempty"`      // For rest: HTTP headers
    QueryParams map[string]string `yaml:"query_params,omitempty"` // For rest: URL query parameters
    ResultPath  string            `yaml:"result_path,omitempty"`  // For rest/graphql: dot-path to extract array
}
```

**`parser.go`:**
- Add same fields to parser's SourceConfig
- Remove `URL` field, replace with `From`

### REST Source Changes

**`internal/source/rest.go`:**

```go
type RestSource struct {
    name           string
    url            string            // Resolved from config.From
    method         string
    headers        map[string]string // From config.Headers
    queryParams    map[string]string // From config.QueryParams
    resultPath     string            // Dot-path like "data.items"
    client         *http.Client
    retryConfig    RetryConfig
    circuitBreaker *CircuitBreaker
}

// navigateJSONPath extracts nested data using dot notation
// "data.items" on {"data": {"items": [...]}} returns the array
func navigateJSONPath(data interface{}, path string) (interface{}, error)

// String returns a debug representation with masked sensitive headers
func (s *RestSource) String() string
```

**Behavior:**
- Expand env vars in `from`, all headers, all query params
- Merge `query_params` with any params already in URL
- Parse JSON response:
  - If `result_path` specified: navigate to that path
  - Otherwise: expect array at root, wrap single object in array

### Files to Modify

| File | Changes |
|------|---------|
| `internal/config/config.go` | Add `From`, `Headers`, `QueryParams` fields |
| `parser.go` | Same fields, remove `URL` |
| `internal/source/rest.go` | New constructor logic, query params, result_path, String() |
| `internal/source/rest_test.go` | New tests + update existing |
| `page.go` | Update source creation to use `From` |
| Examples using REST | Update `url:` to `from:` |

## Testing

```go
func TestRestSource_BasicFetch(t *testing.T)           // JSON array at root
func TestRestSource_WithHeaders(t *testing.T)          // Headers sent, env vars expanded
func TestRestSource_WithQueryParams(t *testing.T)      // Params appended to URL
func TestRestSource_QueryParamsMerge(t *testing.T)     // Merge URL params + query_params
func TestRestSource_ResultPath(t *testing.T)           // Navigate nested JSON
func TestRestSource_ResultPathNotFound(t *testing.T)   // Clear error for missing path
func TestRestSource_NoResultPath_RootArray(t *testing.T) // No path, expect root array
func TestRestSource_String_MasksAuth(t *testing.T)     // Authorization masked in output
```

## Definition of Done

- [ ] `from:` field works for REST sources
- [ ] `headers:` as YAML map with env var expansion
- [ ] `query_params:` merged with URL params
- [ ] `result_path:` navigates nested JSON
- [ ] No `result_path` → expects array at root
- [ ] `String()` masks sensitive headers
- [ ] All existing tests pass
- [ ] New unit tests for each feature
