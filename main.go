package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func initialModel() model {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".csv"}
	fp.CurrentDirectory, _ = os.UserHomeDir()

	return model{
		screen:      screenWelcome,
		input:       "",
		currentBand: 1,
		capacitorReading: CapacitorReading{
			BandCount: 5, // Default to 5 bands
		},
		resistorReading: ResistorReading{
			BandCount: 4, // Default to 4 bands
		},
		history:    []ComponentEntry{},
		filepicker: fp,
	}
}

type screenType int

const (
	screenWelcome screenType = iota
	screenComponentSelection
	screenTypeSelection
	screenBandCountSelection
	screenBandInput
	screenReview
	screenResults
	screenNoteInput
	screenEdit
	screenFilePicker
)

type model struct {
	screen          screenType
	input           string
	suggestion      string // Autocomplete suggestion for current input
	err             error
	successMsg      string // Success message (e.g., export success)
	quitting        bool
	currentBand     int // Current band being input (1-6)
	componentType   ComponentType
	capacitorReading CapacitorReading
	resistorReading  ResistorReading
	capacitorResult  *CalculationResult
	resistorResult   *ResistorResult
	editBandIndex   int // For edit mode
	currentNote     string // Current note being edited
	history         []ComponentEntry // History of decoded components
	filepicker      filepicker.Model // File picker for export
	selectedFile    string // Selected export file path
}

func (m model) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	// Handle filepicker messages when on filepicker screen
	if m.screen == screenFilePicker {
		var cmd tea.Cmd
		m.filepicker, cmd = m.filepicker.Update(msg)

		// Check if a file was selected
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			m.selectedFile = path
			// Perform export
			err := ExportToCSV(m.history, path)
			if err != nil {
				m.err = fmt.Errorf("export failed: %v", err)
				m.successMsg = ""
			} else {
				m.err = nil
				m.successMsg = fmt.Sprintf("✓ Successfully exported %d capacitor%s to %s",
					len(m.history),
					map[bool]string{true: "", false: "s"}[len(m.history) == 1],
					path)
			}
			// Return to results screen
			m.screen = screenResults
		}

		return m, cmd
	}

	return m, nil
}

func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global quit
	if key == "ctrl+c" {
		m.quitting = true
		return m, tea.Quit
	}

	switch m.screen {
	case screenWelcome:
		return m.handleWelcomeInput(key)
	case screenComponentSelection:
		return m.handleComponentSelectionInput(key)
	case screenTypeSelection:
		return m.handleTypeSelectionInput(key)
	case screenBandCountSelection:
		return m.handleBandCountInput(key)
	case screenBandInput:
		return m.handleBandInputInput(key)
	case screenReview:
		return m.handleReviewInput(key)
	case screenResults:
		return m.handleResultsInput(key)
	case screenNoteInput:
		return m.handleNoteInputInput(key)
	case screenEdit:
		return m.handleEditInput(key)
	}

	return m, nil
}

