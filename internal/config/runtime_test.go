package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetGetOperator(t *testing.T) {
	// Reset state
	globalRuntime.operator = ""

	// Test explicit operator
	SetOperator("alice")
	assert.Equal(t, "alice", GetOperator())

	// Test changing operator
	SetOperator("bob")
	assert.Equal(t, "bob", GetOperator())
}

func TestSetOperatorDefaultsToUser(t *testing.T) {
	// Reset state
	globalRuntime.operator = ""

	// Set a known USER env var
	oldUser := os.Getenv("USER")
	os.Setenv("USER", "testuser")
	defer os.Setenv("USER", oldUser)

	// Empty operator should default to $USER
	SetOperator("")
	assert.Equal(t, "testuser", GetOperator())
}

func TestGetOperatorEmpty(t *testing.T) {
	// Reset state completely
	globalRuntime.operator = ""

	// Without calling SetOperator, GetOperator returns empty
	assert.Equal(t, "", GetOperator())
}
