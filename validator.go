package main

import "fmt"

// ValidationError represents a validation error with context
type ValidationError struct {
	BandNumber int
	Message    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Band %d: %s", e.BandNumber, e.Message)
}

// ValidateBand1 validates the first digit band
func ValidateBand1(color Color) error {
	info := GetColorInfo(color)
	if !info.ValidDigit {
		return &ValidationError{
			BandNumber: 1,
			Message:    fmt.Sprintf("%s is not valid for first digit band (must be Black-White, 0-9)", info.Name),
		}
	}
	return nil
}

// ValidateBand2 validates the second digit band
func ValidateBand2(color Color) error {
	info := GetColorInfo(color)
	if !info.ValidDigit {
		return &ValidationError{
			BandNumber: 2,
			Message:    fmt.Sprintf("%s is not valid for second digit band (must be Black-White, 0-9)", info.Name),
		}
	}
	return nil
}

// ValidateBand3 validates the multiplier band
func ValidateBand3(color Color) error {
	info := GetColorInfo(color)
	if !info.ValidMult {
		return &ValidationError{
			BandNumber: 3,
			Message:    fmt.Sprintf("%s is not valid for multiplier band", info.Name),
		}
	}
	return nil
}

// ValidateBand4 validates the tolerance band
func ValidateBand4(color Color, capacitancePF float64) error {
	_, exists := GetToleranceInfo(color)
	if !exists {
		info := GetColorInfo(color)
		return &ValidationError{
			BandNumber: 4,
			Message:    fmt.Sprintf("%s is not valid for tolerance band", info.Name),
		}
	}

	// Additional validation for small capacitors (<=10pF)
	tolInfo, _ := GetToleranceInfo(color)
	if capacitancePF <= 10.0 && tolInfo.AbsolutePF == 0 {
		info := GetColorInfo(color)
		return &ValidationError{
			BandNumber: 4,
			Message: fmt.Sprintf("%s tolerance not typically used for capacitors ≤10pF (use Brown, Red, Green, or White for absolute tolerance)", info.Name),
		}
	}

	return nil
}

// ValidateBand5 validates the voltage/temp coefficient band
func ValidateBand5(color Color, capType CapacitorType, bandCount int) error {
	if bandCount < 4 {
		return nil // No band 5 for 3-band capacitors
	}

	info := GetColorInfo(color)

	// For voltage rating (4-band and 5-band), validate against type
	if bandCount >= 4 {
		_, valid := GetVoltageRatingFractional(capType, color)
		if !valid {
			typeInfo, _ := GetTypeInfo(capType)
			return &ValidationError{
				BandNumber: 5,
				Message:    fmt.Sprintf("%s is not a valid voltage code for %s capacitors", info.Name, typeInfo.Description),
			}
		}
	}

	// For 5-band, also check if temp coefficient exists
	if bandCount == 5 {
		_, hasTempCoeff := GetTempCoefficient(color)
		if !hasTempCoeff {
			// This is just a warning - some colors don't have temp coefficients
			// but still have voltage ratings
		}
	}

	return nil
}

