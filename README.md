# Tropical Fish Color Code Decoder

Terminal-based utility for decoding resistor and capacitor color codes per IEC 60062 standard.

## Overview

Tropical Fish is a TUI application for decoding component color bands. It identifies capacitance values, resistance values, tolerance, voltage ratings, and temperature coefficients from standard color codes.

Supports:
- **Capacitors**: 3, 4, and 5-band configurations across 5 types (J, K, L, M, N)
- **Resistors**: 4, 5, and 6-band configurations with temperature coefficient data for 6-band types

Automatic unit scaling (pF→nF→µF→mF for caps, Ω→kΩ→MΩ→GΩ for resistors), tolerance calculations, and component-specific parameters.

## Installation

### Build from Source

```bash
git clone https://github.com/JoNilsson/tropical-fish
cd tropical-fish
go mod tidy
CGO_ENABLED=0 go build -o tropical-fish
./tropical-fish
```

Requires Go 1.21+

### Pre-built Binaries

Available on [Releases](https://github.com/JoNilsson/tropical-fish/releases):
- Linux x86_64, ARM64
- macOS Intel, Apple Silicon
- Windows x86_64

## Usage

Launch the application:
```bash
./tropical-fish
```

Select component type (capacitor or resistor), enter band count and colors sequentially. Review and confirm before calculation.

### Capacitor Example

5-band mica capacitor (27 nF, 1% tolerance, 400V):
- Type: K (Mica)
- Bands: red, violet, orange, brown, orange
- Result: 27 nF ±1% (26.73–27.27 nF), 400V

### Resistor Example

5-band resistor (4700 Ω, 1% tolerance):
- Bands: yellow, violet, black, brown, brown
- Result: 4.7 kΩ ±1% (4.653–4.747 kΩ)

## Component Support

### Capacitors

| Type | Description | Voltage Ratings |
|------|-------------|-----------------|
| J | Dipped Tantalum | 3V–50V |
| K | Mica | 100V–2000V |
| L | Polyester/Polystyrene | 100V–630V |
| M | Electrolytic (4-band) | 1.6V–40V |
| N | Electrolytic (3-band) | 3V–35V |

### Resistors

4-band: First digit, second digit, multiplier, tolerance
5-band: First digit, second digit, third digit, multiplier, tolerance
6-band: First digit, second digit, third digit, multiplier, tolerance, temperature coefficient

## Color Code Reference

### Digit Values (Bands 1-2)
| Color | Value |
|-------|-------|
| Black | 0 |
| Brown | 1 |
| Red | 2 |
| Orange | 3 |
| Yellow | 4 |
| Green | 5 |
| Blue | 6 |
| Violet | 7 |
| Grey | 8 |
| White | 9 |

### Multipliers (Band 3)
| Color | Multiplier |
|-------|------------|
| Black | ×1 |
| Brown | ×10 |
| Red | ×100 |
| Orange | ×1,000 |
| Yellow | ×10,000 |
| Green | ×100,000 |
| Blue | ×1,000,000 |
| Grey | ×0.01 |
| White | ×0.1 |
| Gold | ×0.1 |
| Silver | ×0.01 |

### Tolerance (Band 4)

**For Capacitance > 10pF (Percentage):**
| Color | Tolerance |
|-------|-----------|
| Black | ±20% |
| Brown | ±1% |
| Red | ±2% |
| Orange | ±3% |
| Yellow | ±4% |
| Green | ±5% |
| Gold | ±5% |
| White | ±10% |
| Silver | ±10% |
| Grey | +80% / -20% |

**For Capacitance ≤ 10pF (Absolute):**
| Color | Tolerance |
|-------|-----------|
| Brown | ±0.1 pF |
| Red | ±0.25 pF |
| Green | ±0.5 pF |
| White | ±1.0 pF |

## Keyboard Controls

| Key | Function |
|-----|----------|
| Enter | Confirm / Next step |
| Backspace | Delete character |
| C | Correct band |
| D | Decode another component |
| E | Edit component |
| Q | Quit |
| Ctrl+C | Force quit |

## Building

### Cross-Platform Binaries

```bash
# Linux x86_64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o tropical-fish-linux-amd64

# Linux ARM64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o tropical-fish-linux-arm64

# macOS Intel
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o tropical-fish-darwin-amd64

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o tropical-fish-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o tropical-fish.exe
```

### Testing

```bash
go test -v
```

Tests cover calculations, tolerance logic, unit scaling, color parsing, and validation.

## Calculations

### Capacitors

Base value from first 2 bands × multiplier. Tolerance calculated as percentage (>10pF) or absolute pF (≤10pF) per IEC 60062.

### Resistors

Base value from first 2 or 3 bands × multiplier. 6-band types include temperature coefficient (ppm/°C). Tolerance specified as ±%.

Resistance scales to Ω, kΩ, MΩ, or GΩ based on value.

## Resistor Tolerance

| Color | ±% |
|-------|-----|
| Brown | 1 |
| Red | 2 |
| Green | 0.5 |
| Blue | 0.25 |
| Violet | 0.1 |
| Grey | 0.05 |
| Gold | 5 |
| Silver | 10 |

## Resistor Temperature Coefficient (6-band)

| Color | ppm/°C |
|-------|--------|
| Black | 250 |
| Brown | 100 |
| Red | 50 |
| Orange | 15 |
| Yellow | 25 |
| Green | 20 |
| Blue | 10 |
| Violet | 5 |
| Grey | 1 |

## Standards

Implements IEC 60062 color coding for both capacitors and resistors.

## License

WTFPL (Do What The F*** You Want To Public License)
