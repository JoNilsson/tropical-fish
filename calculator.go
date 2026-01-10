package main

import (
	"fmt"
	"math"
)

// CapacitorReading represents the parsed bands from user input
type CapacitorReading struct {
	Band1     Color         // First digit
	Band2     Color         // Second digit
	Band3     Color         // Multiplier
	Band4     Color         // Tolerance
	Band5     Color         // Voltage rating / temp coefficient
	BandCount int           // 3, 4, or 5 bands
	CapType   CapacitorType // J, K, L, M, or N
}

// CalculationResult contains all calculated values
type CalculationResult struct {
	// Capacitance
	CapacitancePF    float64 // Raw value in picofarads
	CapacitanceValue float64 // Scaled value
	CapacitanceUnit  string  // pF, nF, µF, or mF

	// Tolerance
	ToleranceType       string  // "percentage" or "absolute"
	TolerancePercent    float64 // For percentage-based
	ToleranceAbsolutePF float64 // For absolute (small caps)
	ToleranceSymmetric  bool    // True for symmetric tolerance
	ToleranceHigh       float64 // +% for asymmetric (e.g., Grey)
	ToleranceLow        float64 // -% for asymmetric

	// Range (min/max values)
	MinValue float64
	MaxValue float64
	MinUnit  string
	MaxUnit  string

	// Voltage
	VoltageRating float64
	VoltageValid  bool

	// Temperature Coefficient
	TempCoefficient int
	TempCoeffValid  bool

	// Original reading
	Reading CapacitorReading
}

// Calculate performs all calculations for a capacitor reading
func Calculate(reading CapacitorReading) (*CalculationResult, error) {
	result := &CalculationResult{
		Reading: reading,
	}

	// Step 1: Calculate capacitance in pF
	// Formula: (Band1 × 10 + Band2) × Multiplier
	info1 := GetColorInfo(reading.Band1)
	info2 := GetColorInfo(reading.Band2)
	info3 := GetColorInfo(reading.Band3)

	if !info1.ValidDigit || !info2.ValidDigit {
		return nil, fmt.Errorf("invalid digit colors for bands 1 or 2")
	}

	if !info3.ValidMult {
		return nil, fmt.Errorf("invalid multiplier color for band 3")
	}

	baseValue := float64(info1.Digit*10 + info2.Digit)
	result.CapacitancePF = baseValue * info3.Multiplier

	// Step 2: Auto-scale units (pF → nF → µF → mF)
	result.CapacitanceValue, result.CapacitanceUnit = scaleCapacitance(result.CapacitancePF)

	// Step 3: Calculate tolerance
	if err := calculateTolerance(result); err != nil {
		return nil, err
	}

	// Step 4: Get voltage rating (if band 5 exists and band count >= 4)
	if reading.BandCount >= 4 && reading.Band5 != ColorBlack {
		voltage, valid := GetVoltageRatingFractional(reading.CapType, reading.Band5)
		result.VoltageRating = voltage
		result.VoltageValid = valid
	}

	// Step 5: Get temperature coefficient (if band 5 exists and band count == 5)
	if reading.BandCount == 5 {
		coeff, valid := GetTempCoefficient(reading.Band5)
		result.TempCoefficient = coeff
		result.TempCoeffValid = valid
	}

	return result, nil
}

// scaleCapacitance converts pF to the most appropriate unit
// Returns value and unit as separate values
func scaleCapacitance(pF float64) (float64, string) {
	// Conversion factors
	const (
		pFToNF = 1000.0       // 1 nF = 1000 pF
		pFToUF = 1000000.0    // 1 µF = 1,000,000 pF
		pFToMF = 1000000000.0 // 1 mF = 1,000,000,000 pF
	)

	// Try mF (millifarads)
	if pF >= pFToMF {
		return pF / pFToMF, "mF"
	}

	// Try µF (microfarads)
	if pF >= pFToUF {
		return pF / pFToUF, "µF"
	}

	// Try nF (nanofarads)
	if pF >= pFToNF {
		return pF / pFToNF, "nF"
	}

	// Use pF (picofarads)
	return pF, "pF"
}

