// internal/source/graphql_test.go
package source

import (
	"testing"
)

func TestExtractPath_Simple(t *testing.T) {
	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice"},
			map[string]interface{}{"name": "Bob"},
		},
	}

	result, err := extractPath(data, "users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if result[0]["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", result[0]["name"])
	}
}

func TestExtractPath_Nested(t *testing.T) {
	data := map[string]interface{}{
		"repository": map[string]interface{}{
			"issues": map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{"title": "Bug"},
					map[string]interface{}{"title": "Feature"},
				},
			},
		},
	}

	result, err := extractPath(data, "repository.issues.nodes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if result[0]["title"] != "Bug" {
		t.Errorf("expected Bug, got %v", result[0]["title"])
	}
}

func TestExtractPath_NotFound(t *testing.T) {
	data := map[string]interface{}{
		"users": []interface{}{},
	}

	_, err := extractPath(data, "nonexistent.path")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestExtractPath_NotArray(t *testing.T) {
	data := map[string]interface{}{
		"user": map[string]interface{}{"name": "Alice"},
	}

	_, err := extractPath(data, "user")
	if err == nil {
		t.Error("expected error when path doesn't resolve to array")
	}
}

func TestExtractPath_EmptyPath(t *testing.T) {
	data := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"id": 1},
		},
	}

	_, err := extractPath(data, "")
	if err == nil {
		t.Error("expected error for empty path")
	}
}
