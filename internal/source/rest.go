package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/livetemplate/tinkerdown/internal/config"
)

// RestSource fetches data from a REST API endpoint
type RestSource struct {
	name           string
	url            string
	method         string
	headers        map[string]string
	queryParams    map[string]string
	resultPath     string
	client         *http.Client
	retryConfig    RetryConfig
	circuitBreaker *CircuitBreaker
}

// NewRestSource creates a new REST API source (legacy, uses URL directly)
func NewRestSource(name, apiURL string, options map[string]string) (*RestSource, error) {
	cfg := config.SourceConfig{
		From:    apiURL,
		Options: options,
	}
	return NewRestSourceWithConfig(name, cfg)
}

// NewRestSourceWithConfig creates a new REST API source with full configuration
func NewRestSourceWithConfig(name string, cfg config.SourceConfig) (*RestSource, error) {
	apiURL := cfg.From
	if apiURL == "" {
		return nil, &ValidationError{Source: name, Field: "from", Reason: "from is required"}
	}

	// Expand environment variables in URL
	apiURL = os.ExpandEnv(apiURL)

	// Get method from options (default: GET)
	method := "GET"
	if cfg.Options != nil && cfg.Options["method"] != "" {
		method = strings.ToUpper(cfg.Options["method"])
	}

	// Build headers map with env var expansion
	headers := make(map[string]string)

	// Use headers from config (new YAML map format)
	for key, value := range cfg.Headers {
		headers[key] = os.ExpandEnv(value)
	}

	// Legacy: Parse headers from options (format: "key1:value1,key2:value2")
	if cfg.Options != nil && cfg.Options["headers"] != "" {
		for _, h := range strings.Split(cfg.Options["headers"], ",") {
			parts := strings.SplitN(h, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				headers[key] = os.ExpandEnv(value)
			}
		}
	}

	// Legacy: Check for common auth headers from options
	if cfg.Options != nil {
		if authHeader := cfg.Options["auth_header"]; authHeader != "" {
			headers["Authorization"] = os.ExpandEnv(authHeader)
		}
		if apiKey := cfg.Options["api_key"]; apiKey != "" {
			headers["X-API-Key"] = os.ExpandEnv(apiKey)
		}
	}

	// Build query params with env var expansion
	queryParams := make(map[string]string)
	for key, value := range cfg.QueryParams {
		queryParams[key] = os.ExpandEnv(value)
	}

	// Get timeout from config or default
	timeout := cfg.GetTimeout()

	// Build retry config
	retryConfig := RetryConfig{
		MaxRetries: cfg.GetRetryMaxRetries(),
		BaseDelay:  cfg.GetRetryBaseDelay(),
		MaxDelay:   cfg.GetRetryMaxDelay(),
		Multiplier: 2.0,
		EnableLog:  true,
	}

	// Create circuit breaker
	cbConfig := DefaultCircuitBreakerConfig()
	circuitBreaker := NewCircuitBreaker(name, cbConfig)

	return &RestSource{
		name:           name,
		url:            apiURL,
		method:         method,
		headers:        headers,
		queryParams:    queryParams,
		resultPath:     cfg.ResultPath,
		retryConfig:    retryConfig,
		circuitBreaker: circuitBreaker,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Name returns the source identifier
func (s *RestSource) Name() string {
	return s.name
}

// Fetch makes an HTTP request and parses JSON response with retry and circuit breaker
func (s *RestSource) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	// Use circuit breaker + retry
	return s.circuitBreaker.Execute(ctx, func(ctx context.Context) ([]map[string]interface{}, error) {
		return WithRetry(ctx, s.name, s.retryConfig, func(ctx context.Context) ([]map[string]interface{}, error) {
			return s.doFetch(ctx)
		})
	})
}

// buildURLWithQueryParams merges queryParams with any existing URL query parameters
func (s *RestSource) buildURLWithQueryParams() (string, error) {
	if len(s.queryParams) == 0 {
		return s.url, nil
	}

	parsedURL, err := url.Parse(s.url)
	if err != nil {
		return "", err
	}

	// Get existing query params from URL
	query := parsedURL.Query()

	// Merge in our queryParams (these take precedence)
	for key, value := range s.queryParams {
		query.Set(key, value)
	}

	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

// doFetch performs the actual HTTP request
func (s *RestSource) doFetch(ctx context.Context) ([]map[string]interface{}, error) {
	// Build URL with merged query parameters
	requestURL, err := s.buildURLWithQueryParams()
	if err != nil {
		return nil, &SourceError{
			Source:    s.name,
			Operation: "build URL",
			Err:       err,
			Retryable: false,
		}
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, s.method, requestURL, nil)
	if err != nil {
		return nil, &SourceError{
			Source:    s.name,
			Operation: "create request",
			Err:       err,
			Retryable: false,
		}
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	for key, value := range s.headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, NewSourceError(s.name, "request", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, &HTTPError{
			Source:     s.name,
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       strings.TrimSpace(string(body)),
		}
	}

	// Read response body with size limit to prevent OOM
	const maxResponseSize = 10 * 1024 * 1024 // 10MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, &SourceError{
			Source:    s.name,
			Operation: "read response",
			Err:       err,
			Retryable: false,
		}
	}

	// Parse JSON response
	return s.parseJSON(body)
}

// navigateJSONPath extracts nested data using dot notation
// e.g., "data.items" on {"data": {"items": [...]}} returns the array
func navigateJSONPath(data interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("path '%s' not found at key '%s'", path, part)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot navigate path '%s': expected object at '%s', got %T", path, part, current)
		}
	}

	return current, nil
}

