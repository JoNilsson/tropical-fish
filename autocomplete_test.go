package main

import (
	"testing"
)

// TestGetColorSuggestion tests the autocomplete suggestion function
func TestGetColorSuggestion(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		bandNum        int
		expectedSuffix string // The part that should be suggested
	}{
		{
			name:           "Empty input",
			input:          "",
			bandNum:        1,
			expectedSuffix: "",
		},
		{
			name:           "Single letter 'r' suggests 'ed'",
			input:          "r",
			bandNum:        1,
			expectedSuffix: "ed", // Completes to "Red"
		},
		{
			name:           "Partial 'bro' suggests 'wn'",
			input:          "bro",
			bandNum:        1,
			expectedSuffix: "wn", // Completes to "Brown"
		},
		{
			name:           "Complete 'red' no suggestion",
			input:          "red",
			bandNum:        1,
			expectedSuffix: "",
		},
		{
			name:           "Case insensitive 'R' suggests 'ed'",
			input:          "R",
			bandNum:        1,
			expectedSuffix: "ed",
		},
		{
			name:           "Violet starts with 'v'",
			input:          "v",
			bandNum:        1,
			expectedSuffix: "iolet",
		},
		{
			name:           "Orange starts with 'o'",
			input:          "o",
			bandNum:        1,
			expectedSuffix: "range",
		},
		{
			name:           "Grey needs 'gre' to disambiguate from Green",
			input:          "gre",
			bandNum:        1,
			expectedSuffix: "en", // Green comes first alphabetically
		},
		{
			name:           "Green suggested for 'g' on band 1",
			input:          "g",
			bandNum:        1,
			expectedSuffix: "reen", // Green comes before Grey alphabetically
		},
		{
			name:           "Green suggested for 'g' on band 3 (not Gold)",
			input:          "g",
			bandNum:        3,
			expectedSuffix: "reen", // Green comes before Gold alphabetically
		},
		{
			name:           "Gold needs 'go' to disambiguate from Green",
			input:          "go",
			bandNum:        3,
			expectedSuffix: "ld",
		},
		{
			name:           "Silver not suggested for band 2",
			input:          "s",
			bandNum:        2,
			expectedSuffix: "", // No valid colors start with 's' for digit bands
		},
		{
			name:           "Silver suggested for band 4",
			input:          "s",
			bandNum:        4,
			expectedSuffix: "ilver", // Tolerance band can use Silver
		},
		{
			name:           "Yellow with 'y'",
			input:          "y",
			bandNum:        1,
			expectedSuffix: "ellow",
		},
		{
			name:           "White with 'w'",
			input:          "w",
			bandNum:        1,
			expectedSuffix: "hite",
		},
		{
			name:           "Black suggested for 'bl'",
			input:          "bl",
			bandNum:        1,
			expectedSuffix: "ack", // Black comes before Blue alphabetically
		},
		{
			name:           "Blue needs 'blu' to disambiguate from Black",
			input:          "blu",
			bandNum:        1,
			expectedSuffix: "e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetColorSuggestion(tt.input, tt.bandNum)
			if result != tt.expectedSuffix {
				t.Errorf("GetColorSuggestion(%q, %d) = %q, want %q", tt.input, tt.bandNum, result, tt.expectedSuffix)
			}
		})
	}
}

// TestGetFullColorFromInput tests combining input with suggestion
func TestGetFullColorFromInput(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		suggestion string
		expected   string
	}{
		{
			name:       "Combine 'r' and 'ed'",
			input:      "r",
			suggestion: "ed",
			expected:   "red",
		},
		{
			name:       "Combine 'bro' and 'wn'",
			input:      "bro",
			suggestion: "wn",
			expected:   "brown",
		},
		{
			name:       "No suggestion",
			input:      "red",
			suggestion: "",
			expected:   "red",
		},
		{
			name:       "Empty input with suggestion",
			input:      "",
			suggestion: "black",
			expected:   "black",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFullColorFromInput(tt.input, tt.suggestion)
			if result != tt.expected {
				t.Errorf("GetFullColorFromInput(%q, %q) = %q, want %q", tt.input, tt.suggestion, result, tt.expected)
			}
		})
	}
}

// TestAutocompleteBandValidation ensures suggestions respect band restrictions
func TestAutocompleteBandValidation(t *testing.T) {
	// Test that Gold/Silver are not suggested for bands 1-2
	band1Suggestion := GetColorSuggestion("go", 1)
	if band1Suggestion == "ld" {
		t.Error("Gold should not be suggested for band 1 (digit band)")
	}

	band2Suggestion := GetColorSuggestion("si", 2)
	if band2Suggestion == "lver" {
		t.Error("Silver should not be suggested for band 2 (digit band)")
	}

	// Test that Gold/Silver are available for other bands
	// Note: "g" will suggest "reen" (Green) first, need "go" for Gold
	band3Suggestion := GetColorSuggestion("go", 3)
	if band3Suggestion != "ld" {
		t.Errorf("Gold should be suggested for band 3 with 'go', got %q", band3Suggestion)
	}

	band4Suggestion := GetColorSuggestion("si", 4)
	if band4Suggestion != "lver" {
		t.Errorf("Silver should be suggested for band 4 with 'si', got %q", band4Suggestion)
	}
}
