package source

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/livetemplate/tinkerdown/internal/config"
)

func TestRestSource_BasicFetch(t *testing.T) {
	// Test server returns a JSON array at root
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
		})
	}))
	defer server.Close()

	cfg := config.SourceConfig{
		Type: "rest",
		From: server.URL,
	}
	src, err := NewRestSourceWithConfig("test", cfg)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	results, err := src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[0]["name"] != "Alice" {
		t.Errorf("Expected first name to be Alice, got %v", results[0]["name"])
	}
}

func TestRestSource_WithHeaders(t *testing.T) {
	var receivedAuth string
	var receivedCustom string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		receivedCustom = r.Header.Get("X-Custom-Header")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{{"ok": true}})
	}))
	defer server.Close()

	cfg := config.SourceConfig{
		Type: "rest",
		From: server.URL,
		Headers: map[string]string{
			"Authorization":   "Bearer test-token",
			"X-Custom-Header": "custom-value",
		},
	}
	src, err := NewRestSourceWithConfig("test", cfg)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	_, err = src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if receivedAuth != "Bearer test-token" {
		t.Errorf("Expected Authorization header 'Bearer test-token', got %q", receivedAuth)
	}
	if receivedCustom != "custom-value" {
		t.Errorf("Expected X-Custom-Header 'custom-value', got %q", receivedCustom)
	}
}

func TestRestSource_HeadersEnvExpansion(t *testing.T) {
	os.Setenv("TEST_API_TOKEN", "secret-env-token")
	defer os.Unsetenv("TEST_API_TOKEN")

	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{{"ok": true}})
	}))
	defer server.Close()

	cfg := config.SourceConfig{
		Type: "rest",
		From: server.URL,
		Headers: map[string]string{
			"Authorization": "Bearer ${TEST_API_TOKEN}",
		},
	}
	src, err := NewRestSourceWithConfig("test", cfg)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	_, err = src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if receivedAuth != "Bearer secret-env-token" {
		t.Errorf("Expected Authorization header with expanded env var, got %q", receivedAuth)
	}
}

func TestRestSource_WithQueryParams(t *testing.T) {
	var receivedLimit string
	var receivedStatus string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedLimit = r.URL.Query().Get("limit")
		receivedStatus = r.URL.Query().Get("status")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{{"ok": true}})
	}))
	defer server.Close()

	cfg := config.SourceConfig{
		Type: "rest",
		From: server.URL,
		QueryParams: map[string]string{
			"limit":  "100",
			"status": "active",
		},
	}
	src, err := NewRestSourceWithConfig("test", cfg)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	_, err = src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if receivedLimit != "100" {
		t.Errorf("Expected limit=100, got %q", receivedLimit)
	}
	if receivedStatus != "active" {
		t.Errorf("Expected status=active, got %q", receivedStatus)
	}
}

func TestRestSource_QueryParamsMerge(t *testing.T) {
	var receivedExisting string
	var receivedNew string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedExisting = r.URL.Query().Get("existing")
		receivedNew = r.URL.Query().Get("new_param")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{{"ok": true}})
	}))
	defer server.Close()

	// URL already has ?existing=value
	cfg := config.SourceConfig{
		Type: "rest",
		From: server.URL + "?existing=value",
		QueryParams: map[string]string{
			"new_param": "new_value",
		},
	}
	src, err := NewRestSourceWithConfig("test", cfg)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	_, err = src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if receivedExisting != "value" {
		t.Errorf("Expected existing param preserved, got %q", receivedExisting)
	}
	if receivedNew != "new_value" {
		t.Errorf("Expected new_param=new_value, got %q", receivedNew)
	}
}

func TestRestSource_ResultPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"items": []map[string]interface{}{
					{"id": 1, "name": "Alice"},
					{"id": 2, "name": "Bob"},
				},
			},
		})
	}))
	defer server.Close()

	cfg := config.SourceConfig{
		Type:       "rest",
		From:       server.URL,
		ResultPath: "data.items",
	}
	src, err := NewRestSourceWithConfig("test", cfg)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	results, err := src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results from data.items, got %d", len(results))
	}
	if results[0]["name"] != "Alice" {
		t.Errorf("Expected first name to be Alice, got %v", results[0]["name"])
	}
}

