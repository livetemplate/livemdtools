package runtime

import (
	"bytes"
	"strings"
	"text/template"
	"time"
)

// TemplateFuncs provides template functions for action data processing.
// These are used to fill in dynamic values like timestamps and dates.
var TemplateFuncs = template.FuncMap{
	// Time functions
	"now":       func() time.Time { return time.Now() },
	"today":     func() string { return time.Now().Format("2006-01-02") },
	"timestamp": func() string { return time.Now().Format(time.RFC3339) },
	"unix":      func() int64 { return time.Now().Unix() },

	// Date formatting
	"formatDate": formatDate,

	// Math functions (useful for date arithmetic in the future)
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
}

// formatDate formats a time.Time with the given layout
func formatDate(t time.Time, layout string) string {
	return t.Format(layout)
}

// DefaultResolver resolves template expressions in default values.
// It supports time functions and operator identity.
type DefaultResolver struct {
	funcs    template.FuncMap
	data     map[string]interface{}
	operator string
}

// NewDefaultResolver creates a new DefaultResolver with the given operator identity.
func NewDefaultResolver(operator string) *DefaultResolver {
	return &DefaultResolver{
		funcs:    TemplateFuncs,
		operator: operator,
		data: map[string]interface{}{
			"operator": operator,
		},
	}
}

// Resolve processes a string value, expanding any template expressions.
// If the value doesn't contain template syntax, it's returned unchanged.
func (r *DefaultResolver) Resolve(value string) (interface{}, error) {
	// Quick check for template syntax
	if !strings.Contains(value, "{{") {
		return value, nil
	}

	tmpl, err := template.New("default").Funcs(r.funcs).Parse(value)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r.data); err != nil {
		return nil, err
	}

	return buf.String(), nil
}

// ResolveMap processes all string values in a map, expanding template expressions.
// This is used to process action data before writing.
func (r *DefaultResolver) ResolveMap(data map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(data))
	for k, v := range data {
		switch val := v.(type) {
		case string:
			resolved, err := r.Resolve(val)
			if err != nil {
				return nil, err
			}
			result[k] = resolved
		default:
			result[k] = v
		}
	}
	return result, nil
}
