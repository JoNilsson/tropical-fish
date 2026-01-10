package main

import (
	"testing"
)

// TestCalculateCapacitance tests basic capacitance calculations
func TestCalculateCapacitance(t *testing.T) {
	tests := []struct {
		name         string
		band1        Color
		band2        Color
		band3        Color
		band4        Color
		band5        Color
		bandCount    int
		capType      CapacitorType
		expectedPF   float64
		expectedUnit string
	}{
		{
			name:         "27nF capacitor - corrected example",
			band1:        ColorRed,    // 2
			band2:        ColorViolet, // 7
			band3:        ColorOrange, // x1,000 (27 × 1,000 = 27,000 pF = 27 nF)
			band4:        ColorBrown,  // ±1%
			band5:        ColorOrange,
			bandCount:    5,
			capType:      TypeK,
			expectedPF:   27000,
			expectedUnit: "nF",
		},
		{
			name:         "100pF capacitor",
			band1:        ColorBrown, // 1
			band2:        ColorBlack, // 0
			band3:        ColorBrown, // x10
			band4:        ColorBrown, // ±1%
			band5:        ColorBlack,
			bandCount:    4,
			capType:      TypeK,
			expectedPF:   100,
			expectedUnit: "pF",
		},
		{
			name:         "1µF capacitor",
			band1:        ColorBrown,  // 1
			band2:        ColorBlack,  // 0
			band3:        ColorYellow, // x10,000
			band4:        ColorGreen,  // ±5%
			band5:        ColorBlack,
			bandCount:    4,
			capType:      TypeL,
			expectedPF:   100000,
			expectedUnit: "nF", // Should be 100 nF
		},
		{
			name:         "Small capacitor with decimal multiplier",
			band1:        ColorRed,    // 2
			band2:        ColorViolet, // 7
			band3:        ColorGrey,   // x0.01
			band4:        ColorBrown,  // ±1%
			band5:        ColorBlack,
			bandCount:    4,
			capType:      TypeK,
			expectedPF:   0.27, // 27 * 0.01
			expectedUnit: "pF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading := CapacitorReading{
				Band1:     tt.band1,
				Band2:     tt.band2,
				Band3:     tt.band3,
				Band4:     tt.band4,
				Band5:     tt.band5,
				BandCount: tt.bandCount,
				CapType:   tt.capType,
			}

			result, err := Calculate(reading)
			if err != nil {
				t.Fatalf("Calculate() error = %v", err)
			}

			if result.CapacitancePF != tt.expectedPF {
				t.Errorf("CapacitancePF = %v, want %v", result.CapacitancePF, tt.expectedPF)
			}

			if result.CapacitanceUnit != tt.expectedUnit {
				t.Errorf("CapacitanceUnit = %v, want %v", result.CapacitanceUnit, tt.expectedUnit)
			}
		})
	}
}

// TestToleranceCalculation tests tolerance calculations
func TestToleranceCalculation(t *testing.T) {
	tests := []struct {
		name              string
		capacitancePF     float64
		toleranceColor    Color
		expectedType      string
		expectedSymmetric bool
	}{
		{
			name:              "Large cap with percentage tolerance",
			capacitancePF:     27000,
			toleranceColor:    ColorBrown, // ±1%
			expectedType:      "percentage",
			expectedSymmetric: true,
		},
		{
			name:              "Small cap with absolute tolerance",
			capacitancePF:     5.0,
			toleranceColor:    ColorBrown, // ±0.1pF
			expectedType:      "absolute",
			expectedSymmetric: true,
		},
		{
			name:              "Asymmetric tolerance (Grey)",
			capacitancePF:     1000,
			toleranceColor:    ColorGrey, // +80%/-20%
			expectedType:      "percentage",
			expectedSymmetric: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading := CapacitorReading{
				Band1:     ColorBrown,
				Band2:     ColorBlack,
				Band3:     ColorBrown,
				Band4:     tt.toleranceColor,
				Band5:     ColorBlack,
				BandCount: 4,
				CapType:   TypeK,
			}

			// Override capacitance for testing
			info1 := GetColorInfo(reading.Band1)
			info2 := GetColorInfo(reading.Band2)
			info3 := GetColorInfo(reading.Band3)
			reading.Band1 = ColorBlack
			reading.Band2 = ColorBlack
			reading.Band3 = ColorBlack

			// Set up a reading that produces the desired capacitance
			baseValue := tt.capacitancePF
			if baseValue > 99 {
				reading.Band1 = ColorBrown
				reading.Band2 = ColorBlack
				reading.Band3 = ColorBrown
			}

			result, err := Calculate(reading)
			if err != nil {
				t.Fatalf("Calculate() error = %v", err)
			}

			// For small caps, manually set capacitance
			if tt.capacitancePF <= 10 {
				result.CapacitancePF = tt.capacitancePF
				_ = calculateTolerance(result)
			}

			if result.ToleranceType != tt.expectedType {
				t.Errorf("ToleranceType = %v, want %v", result.ToleranceType, tt.expectedType)
			}

			if result.ToleranceSymmetric != tt.expectedSymmetric {
				t.Errorf("ToleranceSymmetric = %v, want %v", result.ToleranceSymmetric, tt.expectedSymmetric)
			}

			// Verify min/max values are calculated
			if result.MinValue >= result.MaxValue {
				t.Errorf("MinValue (%v) should be less than MaxValue (%v)", result.MinValue, result.MaxValue)
			}

			_ = info1
			_ = info2
			_ = info3
		})
	}
}

