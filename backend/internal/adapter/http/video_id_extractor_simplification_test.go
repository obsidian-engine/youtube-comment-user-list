package http

import (
	"testing"
)

// TestVideoIDValidationSimplification tests that the simplified validation logic
// correctly accepts valid 11-character video IDs without false rejections
func TestVideoIDValidationSimplification(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
		reason   string
	}{
		{
			name:     "valid video ID starting with 'test'",
			input:    "testVideoID",
			expected: true,
			reason:   "Should accept valid 11-character ID even if it starts with 'test'",
		},
		{
			name:     "valid video ID containing 'invalid'",
			input:    "invalidID11",
			expected: true,
			reason:   "Should accept valid 11-character ID even if it contains 'invalid'",
		},
		{
			name:     "exactly 'invalid-url' (11 characters)",
			input:    "invalid-url",
			expected: true,
			reason:   "Should accept valid 11-character string regardless of content",
		},
		{
			name:     "valid alphanumeric 11-char ID",
			input:    "dQw4w9WgXcQ",
			expected: true,
			reason:   "Should accept standard YouTube video ID",
		},
		{
			name:     "too short",
			input:    "short",
			expected: false,
			reason:   "Should reject strings shorter than 11 characters",
		},
		{
			name:     "too long",
			input:    "toolongstring",
			expected: false,
			reason:   "Should reject strings longer than 11 characters",
		},
		{
			name:     "11 chars with invalid characters",
			input:    "invalid@url",
			expected: false,
			reason:   "Should reject IDs with invalid characters",
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
			reason:   "Should reject empty string",
		},
		{
			name:     "11 chars with underscores and hyphens",
			input:    "valid_id-11",
			expected: true,
			reason:   "Should accept IDs with underscores and hyphens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidVideoIDSimplified(tt.input)
			if result != tt.expected {
				t.Errorf("isValidVideoIDSimplified(%q) = %v, expected %v. Reason: %s", 
					tt.input, result, tt.expected, tt.reason)
			}
		})
	}
}

// isValidVideoIDSimplified is the proposed simplified validation logic
func isValidVideoIDSimplified(input string) bool {
	// YouTube video IDs are exactly 11 characters long
	if len(input) != 11 {
		return false
	}
	
	// Valid characters: a-z, A-Z, 0-9, underscore, hyphen
	for _, char := range input {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}
	
	return true
}

// Test that the current implementation rejects valid IDs (demonstrating the problem)
func TestCurrentImplementationProblems(t *testing.T) {
	problematicCases := []struct {
		input  string
		reason string
	}{
		{
			input:  "testVideoID",
			reason: "Current logic incorrectly rejects IDs starting with 'test'",
		},
		{
			input:  "invalidID11",
			reason: "Current logic incorrectly rejects IDs containing 'invalid'",
		},
		{
			input:  "invalid-url",
			reason: "Current logic incorrectly rejects this exact string",
		},
	}

	for _, tc := range problematicCases {
		t.Run("current_rejects_"+tc.input, func(t *testing.T) {
			// Test current implementation
			currentResult := isValidVideoID(tc.input)
			
			// Test simplified implementation
			simplifiedResult := isValidVideoIDSimplified(tc.input)
			
			// Current should reject (false), simplified should accept (true)
			if currentResult {
				t.Logf("Current implementation unexpectedly accepts %q", tc.input)
			}
			if !simplifiedResult {
				t.Errorf("Simplified implementation should accept %q, but it doesn't", tc.input)
			}
			
			t.Logf("Current: %v, Simplified: %v for %q (%s)", 
				currentResult, simplifiedResult, tc.input, tc.reason)
		})
	}
}