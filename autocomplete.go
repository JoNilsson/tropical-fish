package main

import (
	"strings"
)

// GetColorSuggestion returns the best autocomplete suggestion for a partial color input
// Returns empty string if no match or input is empty
func GetColorSuggestion(input string, bandNum int) string {
	if input == "" {
		return ""
	}

	input = strings.ToLower(strings.TrimSpace(input))

	// Get all valid colors
	validColors := AllColorNames()

	// Filter based on band number (bands 1-2 can't use Gold/Silver)
	if bandNum == 1 || bandNum == 2 {
		filteredColors := []string{}
		for _, color := range validColors {
			if color != "Gold" && color != "Silver" {
				filteredColors = append(filteredColors, color)
			}
		}
		validColors = filteredColors
	}

	// Find first match that starts with the input
	for _, color := range validColors {
		colorLower := strings.ToLower(color)
		if strings.HasPrefix(colorLower, input) {
			// Return the remaining part of the color (the suggestion)
			return color[len(input):]
		}
	}

	return ""
}

// GetFullColorFromInput returns the complete color name if input + suggestion form a valid color
func GetFullColorFromInput(input string, suggestion string) string {
	if suggestion == "" {
		return input
	}
	return input + suggestion
}
