package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecSource(t *testing.T) {
	tests := []struct {
		name    string
		srcName string
		cmd     string
		wantErr bool
	}{
		{
			name:    "valid command",
			srcName: "test",
			cmd:     "echo hello",
			wantErr: false,
		},
		{
			name:    "empty command",
			srcName: "test",
			cmd:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, err := NewExecSource(tt.srcName, tt.cmd, ".")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.srcName, src.Name())
		})
	}
}

func TestExecSourceFetchJSON(t *testing.T) {
	// Create a temp directory with a script that outputs JSON
	tmpDir := t.TempDir()

	// Create a script that outputs JSON array
	scriptPath := filepath.Join(tmpDir, "data.sh")
	scriptContent := `#!/bin/bash
echo '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]'
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	src, err := NewExecSource("test", "./data.sh", tmpDir)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := src.Fetch(ctx)
	require.NoError(t, err)
	require.Len(t, data, 2)

	assert.Equal(t, float64(1), data[0]["id"])
	assert.Equal(t, "Alice", data[0]["name"])
	assert.Equal(t, float64(2), data[1]["id"])
	assert.Equal(t, "Bob", data[1]["name"])
}

func TestExecSourceFetchSingleObject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a script that outputs a single JSON object
	scriptPath := filepath.Join(tmpDir, "single.sh")
	scriptContent := `#!/bin/bash
echo '{"status":"ok","count":42}'
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	src, err := NewExecSource("test", "./single.sh", tmpDir)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := src.Fetch(ctx)
	require.NoError(t, err)
	require.Len(t, data, 1)

	assert.Equal(t, "ok", data[0]["status"])
	assert.Equal(t, float64(42), data[0]["count"])
}

func TestExecSourceFetchNDJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a script that outputs newline-delimited JSON
	scriptPath := filepath.Join(tmpDir, "ndjson.sh")
	scriptContent := `#!/bin/bash
echo '{"line":1}'
echo '{"line":2}'
echo '{"line":3}'
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	src, err := NewExecSource("test", "./ndjson.sh", tmpDir)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := src.Fetch(ctx)
	require.NoError(t, err)
	require.Len(t, data, 3)

	assert.Equal(t, float64(1), data[0]["line"])
	assert.Equal(t, float64(2), data[1]["line"])
	assert.Equal(t, float64(3), data[2]["line"])
}

func TestExecSourceFetchEmptyOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a script with empty output
	scriptPath := filepath.Join(tmpDir, "empty.sh")
	scriptContent := `#!/bin/bash
echo ""
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	src, err := NewExecSource("test", "./empty.sh", tmpDir)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := src.Fetch(ctx)
	require.NoError(t, err)
	assert.Empty(t, data)
}

func TestExecSourceFetchInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a script with invalid JSON
	scriptPath := filepath.Join(tmpDir, "invalid.sh")
	scriptContent := `#!/bin/bash
echo 'not valid json'
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	src, err := NewExecSource("test", "./invalid.sh", tmpDir)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = src.Fetch(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON output")
}

func TestExecSourceFetchCommandFails(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a script that exits with error
	scriptPath := filepath.Join(tmpDir, "fail.sh")
	scriptContent := `#!/bin/bash
exit 1
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	src, err := NewExecSource("test", "./fail.sh", tmpDir)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = src.Fetch(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command failed")
}

func TestExecSourceClose(t *testing.T) {
	src, err := NewExecSource("test", "echo hello", ".")
	require.NoError(t, err)

	err = src.Close()
	assert.NoError(t, err)
}
