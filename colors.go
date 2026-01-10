package main

import (
	"strings"
)

// Color represents a capacitor band color
type Color int

const (
	ColorBlack Color = iota
	ColorBrown
	ColorRed
	ColorOrange
	ColorYellow
	ColorGreen
	ColorBlue
	ColorViolet
	ColorGrey
	ColorWhite
	ColorGold
	ColorSilver
)

// ColorInfo contains all information about a color band
type ColorInfo struct {
	Name       string
	Digit      int     // For bands 1-2 (0-9)
	Multiplier float64 // For band 3
	HexColor   string  // For terminal display
	ValidDigit bool    // Can be used as digit (bands 1-2)
	ValidMult  bool    // Can be used as multiplier (band 3)
	ValidTol   bool    // Can be used as tolerance (band 4)
}

// colorMap maps color names to their properties
var colorMap = map[Color]ColorInfo{
	ColorBlack: {
		Name:       "Black",
		Digit:      0,
		Multiplier: 1,
		HexColor:   "#000000",
		ValidDigit: true,
		ValidMult:  true,
		ValidTol:   true,
	},
	ColorBrown: {
		Name:       "Brown",
		Digit:      1,
		Multiplier: 10,
		HexColor:   "#8B4513",
		ValidDigit: true,
		ValidMult:  true,
		ValidTol:   true,
	},
	ColorRed: {
		Name:       "Red",
		Digit:      2,
		Multiplier: 100,
		HexColor:   "#FF0000",
		ValidDigit: true,
		ValidMult:  true,
		ValidTol:   true,
	},
	ColorOrange: {
		Name:       "Orange",
		Digit:      3,
		Multiplier: 1000,
		HexColor:   "#FF8C00",
		ValidDigit: true,
		ValidMult:  true,
		ValidTol:   true,
	},
	ColorYellow: {
		Name:       "Yellow",
		Digit:      4,
		Multiplier: 10000,
		HexColor:   "#FFFF00",
		ValidDigit: true,
		ValidMult:  true,
		ValidTol:   true,
	},
	ColorGreen: {
		Name:       "Green",
		Digit:      5,
		Multiplier: 100000,
		HexColor:   "#00FF00",
		ValidDigit: true,
		ValidMult:  true,
		ValidTol:   true,
	},
	ColorBlue: {
		Name:       "Blue",
		Digit:      6,
		Multiplier: 1000000,
		HexColor:   "#0000FF",
		ValidDigit: true,
		ValidMult:  true,
		ValidTol:   false,
	},
	ColorViolet: {
		Name:       "Violet",
		Digit:      7,
		Multiplier: 10000000,
		HexColor:   "#9400D3",
		ValidDigit: true,
		ValidMult:  true,
		ValidTol:   false,
	},
	ColorGrey: {
		Name:       "Grey",
		Digit:      8,
		Multiplier: 0.01,
		HexColor:   "#808080",
		ValidDigit: true,
		ValidMult:  true,
		ValidTol:   true,
	},
	ColorWhite: {
		Name:       "White",
		Digit:      9,
		Multiplier: 0.1,
		HexColor:   "#FFFFFF",
		ValidDigit: true,
		ValidMult:  true,
		ValidTol:   true,
	},
	ColorGold: {
		Name:       "Gold",
		Digit:      -1, // Not valid for digit bands
		Multiplier: 0.1,
		HexColor:   "#FFD700",
		ValidDigit: false,
		ValidMult:  true,
		ValidTol:   true,
	},
	ColorSilver: {
		Name:       "Silver",
		Digit:      -1, // Not valid for digit bands
		Multiplier: 0.01,
		HexColor:   "#C0C0C0",
		ValidDigit: false,
		ValidMult:  true,
		ValidTol:   true,
	},
}