// parseJSON handles both array and object JSON responses
// If resultPath is specified, navigates to that path to extract the array
// Otherwise expects array at root, or wraps single object in array
func (s *RestSource) parseJSON(data []byte) ([]map[string]interface{}, error) {
	data = []byte(strings.TrimSpace(string(data)))

	if len(data) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Parse JSON into generic structure first
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, &ValidationError{Source: s.name, Reason: "could not parse response as JSON"}
	}

	// Navigate to resultPath if specified
	if s.resultPath != "" {
		result, err := navigateJSONPath(parsed, s.resultPath)
		if err != nil {
			return nil, &ValidationError{Source: s.name, Reason: err.Error()}
		}
		parsed = result
	}

	// Convert to []map[string]interface{}
	return s.convertToMapSlice(parsed)
}

// convertToMapSlice converts an interface{} to []map[string]interface{}
func (s *RestSource) convertToMapSlice(data interface{}) ([]map[string]interface{}, error) {
	switch v := data.(type) {
	case []interface{}:
		results := make([]map[string]interface{}, 0, len(v))
		for i, item := range v {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				return nil, &ValidationError{
					Source: s.name,
					Reason: fmt.Sprintf("expected array of objects, but element %d is %T", i, item),
				}
			}
			results = append(results, itemMap)
		}
		return results, nil

	case []map[string]interface{}:
		return v, nil

	case map[string]interface{}:
		// Single object - wrap in array
		return []map[string]interface{}{v}, nil

	default:
		return nil, &ValidationError{Source: s.name, Reason: fmt.Sprintf("expected array or object, got %T", data)}
	}
}

// Close is a no-op for REST sources
func (s *RestSource) Close() error {
	return nil
}

// sensitiveHeaders is a list of header names that should be masked in output
var sensitiveHeaders = map[string]bool{
	"authorization": true,
	"x-api-key":     true,
	"x-auth-token":  true,
	"cookie":        true,
	"set-cookie":    true,
}

// String returns a debug representation with masked sensitive headers
func (s *RestSource) String() string {
	maskedHeaders := make(map[string]string, len(s.headers))
	for key, value := range s.headers {
		if sensitiveHeaders[strings.ToLower(key)] {
			runes := []rune(value)
			if len(runes) > 4 {
				maskedHeaders[key] = string(runes[:4]) + "****"
			} else {
				maskedHeaders[key] = "****"
			}
		} else {
			maskedHeaders[key] = value
		}
	}

	return fmt.Sprintf("RestSource{name: %q, url: %q, method: %q, headers: %v, queryParams: %v, resultPath: %q}",
		s.name, s.url, s.method, maskedHeaders, s.queryParams, s.resultPath)
}
