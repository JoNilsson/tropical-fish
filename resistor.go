package main

import (
	"fmt"
	"math"
)

// ComponentType distinguishes between capacitors and resistors
type ComponentType int

const (
	ComponentCapacitor ComponentType = iota
	ComponentResistor
)

// ResistorReading represents parsed resistor bands
type ResistorReading struct {
	Band1     Color // First digit
	Band2     Color // Second digit
	Band3     Color // Third digit (5/6-band) or Multiplier (4-band)
	Band4     Color // Multiplier (5/6-band) or Tolerance (4-band)
	Band5     Color // Tolerance (5/6-band)
	Band6     Color // Temperature coefficient (6-band only)
	BandCount int   // 4, 5, or 6
}

// ResistorResult contains calculated resistor values
type ResistorResult struct {
	ResistanceOhms  float64 // Raw value in ohms
	ResistanceValue float64 // Scaled value
	ResistanceUnit  string  // Ω, kΩ, MΩ, or GΩ

	TolerancePercent float64 // Tolerance in %
	MinValue         float64 // Min resistance
	MaxValue         float64 // Max resistance
	MinUnit          string
	MaxUnit          string

	TempCoefficient int  // ppm/°C (6-band only)
	TempCoeffValid  bool

	Reading ResistorReading
}

// ResistorToleranceInfo represents tolerance specifications for resistors
type ResistorToleranceInfo struct {
	Percent float64
	Name    string
}

// resistorToleranceMap maps colors to resistor tolerance values
var resistorToleranceMap = map[Color]ResistorToleranceInfo{
	ColorBrown: {
		Percent: 1,
		Name:    "±1% (precision)",
	},
	ColorRed: {
		Percent: 2,
		Name:    "±2% (precision)",
	},
	ColorGreen: {
		Percent: 0.5,
		Name:    "±0.5% (high precision)",
	},
	ColorBlue: {
		Percent: 0.25,
		Name:    "±0.25% (high precision)",
	},
	ColorViolet: {
		Percent: 0.1,
		Name:    "±0.1% (ultra precision)",
	},
	ColorGrey: {
		Percent: 0.05,
		Name:    "±0.05% (ultra precision)",
	},
	ColorGold: {
		Percent: 5,
		Name:    "±5% (standard)",
	},
	ColorSilver: {
		Percent: 10,
		Name:    "±10% (standard)",
	},
}

// resistorTempCoefficientMap maps colors to temperature coefficients (ppm/°C)
var resistorTempCoefficientMap = map[Color]int{
	ColorBlack:  250,
	ColorBrown:  100,
	ColorRed:    50,
	ColorOrange: 15,
	ColorYellow: 25,
	ColorGreen:  20,
	ColorBlue:   10,
	ColorViolet: 5,
	ColorGrey:   1,
}

// resistorMultiplierMap extends the colorMap multipliers with resistor-specific values
// Resistors use the same multipliers but with Gold=0.1 and Silver=0.01
var resistorMultiplierMap = map[Color]float64{
	ColorBlack:  1,
	ColorBrown:  10,
	ColorRed:    100,
	ColorOrange: 1000,
	ColorYellow: 10000,
	ColorGreen:  100000,
	ColorBlue:   1000000,
	ColorViolet: 10000000,
	ColorGrey:   100000000,
	ColorWhite:  1000000000,
	ColorGold:   0.1,
	ColorSilver: 0.01,
}

