// Package source provides data source implementations for lvt-source blocks.
package source

import (
	"fmt"
	"strings"
)

// extractPath extracts an array from nested data using dot-notation path.
// Example: "repository.issues.nodes" extracts data["repository"]["issues"]["nodes"]
func extractPath(data map[string]interface{}, path string) ([]map[string]interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("result_path is required")
	}

	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("path '%s' not found at '%s'", path, part)
			}
		default:
			return nil, fmt.Errorf("path '%s' cannot traverse non-object at '%s'", path, part)
		}
	}

	// Convert to []map[string]interface{}
	arr, ok := current.([]interface{})
	if !ok {
		return nil, fmt.Errorf("path '%s' does not resolve to an array", path)
	}

	result := make([]map[string]interface{}, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}

	return result, nil
}