// ToleranceInfo represents tolerance specifications
type ToleranceInfo struct {
	PercentHigh float64 // For asymmetric tolerance (Grey: +80%)
	PercentLow  float64 // For asymmetric tolerance (Grey: -20%)
	Symmetric   bool    // True if tolerance is symmetric (most colors)
	AbsolutePF  float64 // For capacitance <= 10pF
}

// toleranceMap maps colors to tolerance values
// For capacitance > 10pF: use percentage
// For capacitance <= 10pF: use absolute pF value
var toleranceMap = map[Color]ToleranceInfo{
	ColorBlack: {
		PercentHigh: 20,
		PercentLow:  20,
		Symmetric:   true,
		AbsolutePF:  0, // Not used for small caps
	},
	ColorBrown: {
		PercentHigh: 1,
		PercentLow:  1,
		Symmetric:   true,
		AbsolutePF:  0.1,
	},
	ColorRed: {
		PercentHigh: 2,
		PercentLow:  2,
		Symmetric:   true,
		AbsolutePF:  0.25,
	},
	ColorOrange: {
		PercentHigh: 3,
		PercentLow:  3,
		Symmetric:   true,
		AbsolutePF:  0, // Not used for small caps
	},
	ColorYellow: {
		PercentHigh: 4,
		PercentLow:  4,
		Symmetric:   true,
		AbsolutePF:  0, // Not used for small caps
	},
	ColorGreen: {
		PercentHigh: 5,
		PercentLow:  5,
		Symmetric:   true,
		AbsolutePF:  0.5,
	},
	ColorGold: {
		PercentHigh: 5,
		PercentLow:  5,
		Symmetric:   true,
		AbsolutePF:  0, // Not used for small caps
	},
	ColorWhite: {
		PercentHigh: 10,
		PercentLow:  10,
		Symmetric:   true,
		AbsolutePF:  1.0,
	},
	ColorSilver: {
		PercentHigh: 10,
		PercentLow:  10,
		Symmetric:   true,
		AbsolutePF:  0, // Not used for small caps
	},
	ColorGrey: {
		PercentHigh: 80,    // +80%
		PercentLow:  20,    // -20%
		Symmetric:   false, // Asymmetric
		AbsolutePF:  0,     // Not used for small caps
	},
}

// temperatureCoefficientMap maps colors to temperature coefficients (×10⁻⁶ /°C)
var temperatureCoefficientMap = map[Color]int{
	ColorBrown:  -33,
	ColorRed:    -75,
	ColorOrange: -150,
	ColorYellow: -220,
	ColorGreen:  -330,
	ColorBlue:   -470,
	ColorViolet: -750,
}

// ParseColor converts a string input to a Color
func ParseColor(input string) (Color, bool) {
	input = strings.ToLower(strings.TrimSpace(input))

	colorNameMap := map[string]Color{
		"black":  ColorBlack,
		"brown":  ColorBrown,
		"red":    ColorRed,
		"orange": ColorOrange,
		"yellow": ColorYellow,
		"green":  ColorGreen,
		"blue":   ColorBlue,
		"violet": ColorViolet,
		"grey":   ColorGrey,
		"gray":   ColorGrey, // Alternative spelling
		"white":  ColorWhite,
		"gold":   ColorGold,
		"silver": ColorSilver,
	}

	color, exists := colorNameMap[input]
	return color, exists
}

// GetColorInfo returns information about a color
func GetColorInfo(c Color) ColorInfo {
	return colorMap[c]
}

// GetToleranceInfo returns tolerance information for a color
func GetToleranceInfo(c Color) (ToleranceInfo, bool) {
	info, exists := toleranceMap[c]
	return info, exists
}

// GetTempCoefficient returns temperature coefficient for a color
func GetTempCoefficient(c Color) (int, bool) {
	coeff, exists := temperatureCoefficientMap[c]
	return coeff, exists
}

// AllColorNames returns a list of all valid color names
func AllColorNames() []string {
	return []string{
		"Black", "Brown", "Red", "Orange", "Yellow", "Green",
		"Blue", "Violet", "Grey", "White", "Gold", "Silver",
	}
}
