package livemdtools

import (
	"encoding/json"
	"testing"
)

func TestPageState(t *testing.T) {
	t.Run("new page state", func(t *testing.T) {
		page := New("test")
		ps := NewPageState(page)

		if ps.CurrentStep != 0 {
			t.Errorf("CurrentStep = %d, want 0", ps.CurrentStep)
		}
		if len(ps.InteractiveStates) != 0 {
			t.Error("InteractiveStates should be empty")
		}
		if len(ps.CodeEdits) != 0 {
			t.Error("CodeEdits should be empty")
		}
	})

	t.Run("multi-step navigation", func(t *testing.T) {
		page := New("test")
		page.Config.MultiStep = true
		page.Config.StepCount = 3

		ps := NewPageState(page)

		// Next step
		if err := ps.HandleAction("nextStep", nil); err != nil {
			t.Fatalf("nextStep error: %v", err)
		}
		if ps.CurrentStep != 1 {
			t.Errorf("CurrentStep = %d, want 1", ps.CurrentStep)
		}
		if len(ps.CompletedSteps) != 1 || ps.CompletedSteps[0] != 0 {
			t.Errorf("CompletedSteps = %v, want [0]", ps.CompletedSteps)
		}

		// Another next step
		if err := ps.HandleAction("nextStep", nil); err != nil {
			t.Fatalf("nextStep error: %v", err)
		}
		if ps.CurrentStep != 2 {
			t.Errorf("CurrentStep = %d, want 2", ps.CurrentStep)
		}

		// Can't go beyond last step
		if err := ps.HandleAction("nextStep", nil); err != nil {
			t.Fatalf("nextStep error: %v", err)
		}
		if ps.CurrentStep != 2 {
			t.Errorf("CurrentStep = %d, should stay at 2", ps.CurrentStep)
		}

		// Previous step
		if err := ps.HandleAction("prevStep", nil); err != nil {
			t.Fatalf("prevStep error: %v", err)
		}
		if ps.CurrentStep != 1 {
			t.Errorf("CurrentStep = %d, want 1", ps.CurrentStep)
		}

		// Previous again
		if err := ps.HandleAction("prevStep", nil); err != nil {
			t.Fatalf("prevStep error: %v", err)
		}
		if ps.CurrentStep != 0 {
			t.Errorf("CurrentStep = %d, want 0", ps.CurrentStep)
		}

		// Can't go below 0
		if err := ps.HandleAction("prevStep", nil); err != nil {
			t.Fatalf("prevStep error: %v", err)
		}
		if ps.CurrentStep != 0 {
			t.Errorf("CurrentStep = %d, should stay at 0", ps.CurrentStep)
		}
	})

	t.Run("save code edit", func(t *testing.T) {
		page := New("test")
		ps := NewPageState(page)

		data := map[string]interface{}{
			"blockID": "wasm-1",
			"code":    "package main\nfunc main() {}",
		}

		if err := ps.HandleAction("saveCodeEdit", data); err != nil {
			t.Fatalf("saveCodeEdit error: %v", err)
		}

		if ps.CodeEdits["wasm-1"] != "package main\nfunc main() {}" {
			t.Errorf("CodeEdits[wasm-1] = %q, want edited code", ps.CodeEdits["wasm-1"])
		}
	})
}

func TestMessageRouter(t *testing.T) {
	t.Run("route page action", func(t *testing.T) {
		page := New("test")
		page.Config.MultiStep = true
		page.Config.StepCount = 3

		ps := NewPageState(page)
		router := NewMessageRouter(ps)

		envelope := &MessageEnvelope{
			BlockID: "_page",
			Action:  "nextStep",
			Data:    json.RawMessage("{}"),
		}

		resp, err := router.Route(envelope)
		if err != nil {
			t.Fatalf("Route error: %v", err)
		}

		if resp.BlockID != "_page" {
			t.Errorf("Response BlockID = %s, want _page", resp.BlockID)
		}
		if !resp.Meta["success"].(bool) {
			t.Error("Expected success = true")
		}
		if ps.CurrentStep != 1 {
			t.Errorf("CurrentStep = %d, want 1", ps.CurrentStep)
		}
	})

	t.Run("route interactive block action", func(t *testing.T) {
		page := New("test")
		ps := NewPageState(page)

		// Register an interactive block state
		ps.InteractiveStates["counter"] = nil

		router := NewMessageRouter(ps)

		envelope := &MessageEnvelope{
			BlockID: "counter",
			Action:  "increment",
			Data:    json.RawMessage("{}"),
		}

		resp, err := router.Route(envelope)
		if err != nil {
			t.Fatalf("Route error: %v", err)
		}

		if resp.BlockID != "counter" {
			t.Errorf("Response BlockID = %s, want counter", resp.BlockID)
		}
		if !resp.Meta["success"].(bool) {
			t.Error("Expected success = true")
		}
	})

	t.Run("unknown block error", func(t *testing.T) {
		page := New("test")
		ps := NewPageState(page)
		router := NewMessageRouter(ps)

		envelope := &MessageEnvelope{
			BlockID: "nonexistent",
			Action:  "increment",
			Data:    json.RawMessage("{}"),
		}

		_, err := router.Route(envelope)
		if err == nil {
			t.Fatal("Expected error for unknown block")
		}
	})

	t.Run("save code edit via router", func(t *testing.T) {
		page := New("test")
		ps := NewPageState(page)
		router := NewMessageRouter(ps)

		data := map[string]interface{}{
			"blockID": "wasm-1",
			"code":    "package main",
		}
		jsonData, _ := json.Marshal(data)

		envelope := &MessageEnvelope{
			BlockID: "_page",
			Action:  "saveCodeEdit",
			Data:    jsonData,
		}

		resp, err := router.Route(envelope)
		if err != nil {
			t.Fatalf("Route error: %v", err)
		}

		if !resp.Meta["success"].(bool) {
			t.Error("Expected success = true")
		}
		if ps.CodeEdits["wasm-1"] != "package main" {
			t.Error("Code edit not saved")
		}
	})
}

func TestMessageEnvelope(t *testing.T) {
	t.Run("marshal and unmarshal", func(t *testing.T) {
		original := &MessageEnvelope{
			BlockID: "counter",
			Action:  "increment",
			Data:    json.RawMessage(`{"value":1}`),
		}

		// Marshal
		jsonData, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		// Unmarshal
		var decoded MessageEnvelope
		if err := json.Unmarshal(jsonData, &decoded); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if decoded.BlockID != original.BlockID {
			t.Errorf("BlockID = %s, want %s", decoded.BlockID, original.BlockID)
		}
		if decoded.Action != original.Action {
			t.Errorf("Action = %s, want %s", decoded.Action, original.Action)
		}
	})
}

func TestResponseEnvelope(t *testing.T) {
	t.Run("marshal response", func(t *testing.T) {
		resp := &ResponseEnvelope{
			BlockID: "counter",
			Tree: map[string]interface{}{
				"s": []string{"<div>", "</div>"},
				"0": "5",
			},
			Meta: map[string]interface{}{
				"success": true,
			},
		}

		jsonData, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		// Should contain all fields
		var decoded map[string]interface{}
		if err := json.Unmarshal(jsonData, &decoded); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if decoded["blockID"] != "counter" {
			t.Error("Missing or wrong blockID")
		}
		if decoded["tree"] == nil {
			t.Error("Missing tree")
		}
		if decoded["meta"] == nil {
			t.Error("Missing meta")
		}
	})
}
