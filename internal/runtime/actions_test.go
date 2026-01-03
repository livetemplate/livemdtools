package runtime

import (
	"testing"

	"github.com/livetemplate/tinkerdown/internal/config"
	"github.com/livetemplate/tinkerdown/internal/source"
)

func TestSubstituteParams(t *testing.T) {
	tests := []struct {
		name      string
		stmt      string
		data      map[string]interface{}
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name: "single param",
			stmt: "DELETE FROM tasks WHERE id = :id",
			data: map[string]interface{}{"id": 123},
			wantQuery: "DELETE FROM tasks WHERE id = ?",
			wantArgs:  []interface{}{123},
		},
		{
			name: "multiple params",
			stmt: "UPDATE tasks SET done = :done WHERE id = :id",
			data: map[string]interface{}{"id": 456, "done": true},
			wantQuery: "UPDATE tasks SET done = ? WHERE id = ?",
			wantArgs:  []interface{}{true, 456},
		},
		{
			name: "no params",
			stmt: "DELETE FROM tasks WHERE done = 1",
			data: map[string]interface{}{},
			wantQuery: "DELETE FROM tasks WHERE done = 1",
			wantArgs:  []interface{}{},
		},
		{
			name: "param with underscore",
			stmt: "SELECT * FROM users WHERE created_at < :cutoff_date",
			data: map[string]interface{}{"cutoff_date": "2024-01-01"},
			wantQuery: "SELECT * FROM users WHERE created_at < ?",
			wantArgs:  []interface{}{"2024-01-01"},
		},
		{
			name: "missing param (nil value)",
			stmt: "DELETE FROM tasks WHERE id = :id",
			data: map[string]interface{}{},
			wantQuery: "DELETE FROM tasks WHERE id = ?",
			wantArgs:  []interface{}{nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs := substituteParams(tt.stmt, tt.data)
			if gotQuery != tt.wantQuery {
				t.Errorf("substituteParams() query = %q, want %q", gotQuery, tt.wantQuery)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("substituteParams() args len = %d, want %d", len(gotArgs), len(tt.wantArgs))
				return
			}
			for i := range gotArgs {
				if gotArgs[i] != tt.wantArgs[i] {
					t.Errorf("substituteParams() args[%d] = %v, want %v", i, gotArgs[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestValidateParams(t *testing.T) {
	tests := []struct {
		name    string
		action  *config.Action
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name: "all required params present",
			action: &config.Action{
				Params: map[string]config.ParamDef{
					"id":   {Required: true},
					"name": {Required: true},
				},
			},
			data:    map[string]interface{}{"id": 1, "name": "test"},
			wantErr: false,
		},
		{
			name: "missing required param",
			action: &config.Action{
				Params: map[string]config.ParamDef{
					"id": {Required: true},
				},
			},
			data:    map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "empty required param",
			action: &config.Action{
				Params: map[string]config.ParamDef{
					"name": {Required: true},
				},
			},
			data:    map[string]interface{}{"name": ""},
			wantErr: true,
		},
		{
			name: "optional param missing",
			action: &config.Action{
				Params: map[string]config.ParamDef{
					"optional": {Required: false},
				},
			},
			data:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "no params defined",
			action: &config.Action{
				Params: nil,
			},
			data:    map[string]interface{}{"anything": "value"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParams(tt.action, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpandTemplate(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		data    map[string]interface{}
		want    string
		wantErr bool
	}{
		{
			name: "no template",
			text: "https://api.example.com/users",
			data: map[string]interface{}{},
			want: "https://api.example.com/users",
		},
		{
			name: "simple substitution",
			text: "https://api.example.com/users/{{.id}}",
			data: map[string]interface{}{"id": 123},
			want: "https://api.example.com/users/123",
		},
		{
			name: "json body",
			text: `{"text": "Task: {{.task}}"}`,
			data: map[string]interface{}{"task": "Buy groceries"},
			want: `{"text": "Task: Buy groceries"}`,
		},
		{
			name: "invalid template",
			text: "{{.broken",
			data: map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandTemplate(tt.text, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("expandTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsExecAllowed(t *testing.T) {
	// Save original state and restore after test
	origState := config.IsExecAllowed()
	defer config.SetAllowExec(origState)

	// Test default (should be disabled)
	config.SetAllowExec(false)
	if config.IsExecAllowed() {
		t.Error("IsExecAllowed() should be false by default")
	}

	// Test enabled
	config.SetAllowExec(true)
	if !config.IsExecAllowed() {
		t.Error("IsExecAllowed() should be true after SetAllowExec(true)")
	}

	// Test disabled again
	config.SetAllowExec(false)
	if config.IsExecAllowed() {
		t.Error("IsExecAllowed() should be false after SetAllowExec(false)")
	}
}

func TestExecuteCustomAction_UnknownKind(t *testing.T) {
	state := &GenericState{}
	action := &config.Action{Kind: "unknown"}

	err := state.executeCustomAction(action, nil)
	if err == nil {
		t.Error("executeCustomAction() should error on unknown action kind")
	}
	if err.Error() != "unknown action kind: unknown" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExecuteExecAction_Disabled(t *testing.T) {
	// Save original state and restore after test
	origState := config.IsExecAllowed()
	defer config.SetAllowExec(origState)

	config.SetAllowExec(false)

	state := &GenericState{}
	action := &config.Action{Kind: "exec", Cmd: "echo hello"}

	err := state.executeExecAction(action, nil)
	if err == nil {
		t.Error("executeExecAction() should error when exec is disabled")
	}
	if err.Error() != "exec actions disabled (use --allow-exec flag)" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExecuteSQLAction_NoRegistry(t *testing.T) {
	state := &GenericState{
		registry: nil, // No registry configured
	}
	action := &config.Action{Kind: "sql", Source: "db", Statement: "DELETE FROM tasks"}

	err := state.executeSQLAction(action, nil)
	if err == nil {
		t.Error("executeSQLAction() should error when registry is nil")
	}
	if err.Error() != "source registry not configured" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExecuteSQLAction_SourceNotFound(t *testing.T) {
	state := &GenericState{
		registry: func(name string) (source.Source, bool) {
			return nil, false
		},
	}

	action := &config.Action{Kind: "sql", Source: "missing", Statement: "DELETE FROM tasks"}

	err := state.executeSQLAction(action, nil)
	if err == nil {
		t.Error("executeSQLAction() should error when source not found")
	}
	if err.Error() != `source "missing" not found` {
		t.Errorf("unexpected error message: %v", err)
	}
}
