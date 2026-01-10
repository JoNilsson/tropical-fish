# Resistor Decoding Feature - Scope Document

## Overview

Expand the Tropical Fish decoder to support resistor color code decoding in addition to the existing capacitor functionality. This will make it a comprehensive component decoder tool.

---

## Resistor Color Code Standards

### 4-Band Resistors (Standard)
**Most common type**
- **Band 1**: First significant digit (0-9)
- **Band 2**: Second significant digit (0-9)
- **Band 3**: Multiplier (×1, ×10, ×100, etc.)
- **Band 4**: Tolerance (±%)

**Formula**: Resistance (Ω) = (Band1 × 10 + Band2) × Multiplier

### 5-Band Resistors (Precision)
**Higher precision resistors**
- **Band 1**: First significant digit (0-9)
- **Band 2**: Second significant digit (0-9)
- **Band 3**: Third significant digit (0-9)
- **Band 4**: Multiplier (×1, ×10, ×100, etc.)
- **Band 5**: Tolerance (±%)

**Formula**: Resistance (Ω) = (Band1 × 100 + Band2 × 10 + Band3) × Multiplier

### 6-Band Resistors (Precision + Temperature Coefficient)
**High-precision resistors with temp coefficient**
- **Band 1**: First significant digit (0-9)
- **Band 2**: Second significant digit (0-9)
- **Band 3**: Third significant digit (0-9)
- **Band 4**: Multiplier (×1, ×10, ×100, etc.)
- **Band 5**: Tolerance (±%)
- **Band 6**: Temperature coefficient (ppm/°C)

---

## Color Code Mappings for Resistors

### Band 1-3: Digit Values (0-9)
| Color  | Value |
|--------|-------|
| Black  | 0     |
| Brown  | 1     |
| Red    | 2     |
| Orange | 3     |
| Yellow | 4     |
| Green  | 5     |
| Blue   | 6     |
| Violet | 7     |
| Grey   | 8     |
| White  | 9     |

### Multiplier Band
| Color  | Multiplier | Description |
|--------|------------|-------------|
| Black  | ×1         | 1 Ω         |
| Brown  | ×10        | 10 Ω        |
| Red    | ×100       | 100 Ω       |
| Orange | ×1,000     | 1 kΩ        |
| Yellow | ×10,000    | 10 kΩ       |
| Green  | ×100,000   | 100 kΩ      |
| Blue   | ×1,000,000 | 1 MΩ        |
| Violet | ×10,000,000| 10 MΩ       |
| Grey   | ×100,000,000| 100 MΩ     |
| White  | ×1,000,000,000| 1 GΩ     |
| Gold   | ×0.1       | 0.1 Ω       |
| Silver | ×0.01      | 0.01 Ω      |

### Tolerance Band
| Color  | Tolerance | Common Use        |
|--------|-----------|-------------------|
| Brown  | ±1%       | Precision         |
| Red    | ±2%       | Precision         |
| Green  | ±0.5%     | High precision    |
| Blue   | ±0.25%    | High precision    |
| Violet | ±0.1%     | Ultra precision   |
| Grey   | ±0.05%    | Ultra precision   |
| Gold   | ±5%       | Standard          |
| Silver | ±10%      | Standard          |
| None   | ±20%      | General purpose   |

### Temperature Coefficient (6-band only)
| Color  | Coefficient (ppm/°C) |
|--------|----------------------|
| Black  | 250                  |
| Brown  | 100                  |
| Red    | 50                   |
| Orange | 15                   |
| Yellow | 25                   |
| Green  | 20                   |
| Blue   | 10                   |
| Violet | 5                    |
| Grey   | 1                    |

---

## Unit Auto-Scaling

Resistance values should be automatically scaled to the most readable unit:

- **Ω (Ohms)**: < 1,000 Ω
- **kΩ (Kilohms)**: 1,000 Ω - 999,999 Ω
- **MΩ (Megohms)**: 1,000,000 Ω - 999,999,999 Ω
- **GΩ (Gigohms)**: ≥ 1,000,000,000 Ω

Examples:
- 470 Ω → "470 Ω"
- 4,700 Ω → "4.7 kΩ"
- 47,000 Ω → "47 kΩ"
- 4,700,000 Ω → "4.7 MΩ"

---

## UI/UX Flow Changes

### 1. Component Type Selection (NEW SCREEN)
```
═══════════════════════════════════════════════════════════════
        TROPICAL FISH COMPONENT COLOR CODE DECODER
                    Welcome Screen
═══════════════════════════════════════════════════════════════

What would you like to decode?

  (C) Capacitor - IEC 60062 Standard (3/4/5 bands)
  (R) Resistor - EIA Standard (4/5/6 bands)

Select component type (C/R): _
```