func (m model) handleWelcomeInput(key string) (tea.Model, tea.Cmd) {
	if key == "enter" || key == " " {
		m.screen = screenComponentSelection
		m.input = ""
		m.err = nil
	} else if key == "q" {
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleComponentSelectionInput(key string) (tea.Model, tea.Cmd) {
	lowerKey := strings.ToLower(key)

	if key == "enter" && m.input != "" {
		if lowerKey == "c" {
			m.componentType = ComponentCapacitor
			m.screen = screenTypeSelection
			m.input = ""
			m.err = nil
		} else if lowerKey == "r" {
			m.componentType = ComponentResistor
			m.screen = screenBandCountSelection
			m.input = ""
			m.err = nil
		} else {
			m.err = fmt.Errorf("invalid selection: please enter C for Capacitor or R for Resistor")
		}
	} else if key == "backspace" || key == "delete" {
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	} else if len(key) == 1 {
		m.input += key
	}
	return m, nil
}

func (m model) handleTypeSelectionInput(key string) (tea.Model, tea.Cmd) {
	if key == "enter" && m.input != "" {
		capType, valid := ParseCapacitorType(m.input)
		if valid {
			m.capacitorReading.CapType = capType
			m.screen = screenBandCountSelection
			m.input = ""
			m.err = nil
		} else {
			m.err = fmt.Errorf("invalid type: please enter J, K, L, M, or N")
		}
	} else if key == "backspace" || key == "delete" {
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	} else if len(key) == 1 {
		m.input += key
	}
	return m, nil
}

func (m model) handleBandCountInput(key string) (tea.Model, tea.Cmd) {
	if key == "enter" && m.input != "" {
		bandCount := 0
		if m.input == "3" || m.input == "4" || m.input == "5" || m.input == "6" {
			if m.input == "3" {
				bandCount = 3
			} else if m.input == "4" {
				bandCount = 4
			} else if m.input == "5" {
				bandCount = 5
			} else if m.input == "6" {
				bandCount = 6
			}

			// Validate band count based on component type
			if m.componentType == ComponentCapacitor {
				if bandCount < 3 || bandCount > 5 {
					m.err = fmt.Errorf("invalid band count for capacitor: please enter 3, 4, or 5")
					return m, nil
				}
				m.capacitorReading.BandCount = bandCount
			} else if m.componentType == ComponentResistor {
				if bandCount < 4 || bandCount > 6 {
					m.err = fmt.Errorf("invalid band count for resistor: please enter 4, 5, or 6")
					return m, nil
				}
				m.resistorReading.BandCount = bandCount
			}

			m.screen = screenBandInput
			m.currentBand = 1
			m.input = ""
			m.err = nil
		} else {
			if m.componentType == ComponentCapacitor {
				m.err = fmt.Errorf("invalid band count: please enter 3, 4, or 5")
			} else {
				m.err = fmt.Errorf("invalid band count: please enter 4, 5, or 6")
			}
		}
	} else if key == "backspace" || key == "delete" {
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	} else if len(key) == 1 {
		m.input += key
	}
	return m, nil
}

func (m model) handleBandInputInput(key string) (tea.Model, tea.Cmd) {
	if key == "tab" {
		// Accept autocomplete suggestion
		if m.suggestion != "" {
			m.input = GetFullColorFromInput(m.input, m.suggestion)
			m.suggestion = "" // Clear suggestion after accepting
		}
		return m, nil
	} else if key == "enter" && m.input != "" {
		color, valid := ParseColor(m.input)
		if !valid {
			m.err = fmt.Errorf("invalid color: '%s' - please enter a valid color name", m.input)
			return m, nil
		}

		// Validate and store based on component type
		var validationErr error
		var bandCount int

		if m.componentType == ComponentCapacitor {
			bandCount = m.capacitorReading.BandCount
			// Validate based on current band
			switch m.currentBand {
			case 1:
				validationErr = ValidateBand1(color)
				if validationErr == nil {
					m.capacitorReading.Band1 = color
				}
			case 2:
				validationErr = ValidateBand2(color)
				if validationErr == nil {
					m.capacitorReading.Band2 = color
				}
			case 3:
				validationErr = ValidateBand3(color)
				if validationErr == nil {
					m.capacitorReading.Band3 = color
				}
			case 4:
				// Calculate capacitance for validation
				info1 := GetColorInfo(m.capacitorReading.Band1)
				info2 := GetColorInfo(m.capacitorReading.Band2)
				info3 := GetColorInfo(m.capacitorReading.Band3)
				capacitancePF := float64(info1.Digit*10+info2.Digit) * info3.Multiplier

				validationErr = ValidateBand4(color, capacitancePF)
				if validationErr == nil {
					m.capacitorReading.Band4 = color
				}
			case 5:
				validationErr = ValidateBand5(color, m.capacitorReading.CapType, m.capacitorReading.BandCount)
				if validationErr == nil {
					m.capacitorReading.Band5 = color
				}
			}
		} else if m.componentType == ComponentResistor {
			bandCount = m.resistorReading.BandCount
			// Validate resistor bands
			switch m.currentBand {
			case 1:
				validationErr = ValidateResistorBand1(color)
				if validationErr == nil {
					m.resistorReading.Band1 = color
				}
			case 2:
				validationErr = ValidateResistorBand2(color)
				if validationErr == nil {
					m.resistorReading.Band2 = color
				}
			case 3:
				// For 4-band resistors, band 3 is the multiplier
				// For 5/6-band resistors, band 3 is the third digit
				if bandCount == 4 {
					validationErr = ValidateResistorMultiplier(color, 3)
					if validationErr == nil {
						m.resistorReading.Band3 = color
					}
				} else {
					validationErr = ValidateResistorBand3(color)
					if validationErr == nil {
						m.resistorReading.Band3 = color
					}
				}
			case 4:
				// For 4-band resistors, band 4 is tolerance
				// For 5/6-band resistors, band 4 is multiplier
				if bandCount == 4 {
					validationErr = ValidateResistorTolerance(color, 4)
					if validationErr == nil {
						m.resistorReading.Band4 = color
					}
				} else {
					validationErr = ValidateResistorMultiplier(color, 4)
					if validationErr == nil {
						m.resistorReading.Band4 = color
					}
				}
			case 5:
				validationErr = ValidateResistorTolerance(color, 5)
				if validationErr == nil {
					m.resistorReading.Band5 = color
				}
			case 6:
				validationErr = ValidateResistorTempCoeff(color)
				if validationErr == nil {
					m.resistorReading.Band6 = color
				}
			}
		}

		if validationErr != nil {
			m.err = validationErr
			return m, nil
		}

		// Move to next band or review screen
		if m.currentBand < bandCount {
			m.currentBand++
			m.input = ""
			m.suggestion = "" // Clear suggestion
			m.err = nil
		} else {
			// All bands entered, go to review
			m.screen = screenReview
			m.input = ""
			m.suggestion = "" // Clear suggestion
			m.err = nil
		}
	} else if key == "backspace" || key == "delete" {
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
			// Update suggestion after deleting character
			m.suggestion = GetColorSuggestion(m.input, m.currentBand)
		}
	} else if len(key) == 1 {
		m.input += key
		// Update suggestion after adding character
		m.suggestion = GetColorSuggestion(m.input, m.currentBand)
	}
	return m, nil
}

func (m model) handleReviewInput(key string) (tea.Model, tea.Cmd) {
	lowerKey := strings.ToLower(key)

	if lowerKey == "enter" || lowerKey == " " {
		// Calculate and show results
		result, err := Calculate(m.reading)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.result = result
		m.screen = screenResults
		m.err = nil
	} else if lowerKey == "c" {
		// Go to edit mode
		m.screen = screenEdit
		m.input = ""
		m.err = nil
	} else if lowerKey == "q" {
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) handleResultsInput(key string) (tea.Model, tea.Cmd) {
	lowerKey := strings.ToLower(key)

	if lowerKey == "d" {
		// Decode another - reset to type selection
		m.screen = screenTypeSelection
		m.input = ""
		m.currentBand = 1
		m.err = nil
		m.successMsg = ""
		m.reading = CapacitorReading{BandCount: 5}
		m.result = nil
		m.currentNote = ""
	} else if lowerKey == "e" {
		// Edit current - go to edit mode
		m.screen = screenEdit
		m.input = ""
		m.err = nil
		m.successMsg = ""
	} else if lowerKey == "n" {
		// Add/Edit note
		m.screen = screenNoteInput
		m.input = m.currentNote // Pre-fill with existing note
		m.err = nil
		m.successMsg = ""
	} else if lowerKey == "x" {
		// Check if there's data to export
		if len(m.history) == 0 {
			m.err = fmt.Errorf("no data to export (decode at least one capacitor first)")
			m.successMsg = ""
		} else {
			// Navigate to file picker
			m.screen = screenFilePicker
			m.err = nil
			m.successMsg = ""
		}
	} else if lowerKey == "q" {
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) handleEditInput(key string) (tea.Model, tea.Cmd) {
	if key == "enter" && m.input != "" {
		// Parse band number to edit
		if m.input == "1" {
			m.editBandIndex = 1
			m.currentBand = 1
			m.screen = screenBandInput
			m.input = ""
			m.err = nil
		} else if m.input == "2" {
			m.editBandIndex = 2
			m.currentBand = 2
			m.screen = screenBandInput
			m.input = ""
			m.err = nil
		} else if m.input == "3" {
			m.editBandIndex = 3
			m.currentBand = 3
			m.screen = screenBandInput
			m.input = ""
			m.err = nil
		} else if m.input == "4" {
			m.editBandIndex = 4
			m.currentBand = 4
			m.screen = screenBandInput
			m.input = ""
			m.err = nil
		} else if m.input == "5" && m.reading.BandCount >= 4 {
			m.editBandIndex = 5
			m.currentBand = 5
			m.screen = screenBandInput
			m.input = ""
			m.err = nil
		} else {
			m.err = fmt.Errorf("invalid band number")
		}
	} else if strings.ToLower(key) == "q" {
		// Cancel edit, go back to review
		m.screen = screenReview
		m.input = ""
		m.err = nil
	} else if key == "backspace" || key == "delete" {
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	} else if len(key) == 1 {
		m.input += key
	}

	return m, nil
}

func (m model) handleNoteInputInput(key string) (tea.Model, tea.Cmd) {
	if key == "enter" {
		// Save note and add to history
		m.currentNote = m.input

		// Add current result with note to history
		if m.result != nil {
			entry := CapacitorEntry{
				Result: m.result,
				Note:   m.currentNote,
			}
			m.history = append(m.history, entry)
		}

		// Go back to results screen
		m.screen = screenResults
		m.input = ""
		m.err = nil
	} else if key == "esc" {
		// Cancel note input, go back to results
		m.screen = screenResults
		m.input = ""
		m.err = nil
	} else if key == "backspace" || key == "delete" {
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	} else if len(key) == 1 {
		// Limit note to 200 characters
		if len(m.input) < 200 {
			m.input += key
		} else {
			m.err = fmt.Errorf("note too long (max 200 characters)")
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return successStyle.Render("\n✓ Thanks for using Tropical Fish Decoder!\n\n")
	}

	switch m.screen {
	case screenWelcome:
		return m.renderWelcome()
	case screenTypeSelection:
		return m.renderTypeSelection()
	case screenBandCountSelection:
		return m.renderBandCountSelection()
	case screenBandInput:
		return m.renderBandInput()
	case screenReview:
		return m.renderReview()
	case screenResults:
		return m.renderResults()
	case screenNoteInput:
		return m.renderNoteInput()
	case screenEdit:
		return m.renderEdit()
	case screenFilePicker:
		return m.renderFilePicker()
	}

	return "Unknown screen\n"
}

func (m model) renderWelcome() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render("═══════════════════════════════════════════════════════════════"))
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("    TROPICAL FISH CAPACITOR COLOR CODE DECODER"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("        IEC 60062 Standard 5-Band Capacitor Calculator"))
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("═══════════════════════════════════════════════════════════════"))
	b.WriteString("\n\n")

	b.WriteString(valueStyle.Render("Decode color bands from your capacitor to determine its value,"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("tolerance, and voltage rating."))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Available capacitor types:"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  J = Dipped Tantalum"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  K = Mica"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  L = Polyester / Polystyrene"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  M = Electrolytic (4-band style)"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  N = Electrolytic (3-band style)"))
	b.WriteString("\n\n")

	b.WriteString(RenderSeparator(64))
	b.WriteString("\n\n")

	b.WriteString(promptStyle.Render("Press ENTER to begin, or Q to quit"))
	b.WriteString("\n")

	return b.String()
}

