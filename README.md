# Tropical Fish Capacitor Color Code Decoder

A beautiful, interactive terminal-based application for decoding "Tropical Fish" capacitor color codes according to the IEC 60062 standard.

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Go](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

## Overview

Tropical Fish Capacitor Decoder is a lightweight TUI (Terminal User Interface) application that helps you decode the color bands on capacitors to determine their:
- **Capacitance value** (automatically scaled to pF, nF, µF, or mF)
- **Tolerance** (percentage-based or absolute pF for small capacitors)
- **Voltage rating** (based on capacitor type)
- **Temperature coefficient** (when applicable)

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for an engaging, colorful terminal experience!

## Features

✨ **Rich Terminal UI**
- Colorful, intuitive interface with ANSI color support
- Real-time input validation with helpful error messages
- Visual color band representations

🎯 **Comprehensive Support**
- 3, 4, and 5-band capacitor configurations
- 5 capacitor types: J (Tantalum), K (Mica), L (Polyester/Polystyrene), M (Electrolytic 4-band), N (Electrolytic 3-band)
- 12 standard color codes (Black through White, plus Gold and Silver)

🧮 **Accurate Calculations**
- IEC 60062 compliant calculations
- Automatic unit scaling (pF → nF → µF → mF)
- Dual tolerance modes: percentage (>10pF) and absolute (≤10pF)
- Type-specific voltage rating lookup
- Temperature coefficient display

⚡ **User-Friendly**
- Sequential band-by-band input with confirmation
- Edit mode to correct individual bands
- Review screen before calculation
- Decode multiple capacitors in one session

## Installation

### From Source

```bash
# Clone the repository
git clone <repository-url>
cd tropical-fish

# Build the application
go build -o tropical-fish

# Run it
./tropical-fish
```

### Pre-built Binaries

Download the latest release for your platform from the [Releases](releases) page:
- Linux (x86_64, ARM64)
- macOS (Intel, Apple Silicon)
- Windows (x86_64)

## Usage

### Quick Start

1. Launch the application:
   ```bash
   ./tropical-fish
   ```

2. Follow the on-screen prompts:
   - Select capacitor type (J, K, L, M, or N)
   - Select band count (3, 4, or 5)
   - Enter each color band sequentially
   - Review your input
   - View the calculated results

### Example Walkthrough

Decoding a **27 nF mica capacitor** with 5 bands:

```
Step 1: Capacitor Type
→ Enter: K (for Mica)

Step 2: Band Count
→ Enter: 5

Step 3: Band Input
→ Band 1 (First Digit): red
→ Band 2 (Second Digit): violet
→ Band 3 (Multiplier): orange
→ Band 4 (Tolerance): brown
→ Band 5 (Voltage): orange

Result:
  Capacitance: 27 nF (27,000 pF)
  Tolerance: ±1% (26.73 nF ──► 27.27 nF)
  Voltage: 400 V (Type K Mica)
  Temperature Coefficient: -150 × 10⁻⁶ /°C
```

## Capacitor Types

| Type | Description | Voltage Options |
|------|-------------|-----------------|
| **J** | Dipped Tantalum | 3V, 4V, 6V, 10V, 15V, 20V, 25V, 35V, 50V |
| **K** | Mica | 100V - 2000V (11 options) |
| **L** | Polyester / Polystyrene | 100V, 250V, 400V, 630V |
| **M** | Electrolytic (4-band) | 1.6V, 2.5V, 4V, 6.3V, 10V, 16V, 25V, 40V |
| **N** | Electrolytic (3-band) | 3V, 6V, 6.3V, 10V, 15V, 20V, 25V, 35V |

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

- **Enter** - Confirm input / Proceed to next step
- **Backspace/Delete** - Delete last character
- **C** - Correct/Edit a band (on review screen)
- **D** - Decode another capacitor (on results screen)
- **E** - Edit current capacitor (on results screen)
- **Q** - Quit application
- **Ctrl+C** - Quit immediately

## Development

### Building from Source

Requirements:
- Go 1.21 or later

```bash
# Clone and enter directory
git clone <repository-url>
cd tropical-fish

# Install dependencies
go mod tidy

# Build
CGO_ENABLED=0 go build -o tropical-fish

# Run tests
go test -v
```

### Project Structure

```
tropical-fish/
├── main.go           # TUI application and screen rendering
├── colors.go         # Color code definitions and mappings
├── types.go          # Capacitor type definitions
├── calculator.go     # Calculation engine
├── validator.go      # Input validation logic
├── styles.go         # Lipgloss styling and visual themes
├── calculator_test.go # Unit tests
└── README.md         # This file
```

### Building Cross-Platform Binaries

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o tropical-fish-linux-amd64

# Linux ARM64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o tropical-fish-linux-arm64

# macOS Intel
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o tropical-fish-darwin-amd64

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o tropical-fish-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o tropical-fish-windows-amd64.exe
```

## Testing

The project includes comprehensive unit tests covering:
- Capacitance calculations
- Tolerance calculations (percentage and absolute)
- Unit auto-scaling
- Voltage rating lookups
- Color parsing
- Input validation

Run tests:
```bash
go test -v
```

## Technical Details

### Calculation Formula

```
Capacitance (pF) = (Band1 × 10 + Band2) × Multiplier
```

Example: Red (2) + Violet (7) + Orange (×1,000) = 27 × 1,000 = 27,000 pF = 27 nF

### Unit Auto-Scaling

The application automatically selects the most readable unit:
- **pF** (picofarad): < 1,000 pF
- **nF** (nanofarad): 1,000 pF - 999,999 pF
- **µF** (microfarad): 1,000,000 pF - 999,999,999 pF
- **mF** (millifarad): ≥ 1,000,000,000 pF

### Tolerance Logic

- **Capacitance > 10 pF**: Percentage-based tolerance (±%)
- **Capacitance ≤ 10 pF**: Absolute tolerance (±pF)

This follows the IEC 60062 standard for precision capacitors.

## Standards Compliance

This application implements:
- **IEC 60062** - Color coding for capacitors
- Standard "Tropical Fish" 5-band capacitor format
- Type-specific voltage rating tables (J, K, L, M, N)

## Troubleshooting

### Colors Not Displaying Properly

Ensure your terminal supports:
- ANSI escape codes
- 24-bit true color (or at least 256 colors)

Recommended terminals:
- **Linux**: GNOME Terminal, Konsole, Alacritty, kitty
- **macOS**: iTerm2, Terminal.app (macOS 10.14+), Alacritty
- **Windows**: Windows Terminal, ConEmu, Alacritty

### Invalid Color Error

The application accepts color names case-insensitively:
- ✅ `red`, `Red`, `RED` (all valid)
- ✅ `grey` or `gray` (both spellings accepted)
- ❌ `rd`, `r` (invalid - use full color name)

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is licensed under the MIT License.

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm
- Styled with [Lipgloss](https://github.com/charmbracelet/lipgloss)
- Based on IEC 60062 standard specifications

## Support

For bugs, questions, or feature requests, please open an issue on GitHub.

---

**Made with ❤️ for electronics enthusiasts and engineers**
