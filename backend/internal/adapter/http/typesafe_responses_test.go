package http

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTypeSafeStatusResponse_JSONSerialization(t *testing.T) {
	startedAt := time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC)
	endedAt := time.Date(2023, 12, 25, 17, 45, 0, 0, time.UTC)
	lastPulledAt := time.Date(2023, 12, 25, 17, 44, 30, 0, time.UTC)

	response := TypeSafeStatusResponse{
		Status:       "active",
		Count:        42,
		VideoID:      "abc123xyz",
		LiveChatID:   "chat456",
		StartedAt:    &startedAt,
		EndedAt:      &endedAt,
		LastPulledAt: &lastPulledAt,
	}

	// JSONシリアライゼーション
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// 結果の検証
	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// 値の検証
	if unmarshaled["status"] != "active" {
		t.Errorf("Expected status=active, got %v", unmarshaled["status"])
	}
	if unmarshaled["count"].(float64) != 42 {
		t.Errorf("Expected count=42, got %v", unmarshaled["count"])
	}
	if unmarshaled["videoId"] != "abc123xyz" {
		t.Errorf("Expected videoId=abc123xyz, got %v", unmarshaled["videoId"])
	}

	// 時刻フィールドがISO8601形式であることを確認
	if startedAtStr, ok := unmarshaled["startedAt"].(string); ok {
		if _, err := time.Parse(time.RFC3339, startedAtStr); err != nil {
			t.Errorf("startedAt is not in RFC3339 format: %v", startedAtStr)
		}
	} else {
		t.Errorf("startedAt should be a string, got %T", unmarshaled["startedAt"])
	}
}

func TestTypeSafeStatusResponse_NilTimeHandling(t *testing.T) {
	response := TypeSafeStatusResponse{
		Status:       "waiting",
		Count:        0,
		VideoID:      "",
		LiveChatID:   "",
		StartedAt:    nil,  // nilの時刻
		EndedAt:      nil,
		LastPulledAt: nil,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response with nil times: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// nil時刻フィールドがnullとしてシリアライズされることを確認
	if unmarshaled["startedAt"] != nil {
		t.Errorf("Expected startedAt=null, got %v", unmarshaled["startedAt"])
	}
	if unmarshaled["endedAt"] != nil {
		t.Errorf("Expected endedAt=null, got %v", unmarshaled["endedAt"])
	}
}

func TestTypeSafeSwitchVideoResponse_JSONSerialization(t *testing.T) {
	startedAt := time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC)

	response := TypeSafeSwitchVideoResponse{
		Status:     "active",
		VideoID:    "new_video_123",
		LiveChatID: "new_chat_456",
		StartedAt:  &startedAt,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal SwitchVideoResponse: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if unmarshaled["status"] != "active" {
		t.Errorf("Expected status=active, got %v", unmarshaled["status"])
	}
	if unmarshaled["videoId"] != "new_video_123" {
		t.Errorf("Expected videoId=new_video_123, got %v", unmarshaled["videoId"])
	}
	if unmarshaled["liveChatId"] != "new_chat_456" {
		t.Errorf("Expected liveChatId=new_chat_456, got %v", unmarshaled["liveChatId"])
	}
}

func TestBackwardCompatibility_StatusResponse(t *testing.T) {
	// 既存のStatusResponseと新しいTypeSafeStatusResponseの互換性をテスト
	startedAt := time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC)

	// 既存の形式
	oldResponse := StatusResponse{
		Status:       "active",
		Count:        42,
		VideoID:      "abc123",
		LiveChatID:   "chat456",
		StartedAt:    startedAt,  // interface{}として格納
		EndedAt:      nil,
		LastPulledAt: time.Now(),
	}

	// 新しい型安全な形式
	newResponse := TypeSafeStatusResponse{
		Status:       "active",
		Count:        42,
		VideoID:      "abc123",
		LiveChatID:   "chat456",
		StartedAt:    &startedAt,
		EndedAt:      nil,
		LastPulledAt: nil,
	}

	// 両方とも正常にJSONシリアライズできることを確認
	_, err1 := json.Marshal(oldResponse)
	_, err2 := json.Marshal(newResponse)

	if err1 != nil {
		t.Errorf("Old StatusResponse serialization failed: %v", err1)
	}
	if err2 != nil {
		t.Errorf("New TypeSafeStatusResponse serialization failed: %v", err2)
	}
}