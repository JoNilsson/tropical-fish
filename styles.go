package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	colorPrimary   = lipgloss.Color("#00D7FF") // Cyan
	colorSecondary = lipgloss.Color("#FFD700") // Gold
	colorSuccess   = lipgloss.Color("#00FF87") // Green
	colorWarning   = lipgloss.Color("#FFAF00") // Orange
	colorError     = lipgloss.Color("#FF5F87") // Pink/Red
	colorMuted     = lipgloss.Color("#6C7086") // Grey
	colorWhite     = lipgloss.Color("#FFFFFF")
	colorBorder    = lipgloss.Color("#5F87AF") // Blue-grey
)

// Base styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Italic(true)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorPrimary).
			Padding(0, 2).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.Border{
			Top:         "─",
			Bottom:      "",
			Left:        "",
			Right:       "",
			TopLeft:     "─",
			TopRight:    "─",
			BottomLeft:  "",
			BottomRight: "",
		}).
		BorderForeground(colorBorder).
		Padding(1, 0).
		MarginTop(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	promptStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(lipgloss.Color("#1C1C1C")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	// Band-specific styles
	bandHeaderStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			Border(lipgloss.Border{
			Top:    "─",
			Bottom: "─",
			Left:   "",
			Right:  "",
		}).
		BorderForeground(colorBorder).
		Padding(0, 1)

	bandLabelStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)

	confirmStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Padding(0, 1)

	// Results styles
	resultHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWhite).
				Background(colorSuccess).
				Padding(0, 2).
				Align(lipgloss.Center)

	resultLabelStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Bold(true).
				Width(20).
				Align(lipgloss.Right)

	resultValueStyle = lipgloss.NewStyle().
				Foreground(colorWhite).
				Bold(true)

	resultBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorSuccess).
			Padding(2, 3).
			MarginTop(1).
			MarginBottom(1)
)

// GetColorStyle returns a lipgloss style for a capacitor color
func GetColorStyle(color Color) lipgloss.Style {
	// Map capacitor colors to terminal colors
	var termColor lipgloss.Color
	switch color {
	case ColorBlack:
		termColor = lipgloss.Color("#000000")
	case ColorBrown:
		termColor = lipgloss.Color("#8B4513")
	case ColorRed:
		termColor = lipgloss.Color("#FF0000")
	case ColorOrange:
		termColor = lipgloss.Color("#FF8C00")
	case ColorYellow:
		termColor = lipgloss.Color("#FFFF00")
	case ColorGreen:
		termColor = lipgloss.Color("#00FF00")
	case ColorBlue:
		termColor = lipgloss.Color("#0000FF")
	case ColorViolet:
		termColor = lipgloss.Color("#9400D3")
	case ColorGrey:
		termColor = lipgloss.Color("#808080")
	case ColorWhite:
		termColor = lipgloss.Color("#FFFFFF")
	case ColorGold:
		termColor = lipgloss.Color("#FFD700")
	case ColorSilver:
		termColor = lipgloss.Color("#C0C0C0")
	default:
		termColor = lipgloss.Color("#FFFFFF")
	}

	// For light colors, use dark text; for dark colors, use light text
	textColor := lipgloss.Color("#000000")
	if color == ColorBlack || color == ColorBrown || color == ColorRed ||
		color == ColorBlue || color == ColorViolet || color == ColorGrey {
		textColor = lipgloss.Color("#FFFFFF")
	}

	return lipgloss.NewStyle().
		Foreground(textColor).
		Background(termColor).
		Bold(true).
		Padding(0, 2)
}

// RenderColorBand renders a color band with its name and value
func RenderColorBand(color Color, bandNum int) string {
	info := GetColorInfo(color)
	style := GetColorStyle(color)

	var value string
	switch bandNum {
	case 1, 2:
		value = info.Name + " (" + string(rune('0'+info.Digit)) + ")"
	case 3:
		if info.Multiplier >= 1 {
			value = info.Name + " (×" + formatMultiplier(info.Multiplier) + ")"
		} else {
			value = info.Name + " (×" + formatMultiplierDecimal(info.Multiplier) + ")"
		}
	case 4:
		tolInfo, _ := GetToleranceInfo(color)
		if tolInfo.Symmetric {
			value = info.Name + " (±" + formatFloat(tolInfo.PercentHigh) + "%)"
		} else {
			value = info.Name + " (+" + formatFloat(tolInfo.PercentHigh) + "% / -" + formatFloat(tolInfo.PercentLow) + "%)"
		}
	case 5:
		value = info.Name + " (voltage code)"
	}

	return style.Render(" " + value + " ")
}

// Helper functions for formatting
func formatMultiplier(m float64) string {
	if m == 1 {
		return "1"
	}
	if m == 10 {
		return "10"
	}
	if m == 100 {
		return "100"
	}
	if m == 1000 {
		return "1,000"
	}
	if m == 10000 {
		return "10,000"
	}
	if m == 100000 {
		return "100,000"
	}
	if m == 1000000 {
		return "1,000,000"
	}
	return "?"
}

func formatMultiplierDecimal(m float64) string {
	if m == 0.1 {
		return "0.1"
	}
	if m == 0.01 {
		return "0.01"
	}
	return "?"
}

func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return lipgloss.NewStyle().Render(string(rune('0' + int(f))))
	}
	return lipgloss.NewStyle().Render("?")
}

// RenderSeparator renders a visual separator
func RenderSeparator(width int) string {
	if width <= 0 {
		width = 64
	}
	return mutedStyle.Render(lipgloss.NewStyle().Width(width).Render("─" + lipgloss.NewStyle().Render(repeatString("─", width-2)) + "─"))
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