### 2. Resistor Band Count Selection
```
────────────────────────────────────────────────────────────────
You selected: Resistor

How many color bands does your resistor have?
  4 = 4-band (standard, ±5% or ±10% tolerance)
  5 = 5-band (precision, ±1% or ±2% tolerance)
  6 = 6-band (precision + temperature coefficient)

Select band count (4/5/6): _
```

### 3. Resistor Band Input
Similar to capacitor flow, but with:
- No capacitor type selection
- Different band descriptions
- Different validation rules

### 4. Resistor Results Display
```
════════════════════════════════════════════════════════════════
                        RESULTS
════════════════════════════════════════════════════════════════

Component Type:     Resistor
Configuration:      5-band (precision)

RESISTANCE VALUE:
  Value:            4.7 kΩ [0.0047 MΩ]

TOLERANCE:
  Specification:    ±1% (precision)
  Range:            4.653 kΩ ──► 4.747 kΩ

POWER RATING:
  Typical:          1/4 W (based on standard 4.7kΩ resistor)
  Note:             Check physical size for actual rating

TEMPERATURE COEFFICIENT: (6-band only)
  Coefficient:      100 ppm/°C

════════════════════════════════════════════════════════════════

(D)ecode  |  (E)dit  |  (N)ote  |  e(X)port  |  (Q)uit
```

---

## Technical Implementation

### File Structure
```
tropical-fish/
├── main.go              # Updated: Add component type selection
├── colors.go            # Reused: Same color codes
├── types.go             # Updated: Add ComponentType enum
├── calculator.go        # Capacitor calculations
├── resistor.go          # NEW: Resistor calculations
├── validator.go         # Updated: Add resistor validation
├── export.go            # Updated: Handle both types
├── autocomplete.go      # Reused: Same color autocomplete
├── styles.go            # Reused: Same styles
└── README.md            # Updated: Document resistor feature
```

### Data Structures

```go
// ComponentType distinguishes between capacitors and resistors
type ComponentType int

const (
    ComponentCapacitor ComponentType = iota
    ComponentResistor
)

// ResistorReading represents parsed resistor bands
type ResistorReading struct {
    Band1     Color  // First digit
    Band2     Color  // Second digit
    Band3     Color  // Third digit (5/6-band only)
    Band4     Color  // Multiplier (4-band) or Third digit (5/6-band)
    Band5     Color  // Tolerance (4-band) or Multiplier (5/6-band)
    Band6     Color  // Temperature coefficient (6-band only)
    BandCount int    // 4, 5, or 6
}

// ResistorResult contains calculated resistor values
type ResistorResult struct {
    ResistanceOhms    float64 // Raw value in ohms
    ResistanceValue   float64 // Scaled value
    ResistanceUnit    string  // Ω, kΩ, MΩ, or GΩ

    TolerancePercent  float64 // Tolerance in %
    MinValue          float64 // Min resistance
    MaxValue          float64 // Max resistance
    MinUnit           string
    MaxUnit           string

    TempCoefficient   int     // ppm/°C (6-band only)
    TempCoeffValid    bool

    Reading           ResistorReading
}
```

### Validation Rules

**4-band Resistors:**
- Band 1-2: Must be valid digit colors (Black-White)
- Band 3: Must be valid multiplier (Black-Silver)
- Band 4: Must be valid tolerance color

**5-band Resistors:**
- Band 1-3: Must be valid digit colors (Black-White)
- Band 4: Must be valid multiplier (Black-Silver)
- Band 5: Must be valid tolerance color

**6-band Resistors:**
- Band 1-3: Must be valid digit colors (Black-White)
- Band 4: Must be valid multiplier (Black-Silver)
- Band 5: Must be valid tolerance color
- Band 6: Must be valid temp coefficient color

---

## CSV Export Updates

### New CSV Structure
Add a "Component Type" column and handle both resistors and capacitors:

```csv
Timestamp,Component Type,Band Count,Band 1,Band 2,Band 3,Band 4,Band 5,Band 6,Value,Unit,Tolerance (%),Min Value,Max Value,Temp Coefficient,Note
2026-01-07 11:30:00,Resistor,4,Yellow,Violet,,Black,Gold,,470,Ω,5,446.5 Ω,493.5 Ω,,Test resistor
2026-01-07 11:31:00,Resistor,5,Brown,Black,Black,Red,Brown,,10,kΩ,1,9.9 kΩ,10.1 kΩ,,Pull-up resistor
2026-01-07 11:32:00,Capacitor,5,Red,Violet,Orange,Brown,Orange,,27,nF,1,26.73 nF,27.27 nF,,-150×10⁻⁶,Decoupling cap
```

---

## Testing Requirements

