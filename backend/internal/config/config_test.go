package config

import "testing"

func TestValidate_CloudRunRequiresAPIKeyWithoutGoEnv(t *testing.T) {
	t.Setenv("GO_ENV", "")
	t.Setenv("K_SERVICE", "backend")

	c := &Config{Port: "8080", LogLevel: "info", YouTubeAPIKey: "yt", FrontendOrigin: "https://example.com"}
	if err := c.Validate(); err == nil {
		t.Fatal("Expected error when API_KEY is empty on Cloud Run")
	}
}

func TestValidate_LocalAllowsEmptyAPIKey(t *testing.T) {
	t.Setenv("GO_ENV", "")
	t.Setenv("K_SERVICE", "")

	c := &Config{Port: "8080", LogLevel: "info"}
	if err := c.Validate(); err != nil {
		t.Fatalf("Expected no error for local development, got %v", err)
	}
}
