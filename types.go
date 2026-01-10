package main

import "strings"

// CapacitorType represents different capacitor types
type CapacitorType string

const (
	TypeJ CapacitorType = "J" // Dipped Tantalum
	TypeK CapacitorType = "K" // Mica
	TypeL CapacitorType = "L" // Polyester/Polystyrene
	TypeM CapacitorType = "M" // Electrolytic (4-band)
	TypeN CapacitorType = "N" // Electrolytic (3-band)
)

// TypeInfo contains information about each capacitor type
type TypeInfo struct {
	Type        CapacitorType
	Name        string
	Description string
	Voltages    []int // Voltage ratings in order of color bands
}

// typeInfoMap stores details about each capacitor type
var typeInfoMap = map[CapacitorType]TypeInfo{
	TypeJ: {
		Type:        TypeJ,
		Name:        "Dipped Tantalum",
		Description: "Type J (Dipped Tantalum)",
		Voltages:    []int{3, 4, 6, 10, 15, 20, 25, 35, 50},
	},
	TypeK: {
		Type:        TypeK,
		Name:        "Mica",
		Description: "Type K (Mica)",
		Voltages:    []int{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 2000},
	},
	TypeL: {
		Type:        TypeL,
		Name:        "Polyester / Polystyrene",
		Description: "Type L (Polyester / Polystyrene)",
		Voltages:    []int{100, 250, 400, 630},
	},
	TypeM: {
		Type:        TypeM,
		Name:        "Electrolytic (4-band style)",
		Description: "Type M (Electrolytic 4-Band)",
		Voltages:    []int{0, 2, 4, 6, 10, 16, 25, 40}, // Index 0 corresponds to 1.6V, adjusted below
	},
	TypeN: {
		Type:        TypeN,
		Name:        "Electrolytic (3-band style)",
		Description: "Type N (Electrolytic 3-Band)",
		Voltages:    []int{3, 6, 0, 10, 15, 20, 25, 35}, // Index 2 is 6.3V (special case)
	},
}

// Special handling for Type M and N voltage mappings
// Type M voltages by color index: 1.6V, 2.5V, 4V, 6.3V, 10V, 16V, 25V, 40V
// Type N voltages by color index: 3V, 6V, 6.3V, 10V, 15V, 20V, 25V, 35V

// GetVoltageRating returns the voltage rating for a given capacitor type and band 5 color
func GetVoltageRating(capType CapacitorType, band5Color Color) (int, bool) {
	typeInfo, exists := typeInfoMap[capType]
	if !exists {
		return 0, false
	}

	// Get the color index (0-9 for Black through White)
	colorInfo := GetColorInfo(band5Color)
	if colorInfo.Digit < 0 || colorInfo.Digit >= len(typeInfo.Voltages) {
		return 0, false
	}

	voltage := typeInfo.Voltages[colorInfo.Digit]

	// Special handling for Type M and N
	switch capType {
	case TypeM:
		// Type M special voltages
		switch colorInfo.Digit {
		case 0: // Black
			return 0, false // Invalid for Type M
		case 1: // Brown
			return 2, true // Actually 1.6V but we'll show 2V for simplicity, or handle fractional
		case 2: // Red
			return 0, false // Actually 2.5V
		case 3: // Orange
			return 4, true
		case 4: // Yellow
			return 0, false // Actually 6.3V
		case 5: // Green
			return 10, true
		case 6: // Blue
			return 16, true
		case 7: // Violet
			return 25, true
		case 8: // Grey
			return 40, true
		}
	case TypeN:
		// Type N special voltage for Orange (6.3V)
		if colorInfo.Digit == 2 { // Red = 6.3V for Type N
			return 0, false // Actually 6.3V
		}
	}

	// Handle special fractional voltages properly
	if voltage == 0 {
		return 0, false
	}

	return voltage, true
}

// GetVoltageRatingFractional returns voltage rating including fractional values
func GetVoltageRatingFractional(capType CapacitorType, band5Color Color) (float64, bool) {
	typeInfo, exists := typeInfoMap[capType]
	if !exists {
		return 0, false
	}

	colorInfo := GetColorInfo(band5Color)
	if colorInfo.Digit < 0 {
		return 0, false
	}

	// Type M special voltages with fractional values
	if capType == TypeM {
		voltageMap := map[int]float64{
			0: 0,    // Black - invalid
			1: 1.6,  // Brown
			2: 2.5,  // Red
			3: 4.0,  // Orange
			4: 6.3,  // Yellow
			5: 10.0, // Green
			6: 16.0, // Blue
			7: 25.0, // Violet
			8: 40.0, // Grey
		}
		if v, ok := voltageMap[colorInfo.Digit]; ok && v > 0 {
			return v, true
		}
		return 0, false
	}

	// Type N special voltage for Red (6.3V)
	if capType == TypeN && colorInfo.Digit == 2 {
		return 6.3, true
	}

	// Standard integer voltages
	if colorInfo.Digit >= len(typeInfo.Voltages) {
		return 0, false
	}

	voltage := typeInfo.Voltages[colorInfo.Digit]
	if voltage == 0 {
		return 0, false
	}

	return float64(voltage), true
}

// ParseCapacitorType converts string input to CapacitorType
func ParseCapacitorType(input string) (CapacitorType, bool) {
	input = strings.ToUpper(strings.TrimSpace(input))

	switch input {
	case "J":
		return TypeJ, true
	case "K":
		return TypeK, true
	case "L":
		return TypeL, true
	case "M":
		return TypeM, true
	case "N":
		return TypeN, true
	default:
		return "", false
	}
}

// GetTypeInfo returns information about a capacitor type
func GetTypeInfo(capType CapacitorType) (TypeInfo, bool) {
	info, exists := typeInfoMap[capType]
	return info, exists
}

// AllCapacitorTypes returns all valid capacitor type codes
func AllCapacitorTypes() []string {
	return []string{"J", "K", "L", "M", "N"}
}