### Unit Tests
1. **4-band resistor calculations**
   - Standard values (470Ω, 1kΩ, 10kΩ, etc.)
   - Decimal multipliers (Gold, Silver)
   - Different tolerances

2. **5-band resistor calculations**
   - Precision values (1.00kΩ, 4.75kΩ, etc.)
   - High precision tolerances

3. **6-band resistor calculations**
   - With temperature coefficients
   - Edge cases

4. **Unit auto-scaling**
   - Ω → kΩ → MΩ → GΩ transitions

5. **Validation**
   - Invalid color combinations
   - Band count restrictions

### Integration Tests
1. Full decode flow for each resistor type
2. CSV export with mixed capacitors and resistors
3. Autocomplete works for resistor bands
4. Edit mode for resistor readings

---

## Examples

### Example 1: Standard 4-band Resistor
**Input**: Yellow, Violet, Black, Gold
**Calculation**: (4 × 10 + 7) × 1 = 47 Ω
**Tolerance**: ±5%
**Range**: 44.65 Ω ──► 49.35 Ω

### Example 2: Precision 5-band Resistor
**Input**: Brown, Black, Black, Red, Brown
**Calculation**: (1 × 100 + 0 × 10 + 0) × 100 = 10,000 Ω = 10 kΩ
**Tolerance**: ±1%
**Range**: 9.9 kΩ ──► 10.1 kΩ

### Example 3: 6-band with Temp Coefficient
**Input**: Red, Red, Black, Orange, Red, Brown
**Calculation**: (2 × 100 + 2 × 10 + 0) × 1,000 = 220,000 Ω = 220 kΩ
**Tolerance**: ±2%
**Range**: 215.6 kΩ ──► 224.4 kΩ
**Temp Coefficient**: 100 ppm/°C

---

## Implementation Phases

### Phase 1: Core Resistor Engine
- [ ] Create `resistor.go` with calculation functions
- [ ] Implement resistance unit auto-scaling (Ω/kΩ/MΩ/GΩ)
- [ ] Add resistor tolerance calculations
- [ ] Temperature coefficient lookup (6-band)

### Phase 2: UI Integration
- [ ] Add component type selection screen
- [ ] Update band count selection for resistors (4/5/6)
- [ ] Create resistor band input flow
- [ ] Design resistor results display

### Phase 3: Data Management
- [ ] Update `ComponentEntry` to handle both types
- [ ] Modify CSV export for dual-component support
- [ ] Add resistor-specific validation rules
- [ ] Update history tracking

### Phase 4: Testing & Documentation
- [ ] Write comprehensive unit tests
- [ ] Test all resistor band combinations
- [ ] Update README with resistor documentation
- [ ] Add resistor examples to README

### Phase 5: Polish
- [ ] Add resistor color reference table to welcome screen
- [ ] Consider adding power rating hints (based on size)
- [ ] E-series value detection (E12, E24, E96, E192)
- [ ] Reverse lookup: "Find colors for 4.7kΩ"

---

## Future Enhancements (Out of Scope for Initial Release)

1. **SMD Resistor Codes**: 3-digit and 4-digit SMD codes (e.g., "472" = 4.7kΩ)
2. **Inductor Codes**: Similar color coding system
3. **Power Rating**: Physical size-based power rating hints
4. **E-Series Detection**: Identify standard resistor series (E12, E24, E96, E192)
5. **Reverse Lookup**: Input desired resistance, get color codes
6. **Batch Import**: Import multiple components from CSV
7. **Visual Color Band Picker**: Click colors instead of typing

---

## Success Criteria

✅ Successfully decode 4, 5, and 6-band resistors
✅ Accurate resistance calculations with proper unit scaling
✅ Tolerance range calculations
✅ Temperature coefficient support (6-band)
✅ CSV export includes both resistors and capacitors
✅ All existing capacitor functionality remains intact
✅ Comprehensive test coverage (>90%)
✅ Updated documentation
✅ Clean component type selection UX
✅ Autocomplete works for resistor inputs

---

## Estimated Complexity

**Effort Level**: Medium-High

**Breakdown**:
- Core resistor logic: ~200 lines (similar to capacitor.go)
- UI updates: ~150 lines (new screens, modified flows)
- Validation: ~100 lines
- CSV export updates: ~50 lines
- Tests: ~300 lines
- Documentation: ~200 lines

**Total**: ~1,000 new/modified lines of code

**Time Estimate**: 3-4 hours for experienced Go developer

---

## Notes

- Resistors are simpler than capacitors (no voltage rating types)
- Can reuse most existing infrastructure (colors, autocomplete, styles)
- Main complexity is in managing two component types elegantly
- CSV export needs careful design to handle both types
- Consider renaming app to "Component Color Code Decoder" vs "Tropical Fish"
