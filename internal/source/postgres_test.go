package source

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPostgresSourceValidation(t *testing.T) {
	tests := []struct {
		name      string
		srcName   string
		query     string
		options   map[string]string
		wantErr   bool
		errReason string
	}{
		{
			name:      "empty query",
			srcName:   "test",
			query:     "",
			options:   nil,
			wantErr:   true,
			errReason: "query is required",
		},
		{
			name:      "missing dsn",
			srcName:   "test",
			query:     "SELECT * FROM users",
			options:   nil,
			wantErr:   true,
			errReason: "database connection required",
		},
		{
			name:      "missing dsn with empty options",
			srcName:   "test",
			query:     "SELECT * FROM users",
			options:   map[string]string{},
			wantErr:   true,
			errReason: "database connection required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, err := NewPostgresSource(tt.srcName, tt.query, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errReason)
				assert.Nil(t, src)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, src)
		})
	}
}

func TestPostgresSourceName(t *testing.T) {
	// Note: We can't fully test Name() without a valid DB connection
	// This test is here for coverage documentation
	t.Skip("requires valid database connection - tested via integration tests")
}