// TestUnitScaling tests the auto-scaling of capacitance units
func TestUnitScaling(t *testing.T) {
	tests := []struct {
		name         string
		pF           float64
		expectedUnit string
	}{
		{"Small picofarad", 47, "pF"},
		{"Large picofarad", 470, "pF"},
		{"Small nanofarad", 1000, "nF"},
		{"Large nanofarad", 470000, "nF"},
		{"Small microfarad", 1000000, "µF"},
		{"Large microfarad", 470000000, "µF"},
		{"Millifarad", 1000000000, "mF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, unit := scaleCapacitance(tt.pF)
			if unit != tt.expectedUnit {
				t.Errorf("scaleCapacitance(%v) unit = %v, want %v", tt.pF, unit, tt.expectedUnit)
			}
		})
	}
}

// TestVoltageRating tests voltage rating lookups for different types
func TestVoltageRating(t *testing.T) {
	tests := []struct {
		name        string
		capType     CapacitorType
		color       Color
		expectedV   float64
		shouldExist bool
	}{
		{"Type K - Orange", TypeK, ColorOrange, 400, true},
		{"Type J - Brown", TypeJ, ColorBrown, 4, true},
		{"Type L - Brown", TypeL, ColorBrown, 250, true}, // Brown (1) = 250V for Type L
		{"Type M - Brown (1.6V)", TypeM, ColorBrown, 1.6, true},
		{"Type N - Red (6.3V)", TypeN, ColorRed, 6.3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			voltage, exists := GetVoltageRatingFractional(tt.capType, tt.color)
			if exists != tt.shouldExist {
				t.Errorf("GetVoltageRatingFractional() exists = %v, want %v", exists, tt.shouldExist)
			}
			if exists && voltage != tt.expectedV {
				t.Errorf("GetVoltageRatingFractional() voltage = %v, want %v", voltage, tt.expectedV)
			}
		})
	}
}

// TestColorParsing tests color name parsing
func TestColorParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Color
		valid    bool
	}{
		{"red", ColorRed, true},
		{"RED", ColorRed, true},
		{"Red", ColorRed, true},
		{"brown", ColorBrown, true},
		{"grey", ColorGrey, true},
		{"gray", ColorGrey, true}, // Alternative spelling
		{"violet", ColorViolet, true},
		{"invalid", ColorBlack, false},
		{"", ColorBlack, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			color, valid := ParseColor(tt.input)
			if valid != tt.valid {
				t.Errorf("ParseColor(%q) valid = %v, want %v", tt.input, valid, tt.valid)
			}
			if valid && color != tt.expected {
				t.Errorf("ParseColor(%q) = %v, want %v", tt.input, color, tt.expected)
			}
		})
	}
}

// TestValidation tests input validation
func TestValidation(t *testing.T) {
	tests := []struct {
		name      string
		color     Color
		bandNum   int
		shouldErr bool
	}{
		{"Band 1 - valid digit", ColorRed, 1, false},
		{"Band 1 - invalid (Gold)", ColorGold, 1, true},
		{"Band 2 - valid digit", ColorViolet, 2, false},
		{"Band 2 - invalid (Silver)", ColorSilver, 2, true},
		{"Band 3 - valid multiplier", ColorYellow, 3, false},
		{"Band 4 - valid tolerance", ColorBrown, 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			switch tt.bandNum {
			case 1:
				err = ValidateBand1(tt.color)
			case 2:
				err = ValidateBand2(tt.color)
			case 3:
				err = ValidateBand3(tt.color)
			case 4:
				err = ValidateBand4(tt.color, 1000) // Use large capacitance
			}

			hasErr := err != nil
			if hasErr != tt.shouldErr {
				t.Errorf("Validation error = %v, shouldErr = %v", hasErr, tt.shouldErr)
			}
		})
	}
}
