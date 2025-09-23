package http

import (
	"encoding/json"
	"testing"
)

func TestStatusResponse(t *testing.T) {
	response := StatusResponse{
		Status:       "ACTIVE",
		Count:        10,
		VideoID:      "test123",
		LiveChatID:   "chat456",
		StartedAt:    "2023-01-01T00:00:00Z",
		EndedAt:      nil,
		LastPulledAt: "2023-01-01T01:00:00Z",
	}

	// JSON marshaling test
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal StatusResponse: %v", err)
	}

	// JSON unmarshaling test
	var unmarshaled StatusResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal StatusResponse: %v", err)
	}

	// Verify fields
	if unmarshaled.Status != response.Status {
		t.Errorf("Expected Status %q, got %q", response.Status, unmarshaled.Status)
	}
	if unmarshaled.Count != response.Count {
		t.Errorf("Expected Count %d, got %d", response.Count, unmarshaled.Count)
	}
	if unmarshaled.VideoID != response.VideoID {
		t.Errorf("Expected VideoID %q, got %q", response.VideoID, unmarshaled.VideoID)
	}
	if unmarshaled.LiveChatID != response.LiveChatID {
		t.Errorf("Expected LiveChatID %q, got %q", response.LiveChatID, unmarshaled.LiveChatID)
	}
}

func TestSwitchVideoResponse(t *testing.T) {
	response := SwitchVideoResponse{
		Status:     "ACTIVE",
		VideoID:    "test123",
		LiveChatID: "chat456",
		StartedAt:  "2023-01-01T00:00:00Z",
	}

	// JSON marshaling test
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal SwitchVideoResponse: %v", err)
	}

	// JSON unmarshaling test
	var unmarshaled SwitchVideoResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal SwitchVideoResponse: %v", err)
	}

	// Verify fields
	if unmarshaled.Status != response.Status {
		t.Errorf("Expected Status %q, got %q", response.Status, unmarshaled.Status)
	}
	if unmarshaled.VideoID != response.VideoID {
		t.Errorf("Expected VideoID %q, got %q", response.VideoID, unmarshaled.VideoID)
	}
	if unmarshaled.LiveChatID != response.LiveChatID {
		t.Errorf("Expected LiveChatID %q, got %q", response.LiveChatID, unmarshaled.LiveChatID)
	}
	if unmarshaled.StartedAt != response.StartedAt {
		t.Errorf("Expected StartedAt %v, got %v", response.StartedAt, unmarshaled.StartedAt)
	}
}

func TestPullResponse(t *testing.T) {
	response := PullResponse{
		AddedCount:            5,
		AutoReset:             true,
		PollingIntervalMillis: 5000,
	}

	// JSON marshaling test
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal PullResponse: %v", err)
	}

	// JSON unmarshaling test
	var unmarshaled PullResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal PullResponse: %v", err)
	}

	// Verify fields
	if unmarshaled.AddedCount != response.AddedCount {
		t.Errorf("Expected AddedCount %d, got %d", response.AddedCount, unmarshaled.AddedCount)
	}
	if unmarshaled.AutoReset != response.AutoReset {
		t.Errorf("Expected AutoReset %v, got %v", response.AutoReset, unmarshaled.AutoReset)
	}
	if unmarshaled.PollingIntervalMillis != response.PollingIntervalMillis {
		t.Errorf("Expected PollingIntervalMillis %d, got %d", response.PollingIntervalMillis, unmarshaled.PollingIntervalMillis)
	}
}

func TestResetResponse(t *testing.T) {
	response := ResetResponse{
		Status: "WAITING",
	}

	// JSON marshaling test
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal ResetResponse: %v", err)
	}

	// JSON unmarshaling test
	var unmarshaled ResetResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ResetResponse: %v", err)
	}

	// Verify fields
	if unmarshaled.Status != response.Status {
		t.Errorf("Expected Status %q, got %q", response.Status, unmarshaled.Status)
	}
}

// Test backward compatibility: ensure new structures produce the same JSON as old map[string]interface{}
func TestResponseStructuresBackwardCompatibility(t *testing.T) {
	t.Run("StatusResponse vs map", func(t *testing.T) {
		// Old map approach
		oldMap := map[string]interface{}{
			"status":       "ACTIVE",
			"count":        10,
			"videoId":      "test123",
			"liveChatId":   "chat456",
			"startedAt":    "2023-01-01T00:00:00Z",
			"endedAt":      nil,
			"lastPulledAt": "2023-01-01T01:00:00Z",
		}

		// New struct approach
		newStruct := StatusResponse{
			Status:       "ACTIVE",
			Count:        10,
			VideoID:      "test123",
			LiveChatID:   "chat456",
			StartedAt:    "2023-01-01T00:00:00Z",
			EndedAt:      nil,
			LastPulledAt: "2023-01-01T01:00:00Z",
		}

		// Marshal both
		oldJSON, err := json.Marshal(oldMap)
		if err != nil {
			t.Fatalf("Failed to marshal old map: %v", err)
		}

		newJSON, err := json.Marshal(newStruct)
		if err != nil {
			t.Fatalf("Failed to marshal new struct: %v", err)
		}

		// Compare JSON (order-independent comparison)
		var oldUnmarshaled, newUnmarshaled map[string]interface{}
		if err := json.Unmarshal(oldJSON, &oldUnmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal old JSON: %v", err)
		}
		if err := json.Unmarshal(newJSON, &newUnmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal new JSON: %v", err)
		}

		// Compare all fields
		for key, oldValue := range oldUnmarshaled {
			newValue, exists := newUnmarshaled[key]
			if !exists {
				t.Errorf("Missing key %q in new struct", key)
				continue
			}
			if oldValue != newValue {
				t.Errorf("Value mismatch for key %q: old=%v, new=%v", key, oldValue, newValue)
			}
		}
	})

	t.Run("PullResponse vs map", func(t *testing.T) {
		// Old map approach
		oldMap := map[string]interface{}{
			"addedCount":            5,
			"autoReset":             true,
			"pollingIntervalMillis": 5000,
		}

		// New struct approach
		newStruct := PullResponse{
			AddedCount:            5,
			AutoReset:             true,
			PollingIntervalMillis: 5000,
		}

		// Marshal both
		oldJSON, err := json.Marshal(oldMap)
		if err != nil {
			t.Fatalf("Failed to marshal old map: %v", err)
		}

		newJSON, err := json.Marshal(newStruct)
		if err != nil {
			t.Fatalf("Failed to marshal new struct: %v", err)
		}

		// Compare JSON strings
		if string(oldJSON) != string(newJSON) {
			t.Errorf("JSON output differs:\nOld: %s\nNew: %s", oldJSON, newJSON)
		}
	})
}