// ValidateReading validates an entire capacitor reading
func ValidateReading(reading *CapacitorReading) error {
	// Validate band 1
	if err := ValidateBand1(reading.Band1); err != nil {
		return err
	}

	// Validate band 2
	if err := ValidateBand2(reading.Band2); err != nil {
		return err
	}

	// Validate band 3
	if err := ValidateBand3(reading.Band3); err != nil {
		return err
	}

	// Calculate capacitance for band 4 validation
	info1 := GetColorInfo(reading.Band1)
	info2 := GetColorInfo(reading.Band2)
	info3 := GetColorInfo(reading.Band3)
	capacitancePF := float64(info1.Digit*10+info2.Digit) * info3.Multiplier

	// Validate band 4
	if err := ValidateBand4(reading.Band4, capacitancePF); err != nil {
		return err
	}

	// Validate band 5 (if applicable)
	if reading.BandCount >= 4 {
		if err := ValidateBand5(reading.Band5, reading.CapType, reading.BandCount); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBandCount validates the band count selection
func ValidateBandCount(count int) error {
	if count < 3 || count > 5 {
		return fmt.Errorf("band count must be 3, 4, or 5")
	}
	return nil
}

// ValidateCapacitorType validates the capacitor type selection
func ValidateCapacitorType(capType CapacitorType) error {
	_, exists := GetTypeInfo(capType)
	if !exists {
		return fmt.Errorf("invalid capacitor type (must be J, K, L, M, or N)")
	}
	return nil
}

// GetBandName returns a human-readable name for each band
func GetBandName(bandNum int) string {
	switch bandNum {
	case 1:
		return "First Digit"
	case 2:
		return "Second Digit"
	case 3:
		return "Multiplier"
	case 4:
		return "Tolerance"
	case 5:
		return "Voltage / Temp Coeff"
	default:
		return fmt.Sprintf("Band %d", bandNum)
	}
}

// GetBandDescription returns a detailed description for each band
func GetBandDescription(bandNum int, bandCount int) string {
	switch bandNum {
	case 1:
		return "First significant digit (0-9)"
	case 2:
		return "Second significant digit (0-9)"
	case 3:
		return "Multiplier (power of 10)"
	case 4:
		return "Tolerance (±%)"
	case 5:
		if bandCount == 5 {
			return "Voltage rating and/or temperature coefficient"
		}
		return "Voltage rating"
	default:
		return ""
	}
}

// Resistor validation functions

// ValidateResistorBand1 validates the first digit band for resistors
func ValidateResistorBand1(color Color) error {
	info := GetColorInfo(color)
	if !info.ValidDigit {
		return &ValidationError{
			BandNumber: 1,
			Message:    fmt.Sprintf("%s is not valid for first digit band (must be Black-White, 0-9)", info.Name),
		}
	}
	return nil
}

// ValidateResistorBand2 validates the second digit band for resistors
func ValidateResistorBand2(color Color) error {
	info := GetColorInfo(color)
	if !info.ValidDigit {
		return &ValidationError{
			BandNumber: 2,
			Message:    fmt.Sprintf("%s is not valid for second digit band (must be Black-White, 0-9)", info.Name),
		}
	}
	return nil
}

// ValidateResistorBand3 validates the third digit band for resistors (5/6-band only)
func ValidateResistorBand3(color Color) error {
	info := GetColorInfo(color)
	if !info.ValidDigit {
		return &ValidationError{
			BandNumber: 3,
			Message:    fmt.Sprintf("%s is not valid for third digit band (must be Black-White, 0-9)", info.Name),
		}
	}
	return nil
}

// ValidateResistorMultiplier validates the multiplier band for resistors
func ValidateResistorMultiplier(color Color, bandNum int) error {
	_, valid := GetResistorMultiplier(color)
	if !valid {
		info := GetColorInfo(color)
		return &ValidationError{
			BandNumber: bandNum,
			Message:    fmt.Sprintf("%s is not valid for multiplier band", info.Name),
		}
	}
	return nil
}

// ValidateResistorTolerance validates the tolerance band for resistors
func ValidateResistorTolerance(color Color, bandNum int) error {
	_, exists := GetResistorTolerance(color)
	if !exists {
		info := GetColorInfo(color)
		return &ValidationError{
			BandNumber: bandNum,
			Message:    fmt.Sprintf("%s is not valid for tolerance band", info.Name),
		}
	}
	return nil
}

// ValidateResistorTempCoeff validates the temperature coefficient band (6-band only)
func ValidateResistorTempCoeff(color Color) error {
	_, exists := GetResistorTempCoefficient(color)
	if !exists {
		info := GetColorInfo(color)
		return &ValidationError{
			BandNumber: 6,
			Message:    fmt.Sprintf("%s is not valid for temperature coefficient band", info.Name),
		}
	}
	return nil
}

// ValidateResistorReading validates an entire resistor reading
func ValidateResistorReading(reading *ResistorReading) error {
	switch reading.BandCount {
	case 4:
		// 4-band: Band1 Band2 Multiplier Tolerance
		if err := ValidateResistorBand1(reading.Band1); err != nil {
			return err
		}
		if err := ValidateResistorBand2(reading.Band2); err != nil {
			return err
		}
		if err := ValidateResistorMultiplier(reading.Band3, 3); err != nil {
			return err
		}
		if err := ValidateResistorTolerance(reading.Band4, 4); err != nil {
			return err
		}

	case 5:
		// 5-band: Band1 Band2 Band3 Multiplier Tolerance
		if err := ValidateResistorBand1(reading.Band1); err != nil {
			return err
		}
		if err := ValidateResistorBand2(reading.Band2); err != nil {
			return err
		}
		if err := ValidateResistorBand3(reading.Band3); err != nil {
			return err
		}
		if err := ValidateResistorMultiplier(reading.Band4, 4); err != nil {
			return err
		}
		if err := ValidateResistorTolerance(reading.Band5, 5); err != nil {
			return err
		}

	case 6:
		// 6-band: Band1 Band2 Band3 Multiplier Tolerance TempCoeff
		if err := ValidateResistorBand1(reading.Band1); err != nil {
			return err
		}
		if err := ValidateResistorBand2(reading.Band2); err != nil {
			return err
		}
		if err := ValidateResistorBand3(reading.Band3); err != nil {
			return err
		}
		if err := ValidateResistorMultiplier(reading.Band4, 4); err != nil {
			return err
		}
		if err := ValidateResistorTolerance(reading.Band5, 5); err != nil {
			return err
		}
		if err := ValidateResistorTempCoeff(reading.Band6); err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid band count: %d (must be 4, 5, or 6)", reading.BandCount)
	}

	return nil
}

// ValidateResistorBandCount validates the resistor band count selection
func ValidateResistorBandCount(count int) error {
	if count < 4 || count > 6 {
		return fmt.Errorf("resistor band count must be 4, 5, or 6")
	}
	return nil
}