// calculateTolerance computes tolerance range based on capacitance value
func calculateTolerance(result *CalculationResult) error {
	tolInfo, exists := GetToleranceInfo(result.Reading.Band4)
	if !exists {
		return fmt.Errorf("invalid tolerance color for band 4")
	}

	result.ToleranceSymmetric = tolInfo.Symmetric
	result.ToleranceHigh = tolInfo.PercentHigh
	result.ToleranceLow = tolInfo.PercentLow

	// Determine tolerance type based on capacitance value
	if result.CapacitancePF <= 10.0 {
		// Use absolute pF tolerance for small capacitors
		result.ToleranceType = "absolute"
		result.ToleranceAbsolutePF = tolInfo.AbsolutePF

		if tolInfo.AbsolutePF == 0 {
			// This color doesn't have absolute tolerance defined
			// Fall back to percentage
			result.ToleranceType = "percentage"
			result.TolerancePercent = tolInfo.PercentHigh
			result.MinValue = result.CapacitancePF * (1 - tolInfo.PercentLow/100)
			result.MaxValue = result.CapacitancePF * (1 + tolInfo.PercentHigh/100)
		} else {
			// Calculate min/max using absolute tolerance
			result.MinValue = result.CapacitancePF - tolInfo.AbsolutePF
			result.MaxValue = result.CapacitancePF + tolInfo.AbsolutePF

			// Ensure non-negative
			if result.MinValue < 0 {
				result.MinValue = 0
			}
		}
	} else {
		// Use percentage tolerance for larger capacitors
		result.ToleranceType = "percentage"
		result.TolerancePercent = tolInfo.PercentHigh

		// Calculate min/max values
		result.MinValue = result.CapacitancePF * (1 - tolInfo.PercentLow/100)
		result.MaxValue = result.CapacitancePF * (1 + tolInfo.PercentHigh/100)
	}

	// Scale min/max to appropriate units
	result.MinValue, result.MinUnit = scaleCapacitance(result.MinValue)
	result.MaxValue, result.MaxUnit = scaleCapacitance(result.MaxValue)

	return nil
}

// FormatCapacitance formats a capacitance value with unit
func FormatCapacitance(value float64, unit string) string {
	// Format with appropriate precision
	if value >= 100 {
		return fmt.Sprintf("%.1f %s", value, unit)
	} else if value >= 10 {
		return fmt.Sprintf("%.2f %s", value, unit)
	} else {
		return fmt.Sprintf("%.3f %s", value, unit)
	}
}

// FormatCapacitanceWithPF formats capacitance with both scaled unit and pF
func FormatCapacitanceWithPF(value float64, unit string, pF float64) string {
	scaled := FormatCapacitance(value, unit)

	// Don't show pF if it's already in pF
	if unit == "pF" {
		return scaled
	}

	// Format pF with appropriate precision
	var pfStr string
	if pF >= 1000 {
		pfStr = fmt.Sprintf("%.0f pF", pF)
	} else if pF >= 1 {
		pfStr = fmt.Sprintf("%.1f pF", pF)
	} else {
		pfStr = fmt.Sprintf("%.3f pF", pF)
	}

	return fmt.Sprintf("%s  (%s)", scaled, pfStr)
}

// FormatCapacitanceWithUF formats capacitance with µF value in brackets
func FormatCapacitanceWithUF(value float64, unit string, pF float64) string {
	scaled := FormatCapacitance(value, unit)

	// Calculate µF value
	uF := pF / 1000000.0

	// Format µF with appropriate precision
	var ufStr string
	if uF >= 100 {
		ufStr = fmt.Sprintf("%.1f µF", uF)
	} else if uF >= 10 {
		ufStr = fmt.Sprintf("%.2f µF", uF)
	} else if uF >= 1 {
		ufStr = fmt.Sprintf("%.3f µF", uF)
	} else if uF >= 0.001 {
		ufStr = fmt.Sprintf("%.6f µF", uF)
	} else {
		ufStr = fmt.Sprintf("%.9f µF", uF)
	}

	// If already in µF, just return scaled
	if unit == "µF" {
		return scaled + " [" + ufStr + "]"
	}

	return fmt.Sprintf("%s [%s]", scaled, ufStr)
}

// FormatTolerance formats tolerance information
func FormatTolerance(result *CalculationResult) string {
	if result.ToleranceType == "absolute" {
		return fmt.Sprintf("±%.2f pF", result.ToleranceAbsolutePF)
	}

	if result.ToleranceSymmetric {
		return fmt.Sprintf("±%.0f%%", result.TolerancePercent)
	}

	// Asymmetric (Grey)
	return fmt.Sprintf("+%.0f%% / -%.0f%%", result.ToleranceHigh, result.ToleranceLow)
}

// FormatToleranceRange formats the min-max range
func FormatToleranceRange(result *CalculationResult) string {
	minStr := FormatCapacitance(result.MinValue, result.MinUnit)
	maxStr := FormatCapacitance(result.MaxValue, result.MaxUnit)
	return fmt.Sprintf("%s ──► %s", minStr, maxStr)
}

// FormatVoltage formats voltage rating
func FormatVoltage(result *CalculationResult) string {
	if !result.VoltageValid {
		return "N/A"
	}

	// Check if voltage is fractional
	if result.VoltageRating != math.Floor(result.VoltageRating) {
		return fmt.Sprintf("%.1f V", result.VoltageRating)
	}

	return fmt.Sprintf("%.0f V", result.VoltageRating)
}

// FormatTempCoefficient formats temperature coefficient
func FormatTempCoefficient(result *CalculationResult) string {
	if !result.TempCoeffValid {
		return "N/A"
	}

	return fmt.Sprintf("%d × 10⁻⁶ /°C", result.TempCoefficient)
}