func (m model) renderTypeSelection() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" STEP 1: SELECT CAPACITOR TYPE "))
	b.WriteString("\n\n")

	b.WriteString(valueStyle.Render("Available types:"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  J = Dipped Tantalum"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  K = Mica"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  L = Polyester / Polystyrene"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  M = Electrolytic (4-band)"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  N = Electrolytic (3-band)"))
	b.WriteString("\n\n")

	b.WriteString(promptStyle.Render("Enter capacitor type (J/K/L/M/N): "))
	b.WriteString(inputStyle.Render(m.input))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press Ctrl+C to quit"))
	b.WriteString("\n")

	return b.String()
}

func (m model) renderBandCountSelection() string {
	var b strings.Builder

	typeInfo, _ := GetTypeInfo(m.reading.CapType)

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" STEP 2: SELECT BAND COUNT "))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Capacitor Type: "))
	b.WriteString(valueStyle.Render(string(m.reading.CapType) + " (" + typeInfo.Description + ")"))
	b.WriteString("\n\n")

	b.WriteString(valueStyle.Render("How many color bands does your capacitor have?"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  3 = 3-band (value + multiplier + tolerance)"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  4 = 4-band (value + multiplier + tolerance + voltage)"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  5 = 5-band (value + multiplier + tolerance + voltage + temp coeff)"))
	b.WriteString("\n\n")

	b.WriteString(promptStyle.Render("Enter band count (3/4/5): "))
	b.WriteString(inputStyle.Render(m.input))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press Ctrl+C to quit"))
	b.WriteString("\n")

	return b.String()
}

func (m model) renderBandInput() string {
	var b strings.Builder

	typeInfo, _ := GetTypeInfo(m.reading.CapType)

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(fmt.Sprintf(" BAND INPUT (%d of %d) ", m.currentBand, m.reading.BandCount)))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Type: "))
	b.WriteString(valueStyle.Render(string(m.reading.CapType) + " (" + typeInfo.Name + ")"))
	b.WriteString("  ")
	b.WriteString(labelStyle.Render("Bands: "))
	b.WriteString(valueStyle.Render(fmt.Sprintf("%d", m.reading.BandCount)))
	b.WriteString("\n\n")

	b.WriteString(mutedStyle.Render("Valid colors: Black, Brown, Red, Orange, Yellow, Green, Blue,"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("              Violet, Grey, White, Gold, Silver"))
	b.WriteString("\n\n")

	// Show previously entered bands
	if m.currentBand > 1 {
		b.WriteString(labelStyle.Render("Already entered:"))
		b.WriteString("\n")
		for i := 1; i < m.currentBand; i++ {
			var color Color
			switch i {
			case 1:
				color = m.reading.Band1
			case 2:
				color = m.reading.Band2
			case 3:
				color = m.reading.Band3
			case 4:
				color = m.reading.Band4
			case 5:
				color = m.reading.Band5
			}
			b.WriteString(confirmStyle.Render(fmt.Sprintf("  ✓ Band %d: ", i)))
			b.WriteString(RenderColorBand(color, i))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Current band input
	b.WriteString(bandHeaderStyle.Render(fmt.Sprintf("─ BAND %d (%s) ", m.currentBand, GetBandName(m.currentBand))))
	b.WriteString("\n\n")

	b.WriteString(promptStyle.Render(fmt.Sprintf("Enter Band %d color: ", m.currentBand)))
	b.WriteString(inputStyle.Render(m.input))

	// Show autocomplete suggestion in grey
	if m.suggestion != "" {
		b.WriteString(mutedStyle.Render(m.suggestion))
	}
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Show hint for autocomplete
	if m.suggestion != "" {
		b.WriteString(helpStyle.Render("Press Tab to autocomplete, Enter to submit, Ctrl+C to quit"))
	} else {
		b.WriteString(helpStyle.Render("Press Enter to submit, Ctrl+C to quit"))
	}
	b.WriteString("\n")

	return b.String()
}

func (m model) renderReview() string {
	var b strings.Builder

	typeInfo, _ := GetTypeInfo(m.reading.CapType)

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" REVIEW YOUR INPUT "))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Capacitor Type: "))
	b.WriteString(valueStyle.Render(string(m.reading.CapType) + " (" + typeInfo.Description + ")"))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Band Count: "))
	b.WriteString(valueStyle.Render(fmt.Sprintf("%d", m.reading.BandCount)))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Bands entered:"))
	b.WriteString("\n")

	// Band 1
	b.WriteString(valueStyle.Render("  Band 1: "))
	b.WriteString(RenderColorBand(m.reading.Band1, 1))
	b.WriteString("\n")

	// Band 2
	b.WriteString(valueStyle.Render("  Band 2: "))
	b.WriteString(RenderColorBand(m.reading.Band2, 2))
	b.WriteString("\n")

	// Band 3
	b.WriteString(valueStyle.Render("  Band 3: "))
	b.WriteString(RenderColorBand(m.reading.Band3, 3))
	b.WriteString("\n")

	// Band 4
	b.WriteString(valueStyle.Render("  Band 4: "))
	b.WriteString(RenderColorBand(m.reading.Band4, 4))
	b.WriteString("\n")

	// Band 5 (if applicable)
	if m.reading.BandCount >= 4 {
		b.WriteString(valueStyle.Render("  Band 5: "))
		b.WriteString(RenderColorBand(m.reading.Band5, 5))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(RenderSeparator(64))
	b.WriteString("\n\n")

	b.WriteString(promptStyle.Render("Press ENTER to calculate, (C)orrect a band, or (Q)uit"))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) renderResults() string {
	if m.result == nil {
		return errorStyle.Render("\n✗ No calculation results available\n")
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(resultHeaderStyle.Width(64).Render("RESULTS"))
	b.WriteString("\n")
	b.WriteString(resultHeaderStyle.Width(64).Render("════════════════════════════════════════════════════════════════"))
	b.WriteString("\n\n")

	typeInfo, _ := GetTypeInfo(m.result.Reading.CapType)

	// Type and configuration
	b.WriteString(resultLabelStyle.Render("Capacitor Type:"))
	b.WriteString("  ")
	b.WriteString(resultValueStyle.Render(string(m.result.Reading.CapType) + " (" + typeInfo.Name + ")"))
	b.WriteString("\n")

	b.WriteString(resultLabelStyle.Render("Configuration:"))
	b.WriteString("  ")
	b.WriteString(resultValueStyle.Render(fmt.Sprintf("%d-band", m.result.Reading.BandCount)))
	b.WriteString("\n\n")

	// Capacitance value
	b.WriteString(labelStyle.Render("CAPACITANCE VALUE:"))
	b.WriteString("\n")
	b.WriteString(resultLabelStyle.Render("Value:"))
	b.WriteString("  ")
	b.WriteString(resultValueStyle.Render(FormatCapacitanceWithUF(m.result.CapacitanceValue, m.result.CapacitanceUnit, m.result.CapacitancePF)))
	b.WriteString("\n\n")

	// Tolerance
	b.WriteString(labelStyle.Render("TOLERANCE:"))
	b.WriteString("\n")
	b.WriteString(resultLabelStyle.Render("Specification:"))
	b.WriteString("  ")
	tolStr := FormatTolerance(m.result)
	if m.result.ToleranceType == "absolute" {
		tolStr += " (absolute, value ≤ 10pF)"
	} else {
		tolStr += " (percentage-based, value > 10pF)"
	}
	b.WriteString(resultValueStyle.Render(tolStr))
	b.WriteString("\n")

	b.WriteString(resultLabelStyle.Render("Range:"))
	b.WriteString("  ")
	b.WriteString(resultValueStyle.Render(FormatToleranceRange(m.result)))
	b.WriteString("\n\n")

	// Voltage rating
	if m.result.VoltageValid {
		b.WriteString(labelStyle.Render("VOLTAGE RATING:"))
		b.WriteString("\n")
		b.WriteString(resultLabelStyle.Render("Voltage:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render(FormatVoltage(m.result) + " (Type " + string(m.result.Reading.CapType) + " " + typeInfo.Name + ")"))
		b.WriteString("\n\n")
	}

	// Temperature coefficient
	if m.result.TempCoeffValid {
		b.WriteString(labelStyle.Render("TEMPERATURE COEFFICIENT:"))
		b.WriteString("\n")
		b.WriteString(resultLabelStyle.Render("Coefficient:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render(FormatTempCoefficient(m.result)))
		b.WriteString("\n\n")
	}

	// Current note
	if m.currentNote != "" {
		b.WriteString(labelStyle.Render("NOTE:"))
		b.WriteString("\n")
		b.WriteString(resultLabelStyle.Render("User Note:"))
		b.WriteString("  ")
		b.WriteString(valueStyle.Render(m.currentNote))
		b.WriteString("\n\n")
	}

	b.WriteString(resultHeaderStyle.Width(64).Render("════════════════════════════════════════════════════════════════"))
	b.WriteString("\n\n")

	// Show export success message
	if m.successMsg != "" {
		b.WriteString(successStyle.Render(m.successMsg))
		b.WriteString("\n\n")
	}

	// Show export error
	if m.err != nil {
		if strings.Contains(m.err.Error(), "export") {
			b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
			b.WriteString("\n\n")
		}
	}

	b.WriteString(promptStyle.Render("(D)ecode  |  (E)dit  |  (N)ote  |  e(X)port  |  (Q)uit"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("Decoded capacitors in history: %d", len(m.history))))
	b.WriteString("\n")
	b.WriteString(RenderSeparator(64))
	b.WriteString("\n")

	return b.String()
}

func (m model) renderEdit() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" EDIT BAND "))
	b.WriteString("\n\n")

	b.WriteString(valueStyle.Render("Which band do you want to correct?"))
	b.WriteString("\n\n")

	// Show all bands
	for i := 1; i <= m.reading.BandCount; i++ {
		var color Color
		switch i {
		case 1:
			color = m.reading.Band1
		case 2:
			color = m.reading.Band2
		case 3:
			color = m.reading.Band3
		case 4:
			color = m.reading.Band4
		case 5:
			color = m.reading.Band5
		}

		b.WriteString(valueStyle.Render(fmt.Sprintf("  %d = ", i)))
		b.WriteString(RenderColorBand(color, i))
		b.WriteString("\n")
	}

	b.WriteString(valueStyle.Render("  Q = Cancel (keep current values)"))
	b.WriteString("\n\n")

	b.WriteString(promptStyle.Render(fmt.Sprintf("Select band (1-%d/Q): ", m.reading.BandCount)))
	b.WriteString(inputStyle.Render(m.input))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press Ctrl+C to quit"))
	b.WriteString("\n")

	return b.String()
}

func (m model) renderNoteInput() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" ADD / EDIT NOTE "))
	b.WriteString("\n\n")

	b.WriteString(valueStyle.Render("Add a note to this capacitor reading (max 200 characters):"))
	b.WriteString("\n\n")

	b.WriteString(promptStyle.Render("Note: "))
	b.WriteString(inputStyle.Render(m.input))
	b.WriteString("\n")

	// Character count
	charCount := len(m.input)
	charCountStr := fmt.Sprintf("%d/200 characters", charCount)
	if charCount > 200 {
		b.WriteString(errorStyle.Render(charCountStr))
	} else if charCount > 180 {
		b.WriteString(mutedStyle.Render(charCountStr))
	} else {
		b.WriteString(mutedStyle.Render(charCountStr))
	}
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Press ENTER to save, ESC to cancel, Ctrl+C to quit"))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) renderFilePicker() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" EXPORT TO CSV - SELECT FILE LOCATION "))
	b.WriteString("\n\n")

	b.WriteString(m.filepicker.View())
	b.WriteString("\n")

	b.WriteString(helpStyle.Render("Navigate: ↑/↓/←/→  |  ENTER: Select  |  q: Cancel  |  Ctrl+C: Quit"))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("✗ " + m.err.Error()))
		b.WriteString("\n")
	}

	return b.String()
}