func TestRestSource_ResultPathNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"users": []interface{}{}, // Note: "users" not "items"
			},
		})
	}))
	defer server.Close()

	cfg := config.SourceConfig{
		Type:       "rest",
		From:       server.URL,
		ResultPath: "data.items", // Path doesn't exist
	}
	src, err := NewRestSourceWithConfig("test", cfg)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	_, err = src.Fetch(context.Background())
	if err == nil {
		t.Fatal("Expected error for missing result_path, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

func TestRestSource_NoResultPath_RootArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1},
			{"id": 2},
		})
	}))
	defer server.Close()

	cfg := config.SourceConfig{
		Type: "rest",
		From: server.URL,
		// No ResultPath - should work with array at root
	}
	src, err := NewRestSourceWithConfig("test", cfg)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	results, err := src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results from root array, got %d", len(results))
	}
}

func TestRestSource_NoResultPath_SingleObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   1,
			"name": "Single Item",
		})
	}))
	defer server.Close()

	cfg := config.SourceConfig{
		Type: "rest",
		From: server.URL,
		// No ResultPath - should wrap single object in array
	}
	src, err := NewRestSourceWithConfig("test", cfg)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	results, err := src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result (wrapped single object), got %d", len(results))
	}
	if results[0]["name"] != "Single Item" {
		t.Errorf("Expected name to be 'Single Item', got %v", results[0]["name"])
	}
}

func TestRestSource_String_MasksAuth(t *testing.T) {
	cfg := config.SourceConfig{
		Type: "rest",
		From: "https://api.example.com/data",
		Headers: map[string]string{
			"Authorization": "Bearer super-secret-token-12345",
			"Accept":        "application/json",
			"X-API-Key":     "key-12345",
		},
		QueryParams: map[string]string{
			"limit": "100",
		},
		ResultPath: "data.items",
	}
	src, err := NewRestSourceWithConfig("test", cfg)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	str := src.String()

	// Authorization should be masked
	if strings.Contains(str, "super-secret-token-12345") {
		t.Errorf("Authorization token should be masked in String(), got: %s", str)
	}
	if !strings.Contains(str, "Bear****") {
		t.Errorf("Expected masked Authorization like 'Bear****' in String(), got: %s", str)
	}

	// X-API-Key should also be masked
	if strings.Contains(str, "key-12345") {
		t.Errorf("X-API-Key should be masked in String(), got: %s", str)
	}

	// Accept header should NOT be masked
	if !strings.Contains(str, "application/json") {
		t.Errorf("Accept header should not be masked in String(), got: %s", str)
	}

	// Other fields should be present
	if !strings.Contains(str, "https://api.example.com/data") {
		t.Errorf("URL should be present in String(), got: %s", str)
	}
	if !strings.Contains(str, "data.items") {
		t.Errorf("resultPath should be present in String(), got: %s", str)
	}
}

func TestRestSource_MissingFrom(t *testing.T) {
	cfg := config.SourceConfig{
		Type: "rest",
		// From is missing
	}
	_, err := NewRestSourceWithConfig("test", cfg)
	if err == nil {
		t.Fatal("Expected error for missing 'from', got nil")
	}

	if !strings.Contains(err.Error(), "from is required") {
		t.Errorf("Expected 'from is required' in error, got: %v", err)
	}
}

func TestNavigateJSONPath(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		path    string
		want    interface{}
		wantErr bool
	}{
		{
			name:    "empty path returns data",
			data:    map[string]interface{}{"a": 1},
			path:    "",
			want:    map[string]interface{}{"a": 1},
			wantErr: false,
		},
		{
			name: "single level path",
			data: map[string]interface{}{
				"items": []interface{}{1, 2, 3},
			},
			path:    "items",
			want:    []interface{}{1, 2, 3},
			wantErr: false,
		},
		{
			name: "nested path",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"items": []interface{}{"a", "b"},
				},
			},
			path:    "data.items",
			want:    []interface{}{"a", "b"},
			wantErr: false,
		},
		{
			name: "deeply nested path",
			data: map[string]interface{}{
				"response": map[string]interface{}{
					"data": map[string]interface{}{
						"users": []interface{}{"alice", "bob"},
					},
				},
			},
			path:    "response.data.users",
			want:    []interface{}{"alice", "bob"},
			wantErr: false,
		},
		{
			name: "path not found",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"users": []interface{}{},
				},
			},
			path:    "data.items",
			wantErr: true,
		},
		{
			name:    "path on non-object",
			data:    []interface{}{1, 2, 3},
			path:    "data",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := navigateJSONPath(tt.data, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("navigateJSONPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Compare as JSON for deep equality
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(tt.want)
				if string(gotJSON) != string(wantJSON) {
					t.Errorf("navigateJSONPath() = %s, want %s", gotJSON, wantJSON)
				}
			}
		})
	}
}
