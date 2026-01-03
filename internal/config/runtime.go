package config

import (
	"os"
	"sync"
)

// RuntimeConfig stores configuration set at runtime via CLI flags.
// These values are not persisted to config files.
type RuntimeConfig struct {
	mu       sync.RWMutex
	operator string
}

var globalRuntime = &RuntimeConfig{}

// SetOperator sets the operator identity for this session.
// If empty, defaults to the current user from $USER environment variable.
func SetOperator(op string) {
	globalRuntime.mu.Lock()
	defer globalRuntime.mu.Unlock()

	if op == "" {
		// Default to current user
		op = os.Getenv("USER")
	}
	globalRuntime.operator = op
}

// GetOperator returns the current operator identity.
// Returns empty string if not set and $USER is not available.
func GetOperator() string {
	globalRuntime.mu.RLock()
	defer globalRuntime.mu.RUnlock()
	return globalRuntime.operator
}