// CalculateResistor performs all calculations for a resistor reading
func CalculateResistor(reading ResistorReading) (*ResistorResult, error) {
	result := &ResistorResult{
		Reading: reading,
	}

	var baseValue float64
	var multiplierColor Color

	// Calculate base value based on band count
	switch reading.BandCount {
	case 4:
		// 4-band: Band1 Band2 Multiplier Tolerance
		info1 := GetColorInfo(reading.Band1)
		info2 := GetColorInfo(reading.Band2)

		if !info1.ValidDigit || !info2.ValidDigit {
			return nil, fmt.Errorf("invalid digit colors for bands 1 or 2")
		}

		baseValue = float64(info1.Digit*10 + info2.Digit)
		multiplierColor = reading.Band3

	case 5:
		// 5-band: Band1 Band2 Band3 Multiplier Tolerance
		info1 := GetColorInfo(reading.Band1)
		info2 := GetColorInfo(reading.Band2)
		info3 := GetColorInfo(reading.Band3)

		if !info1.ValidDigit || !info2.ValidDigit || !info3.ValidDigit {
			return nil, fmt.Errorf("invalid digit colors for bands 1, 2, or 3")
		}

		baseValue = float64(info1.Digit*100 + info2.Digit*10 + info3.Digit)
		multiplierColor = reading.Band4

	case 6:
		// 6-band: Band1 Band2 Band3 Multiplier Tolerance TempCoeff
		info1 := GetColorInfo(reading.Band1)
		info2 := GetColorInfo(reading.Band2)
		info3 := GetColorInfo(reading.Band3)

		if !info1.ValidDigit || !info2.ValidDigit || !info3.ValidDigit {
			return nil, fmt.Errorf("invalid digit colors for bands 1, 2, or 3")
		}

		baseValue = float64(info1.Digit*100 + info2.Digit*10 + info3.Digit)
		multiplierColor = reading.Band4

		// Get temperature coefficient for 6-band
		tempCoeff, valid := GetResistorTempCoefficient(reading.Band6)
		result.TempCoefficient = tempCoeff
		result.TempCoeffValid = valid

	default:
		return nil, fmt.Errorf("invalid band count: %d (must be 4, 5, or 6)", reading.BandCount)
	}

	// Get multiplier
	multiplier, valid := GetResistorMultiplier(multiplierColor)
	if !valid {
		return nil, fmt.Errorf("invalid multiplier color")
	}

	// Calculate resistance in ohms
	result.ResistanceOhms = baseValue * multiplier

	// Auto-scale to appropriate unit
	result.ResistanceValue, result.ResistanceUnit = scaleResistance(result.ResistanceOhms)

	// Get tolerance
	toleranceColor := reading.Band4
	if reading.BandCount >= 5 {
		toleranceColor = reading.Band5
	}

	tolInfo, exists := GetResistorTolerance(toleranceColor)
	if !exists {
		return nil, fmt.Errorf("invalid tolerance color")
	}

	result.TolerancePercent = tolInfo.Percent

	// Calculate min/max values
	result.MinValue = result.ResistanceOhms * (1 - tolInfo.Percent/100)
	result.MaxValue = result.ResistanceOhms * (1 + tolInfo.Percent/100)

	// Scale min/max to appropriate units
	result.MinValue, result.MinUnit = scaleResistance(result.MinValue)
	result.MaxValue, result.MaxUnit = scaleResistance(result.MaxValue)

	return result, nil
}

// scaleResistance converts ohms to the most appropriate unit
// Returns value and unit as separate values
func scaleResistance(ohms float64) (float64, string) {
	// Conversion factors
	const (
		ohmsToKOhms = 1000.0         // 1 kΩ = 1,000 Ω
		ohmsToMOhms = 1000000.0      // 1 MΩ = 1,000,000 Ω
		ohmsToGOhms = 1000000000.0   // 1 GΩ = 1,000,000,000 Ω
	)

	// Try GΩ (gigohms)
	if ohms >= ohmsToGOhms {
		return ohms / ohmsToGOhms, "GΩ"
	}

	// Try MΩ (megohms)
	if ohms >= ohmsToMOhms {
		return ohms / ohmsToMOhms, "MΩ"
	}

	// Try kΩ (kilohms)
	if ohms >= ohmsToKOhms {
		return ohms / ohmsToKOhms, "kΩ"
	}

	// Use Ω (ohms)
	return ohms, "Ω"
}

// GetResistorMultiplier returns the multiplier for a given color
func GetResistorMultiplier(c Color) (float64, bool) {
	mult, exists := resistorMultiplierMap[c]
	return mult, exists
}

// GetResistorTolerance returns tolerance information for a color
func GetResistorTolerance(c Color) (ResistorToleranceInfo, bool) {
	info, exists := resistorToleranceMap[c]
	return info, exists
}

// GetResistorTempCoefficient returns temperature coefficient for a color (6-band resistors)
func GetResistorTempCoefficient(c Color) (int, bool) {
	coeff, exists := resistorTempCoefficientMap[c]
	return coeff, exists
}

// FormatResistance formats a resistance value with unit
func FormatResistance(value float64, unit string) string {
	// Format with appropriate precision
	if value >= 100 {
		return fmt.Sprintf("%.1f %s", value, unit)
	} else if value >= 10 {
		return fmt.Sprintf("%.2f %s", value, unit)
	} else {
		return fmt.Sprintf("%.3f %s", value, unit)
	}
}

// FormatResistanceWithOhms formats resistance with both scaled unit and Ω
func FormatResistanceWithOhms(value float64, unit string, ohms float64) string {
	scaled := FormatResistance(value, unit)

	// Don't show Ω if it's already in Ω
	if unit == "Ω" {
		return scaled
	}

	// Format Ω with appropriate precision
	var ohmsStr string
	if ohms >= 1000 {
		ohmsStr = fmt.Sprintf("%.0f Ω", ohms)
	} else if ohms >= 1 {
		ohmsStr = fmt.Sprintf("%.1f Ω", ohms)
	} else {
		ohmsStr = fmt.Sprintf("%.3f Ω", ohms)
	}

	return fmt.Sprintf("%s  (%s)", scaled, ohmsStr)
}

// FormatResistorTolerance formats tolerance information for resistors
func FormatResistorTolerance(result *ResistorResult) string {
	toleranceColor := result.Reading.Band4
	if result.Reading.BandCount >= 5 {
		toleranceColor = result.Reading.Band5
	}

	tolInfo, exists := GetResistorTolerance(toleranceColor)
	if !exists {
		return "N/A"
	}

	// Format percentage with appropriate precision
	if tolInfo.Percent < 1 {
		return fmt.Sprintf("±%.2f%% (%s)", tolInfo.Percent, tolInfo.Name)
	}
	return fmt.Sprintf("±%.0f%% (%s)", tolInfo.Percent, tolInfo.Name)
}

// FormatResistorToleranceRange formats the min-max range for resistors
func FormatResistorToleranceRange(result *ResistorResult) string {
	minStr := FormatResistance(result.MinValue, result.MinUnit)
	maxStr := FormatResistance(result.MaxValue, result.MaxUnit)
	return fmt.Sprintf("%s ──► %s", minStr, maxStr)
}

// FormatResistorTempCoefficient formats temperature coefficient for 6-band resistors
func FormatResistorTempCoefficient(result *ResistorResult) string {
	if !result.TempCoeffValid {
		return "N/A"
	}

	return fmt.Sprintf("%d ppm/°C", result.TempCoefficient)
}

// GetResistorBandName returns a human-readable name for each resistor band
func GetResistorBandName(bandNum int, bandCount int) string {
	switch bandCount {
	case 4:
		switch bandNum {
		case 1:
			return "First Digit"
		case 2:
			return "Second Digit"
		case 3:
			return "Multiplier"
		case 4:
			return "Tolerance"
		}
	case 5:
		switch bandNum {
		case 1:
			return "First Digit"
		case 2:
			return "Second Digit"
		case 3:
			return "Third Digit"
		case 4:
			return "Multiplier"
		case 5:
			return "Tolerance"
		}
	case 6:
		switch bandNum {
		case 1:
			return "First Digit"
		case 2:
			return "Second Digit"
		case 3:
			return "Third Digit"
		case 4:
			return "Multiplier"
		case 5:
			return "Tolerance"
		case 6:
			return "Temperature Coefficient"
		}
	}
	return fmt.Sprintf("Band %d", bandNum)
}

// GetResistorBandDescription returns a detailed description for each resistor band
func GetResistorBandDescription(bandNum int, bandCount int) string {
	switch bandCount {
	case 4:
		switch bandNum {
		case 1:
			return "First significant digit (0-9)"
		case 2:
			return "Second significant digit (0-9)"
		case 3:
			return "Multiplier (×1, ×10, ×100, etc.)"
		case 4:
			return "Tolerance (±%)"
		}
	case 5:
		switch bandNum {
		case 1:
			return "First significant digit (0-9)"
		case 2:
			return "Second significant digit (0-9)"
		case 3:
			return "Third significant digit (0-9)"
		case 4:
			return "Multiplier (×1, ×10, ×100, etc.)"
		case 5:
			return "Tolerance (±%)"
		}
	case 6:
		switch bandNum {
		case 1:
			return "First significant digit (0-9)"
		case 2:
			return "Second significant digit (0-9)"
		case 3:
			return "Third significant digit (0-9)"
		case 4:
			return "Multiplier (×1, ×10, ×100, etc.)"
		case 5:
			return "Tolerance (±%)"
		case 6:
			return "Temperature coefficient (ppm/°C)"
		}
	}
	return ""
}
