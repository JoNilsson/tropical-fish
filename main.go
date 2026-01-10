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
	screen           screenType
	input            string
	suggestion       string // Autocomplete suggestion for current input
	err              error
	successMsg       string // Success message (e.g., export success)
	quitting         bool
	currentBand      int // Current band being input (1-6)
	componentType    ComponentType
	capacitorReading CapacitorReading
	resistorReading  ResistorReading
	capacitorResult  *CalculationResult
	resistorResult   *ResistorResult
	editBandIndex    int              // For edit mode
	currentNote      string           // Current note being edited
	history          []ComponentEntry // History of decoded components
	filepicker       filepicker.Model // File picker for export
	selectedFile     string           // Selected export file path
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
		// Calculate and show results based on component type
		if m.componentType == ComponentCapacitor {
			result, err := Calculate(m.capacitorReading)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.capacitorResult = result
			m.resistorResult = nil
		} else if m.componentType == ComponentResistor {
			result, err := CalculateResistor(m.resistorReading)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.resistorResult = result
			m.capacitorResult = nil
		}
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
		// Decode another - reset to component selection
		m.screen = screenComponentSelection
		m.input = ""
		m.currentBand = 1
		m.err = nil
		m.successMsg = ""
		m.capacitorReading = CapacitorReading{BandCount: 5}
		m.resistorReading = ResistorReading{BandCount: 4}
		m.capacitorResult = nil
		m.resistorResult = nil
		m.currentNote = ""
	} else if lowerKey == "e" {
		// Edit current - go to edit mode
		m.screen = screenEdit
		m.input = ""
		m.err = nil
		m.successMsg = ""
	} else if lowerKey == "n" {
		// Add/Edit note - save to history first if not already saved
		if m.currentNote == "" && len(m.history) == 0 ||
			(len(m.history) > 0 && m.history[len(m.history)-1].Note != m.currentNote) {
			// Add current result to history
			entry := ComponentEntry{
				ComponentType:   m.componentType,
				CapacitorResult: m.capacitorResult,
				ResistorResult:  m.resistorResult,
				Note:            m.currentNote,
			}
			m.history = append(m.history, entry)
		}
		m.screen = screenNoteInput
		m.input = m.currentNote // Pre-fill with existing note
		m.err = nil
		m.successMsg = ""
	} else if lowerKey == "x" {
		// Add current result to history if not already there
		if len(m.history) == 0 ||
			(len(m.history) > 0 && (m.history[len(m.history)-1].CapacitorResult != m.capacitorResult ||
				m.history[len(m.history)-1].ResistorResult != m.resistorResult)) {
			entry := ComponentEntry{
				ComponentType:   m.componentType,
				CapacitorResult: m.capacitorResult,
				ResistorResult:  m.resistorResult,
				Note:            m.currentNote,
			}
			m.history = append(m.history, entry)
		}

		// Check if there's data to export
		if len(m.history) == 0 {
			m.err = fmt.Errorf("no data to export (decode at least one component first)")
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
		var maxBand int
		if m.componentType == ComponentCapacitor {
			maxBand = m.capacitorReading.BandCount
		} else {
			maxBand = m.resistorReading.BandCount
		}

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
		} else if m.input == "5" && maxBand >= 5 {
			m.editBandIndex = 5
			m.currentBand = 5
			m.screen = screenBandInput
			m.input = ""
			m.err = nil
		} else if m.input == "6" && maxBand >= 6 {
			m.editBandIndex = 6
			m.currentBand = 6
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
		// Save note and update/add to history
		m.currentNote = m.input

		// Check if we need to update existing entry or add new one
		if len(m.history) > 0 {
			// Update last entry's note if it matches current result
			lastEntry := &m.history[len(m.history)-1]
			if lastEntry.CapacitorResult == m.capacitorResult &&
				lastEntry.ResistorResult == m.resistorResult {
				lastEntry.Note = m.currentNote
			} else {
				// Add new entry
				entry := ComponentEntry{
					ComponentType:   m.componentType,
					CapacitorResult: m.capacitorResult,
					ResistorResult:  m.resistorResult,
					Note:            m.currentNote,
				}
				m.history = append(m.history, entry)
			}
		} else {
			// Add first entry
			entry := ComponentEntry{
				ComponentType:   m.componentType,
				CapacitorResult: m.capacitorResult,
				ResistorResult:  m.resistorResult,
				Note:            m.currentNote,
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
	case screenComponentSelection:
		return m.renderComponentSelection()
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
	b.WriteString(titleStyle.Render("    TROPICAL FISH COMPONENT COLOR CODE DECODER"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("            Welcome Screen"))
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("═══════════════════════════════════════════════════════════════"))
	b.WriteString("\n\n")

	b.WriteString(valueStyle.Render("Decode color bands from capacitors and resistors to determine"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("values, tolerances, and ratings."))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Supported components:"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  • Capacitors - IEC 60062 Standard (3/4/5 bands)"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  • Resistors - EIA Standard (4/5/6 bands)"))
	b.WriteString("\n\n")

	b.WriteString(RenderSeparator(64))
	b.WriteString("\n\n")

	b.WriteString(promptStyle.Render("Press ENTER to begin, or Q to quit"))
	b.WriteString("\n")

	return b.String()
}

func (m model) renderComponentSelection() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" STEP 1: SELECT COMPONENT TYPE "))
	b.WriteString("\n\n")

	b.WriteString(valueStyle.Render("What would you like to decode?"))
	b.WriteString("\n\n")

	b.WriteString(valueStyle.Render("  (C) Capacitor - IEC 60062 Standard (3/4/5 bands)"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render("  (R) Resistor - EIA Standard (4/5/6 bands)"))
	b.WriteString("\n\n")

	b.WriteString(promptStyle.Render("Select component type (C/R): "))
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

	stepNum := "STEP 2"
	if m.componentType == ComponentResistor {
		stepNum = "STEP 2" // Still step 2 for resistors
	} else {
		stepNum = "STEP 3" // Step 3 for capacitors (after type selection)
	}

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(fmt.Sprintf(" %s: SELECT BAND COUNT ", stepNum)))
	b.WriteString("\n\n")

	if m.componentType == ComponentCapacitor {
		typeInfo, _ := GetTypeInfo(m.capacitorReading.CapType)
		b.WriteString(labelStyle.Render("Capacitor Type: "))
		b.WriteString(valueStyle.Render(string(m.capacitorReading.CapType) + " (" + typeInfo.Description + ")"))
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
	} else if m.componentType == ComponentResistor {
		b.WriteString(labelStyle.Render("Component: "))
		b.WriteString(valueStyle.Render("Resistor"))
		b.WriteString("\n\n")

		b.WriteString(valueStyle.Render("How many color bands does your resistor have?"))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render("  4 = 4-band (standard, ±5% or ±10% tolerance)"))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render("  5 = 5-band (precision, ±1% or ±2% tolerance)"))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render("  6 = 6-band (precision + temperature coefficient)"))
		b.WriteString("\n\n")

		b.WriteString(promptStyle.Render("Enter band count (4/5/6): "))
	}

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

	var bandCount int
	var bandName string

	if m.componentType == ComponentCapacitor {
		bandCount = m.capacitorReading.BandCount
		bandName = GetBandName(m.currentBand)
	} else {
		bandCount = m.resistorReading.BandCount
		bandName = GetResistorBandName(m.currentBand, bandCount)
	}

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(fmt.Sprintf(" BAND INPUT (%d of %d) ", m.currentBand, bandCount)))
	b.WriteString("\n\n")

	if m.componentType == ComponentCapacitor {
		typeInfo, _ := GetTypeInfo(m.capacitorReading.CapType)
		b.WriteString(labelStyle.Render("Type: "))
		b.WriteString(valueStyle.Render(string(m.capacitorReading.CapType) + " (" + typeInfo.Name + ")"))
		b.WriteString("  ")
	} else {
		b.WriteString(labelStyle.Render("Component: "))
		b.WriteString(valueStyle.Render("Resistor"))
		b.WriteString("  ")
	}

	b.WriteString(labelStyle.Render("Bands: "))
	b.WriteString(valueStyle.Render(fmt.Sprintf("%d", bandCount)))
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
			if m.componentType == ComponentCapacitor {
				switch i {
				case 1:
					color = m.capacitorReading.Band1
				case 2:
					color = m.capacitorReading.Band2
				case 3:
					color = m.capacitorReading.Band3
				case 4:
					color = m.capacitorReading.Band4
				case 5:
					color = m.capacitorReading.Band5
				}
			} else {
				switch i {
				case 1:
					color = m.resistorReading.Band1
				case 2:
					color = m.resistorReading.Band2
				case 3:
					color = m.resistorReading.Band3
				case 4:
					color = m.resistorReading.Band4
				case 5:
					color = m.resistorReading.Band5
				case 6:
					color = m.resistorReading.Band6
				}
			}
			b.WriteString(confirmStyle.Render(fmt.Sprintf("  ✓ Band %d: ", i)))
			b.WriteString(RenderColorBand(color, i))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Current band input
	b.WriteString(bandHeaderStyle.Render(fmt.Sprintf("─ BAND %d (%s) ", m.currentBand, bandName)))
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

	b.WriteString("\n")
	b.WriteString(headerStyle.Render(" REVIEW YOUR INPUT "))
	b.WriteString("\n\n")

	if m.componentType == ComponentCapacitor {
		typeInfo, _ := GetTypeInfo(m.capacitorReading.CapType)
		b.WriteString(labelStyle.Render("Capacitor Type: "))
		b.WriteString(valueStyle.Render(string(m.capacitorReading.CapType) + " (" + typeInfo.Description + ")"))
		b.WriteString("\n")

		b.WriteString(labelStyle.Render("Band Count: "))
		b.WriteString(valueStyle.Render(fmt.Sprintf("%d", m.capacitorReading.BandCount)))
		b.WriteString("\n\n")

		b.WriteString(labelStyle.Render("Bands entered:"))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render("  Band 1: "))
		b.WriteString(RenderColorBand(m.capacitorReading.Band1, 1))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render("  Band 2: "))
		b.WriteString(RenderColorBand(m.capacitorReading.Band2, 2))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render("  Band 3: "))
		b.WriteString(RenderColorBand(m.capacitorReading.Band3, 3))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render("  Band 4: "))
		b.WriteString(RenderColorBand(m.capacitorReading.Band4, 4))
		b.WriteString("\n")
		if m.capacitorReading.BandCount >= 4 {
			b.WriteString(valueStyle.Render("  Band 5: "))
			b.WriteString(RenderColorBand(m.capacitorReading.Band5, 5))
			b.WriteString("\n")
		}
	} else {
		b.WriteString(labelStyle.Render("Component: "))
		b.WriteString(valueStyle.Render("Resistor"))
		b.WriteString("\n")

		b.WriteString(labelStyle.Render("Band Count: "))
		b.WriteString(valueStyle.Render(fmt.Sprintf("%d", m.resistorReading.BandCount)))
		b.WriteString("\n\n")

		b.WriteString(labelStyle.Render("Bands entered:"))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render("  Band 1: "))
		b.WriteString(RenderColorBand(m.resistorReading.Band1, 1))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render("  Band 2: "))
		b.WriteString(RenderColorBand(m.resistorReading.Band2, 2))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render("  Band 3: "))
		b.WriteString(RenderColorBand(m.resistorReading.Band3, 3))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render("  Band 4: "))
		b.WriteString(RenderColorBand(m.resistorReading.Band4, 4))
		b.WriteString("\n")
		if m.resistorReading.BandCount >= 5 {
			b.WriteString(valueStyle.Render("  Band 5: "))
			b.WriteString(RenderColorBand(m.resistorReading.Band5, 5))
			b.WriteString("\n")
		}
		if m.resistorReading.BandCount == 6 {
			b.WriteString(valueStyle.Render("  Band 6: "))
			b.WriteString(RenderColorBand(m.resistorReading.Band6, 6))
			b.WriteString("\n")
		}
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
	if m.capacitorResult == nil && m.resistorResult == nil {
		return errorStyle.Render("\n✗ No calculation results available\n")
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(resultHeaderStyle.Width(64).Render("RESULTS"))
	b.WriteString("\n")
	b.WriteString(resultHeaderStyle.Width(64).Render("════════════════════════════════════════════════════════════════"))
	b.WriteString("\n\n")

	if m.componentType == ComponentCapacitor && m.capacitorResult != nil {
		result := m.capacitorResult
		typeInfo, _ := GetTypeInfo(result.Reading.CapType)

		// Type and configuration
		b.WriteString(resultLabelStyle.Render("Component Type:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render("Capacitor"))
		b.WriteString("\n")

		b.WriteString(resultLabelStyle.Render("Capacitor Type:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render(string(result.Reading.CapType) + " (" + typeInfo.Name + ")"))
		b.WriteString("\n")

		b.WriteString(resultLabelStyle.Render("Configuration:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render(fmt.Sprintf("%d-band", result.Reading.BandCount)))
		b.WriteString("\n\n")

		// Capacitance value
		b.WriteString(labelStyle.Render("CAPACITANCE VALUE:"))
		b.WriteString("\n")
		b.WriteString(resultLabelStyle.Render("Value:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render(FormatCapacitanceWithUF(result.CapacitanceValue, result.CapacitanceUnit, result.CapacitancePF)))
		b.WriteString("\n\n")

		// Tolerance
		b.WriteString(labelStyle.Render("TOLERANCE:"))
		b.WriteString("\n")
		b.WriteString(resultLabelStyle.Render("Specification:"))
		b.WriteString("  ")
		tolStr := FormatTolerance(result)
		if result.ToleranceType == "absolute" {
			tolStr += " (absolute, value ≤ 10pF)"
		} else {
			tolStr += " (percentage-based, value > 10pF)"
		}
		b.WriteString(resultValueStyle.Render(tolStr))
		b.WriteString("\n")

		b.WriteString(resultLabelStyle.Render("Range:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render(FormatToleranceRange(result)))
		b.WriteString("\n\n")

		// Voltage rating
		if result.VoltageValid {
			b.WriteString(labelStyle.Render("VOLTAGE RATING:"))
			b.WriteString("\n")
			b.WriteString(resultLabelStyle.Render("Voltage:"))
			b.WriteString("  ")
			b.WriteString(resultValueStyle.Render(FormatVoltage(result) + " (Type " + string(result.Reading.CapType) + " " + typeInfo.Name + ")"))
			b.WriteString("\n\n")
		}

		// Temperature coefficient
		if result.TempCoeffValid {
			b.WriteString(labelStyle.Render("TEMPERATURE COEFFICIENT:"))
			b.WriteString("\n")
			b.WriteString(resultLabelStyle.Render("Coefficient:"))
			b.WriteString("  ")
			b.WriteString(resultValueStyle.Render(FormatTempCoefficient(result)))
			b.WriteString("\n\n")
		}
	} else if m.componentType == ComponentResistor && m.resistorResult != nil {
		result := m.resistorResult

		// Type and configuration
		b.WriteString(resultLabelStyle.Render("Component Type:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render("Resistor"))
		b.WriteString("\n")

		b.WriteString(resultLabelStyle.Render("Configuration:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render(fmt.Sprintf("%d-band", result.Reading.BandCount)))
		b.WriteString("\n\n")

		// Resistance value
		b.WriteString(labelStyle.Render("RESISTANCE VALUE:"))
		b.WriteString("\n")
		b.WriteString(resultLabelStyle.Render("Value:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render(FormatResistance(result.ResistanceValue, result.ResistanceUnit)))
		b.WriteString("\n\n")

		// Tolerance
		b.WriteString(labelStyle.Render("TOLERANCE:"))
		b.WriteString("\n")
		b.WriteString(resultLabelStyle.Render("Specification:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render(FormatResistorTolerance(result)))
		b.WriteString("\n")

		b.WriteString(resultLabelStyle.Render("Range:"))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render(FormatResistorToleranceRange(result)))
		b.WriteString("\n\n")

		// Temperature coefficient (6-band only)
		if result.TempCoeffValid {
			b.WriteString(labelStyle.Render("TEMPERATURE COEFFICIENT:"))
			b.WriteString("\n")
			b.WriteString(resultLabelStyle.Render("Coefficient:"))
			b.WriteString("  ")
			b.WriteString(resultValueStyle.Render(FormatResistorTempCoefficient(result)))
			b.WriteString("\n\n")
		}
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
	b.WriteString(mutedStyle.Render(fmt.Sprintf("Decoded components in history: %d", len(m.history))))
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

	var bandCount int
	// Show all bands
	if m.componentType == ComponentCapacitor {
		bandCount = m.capacitorReading.BandCount
		for i := 1; i <= bandCount; i++ {
			var color Color
			switch i {
			case 1:
				color = m.capacitorReading.Band1
			case 2:
				color = m.capacitorReading.Band2
			case 3:
				color = m.capacitorReading.Band3
			case 4:
				color = m.capacitorReading.Band4
			case 5:
				color = m.capacitorReading.Band5
			}

			b.WriteString(valueStyle.Render(fmt.Sprintf("  %d = ", i)))
			b.WriteString(RenderColorBand(color, i))
			b.WriteString("\n")
		}
	} else {
		bandCount = m.resistorReading.BandCount
		for i := 1; i <= bandCount; i++ {
			var color Color
			switch i {
			case 1:
				color = m.resistorReading.Band1
			case 2:
				color = m.resistorReading.Band2
			case 3:
				color = m.resistorReading.Band3
			case 4:
				color = m.resistorReading.Band4
			case 5:
				color = m.resistorReading.Band5
			case 6:
				color = m.resistorReading.Band6
			}

			b.WriteString(valueStyle.Render(fmt.Sprintf("  %d = ", i)))
			b.WriteString(RenderColorBand(color, i))
			b.WriteString("\n")
		}
	}

	b.WriteString(valueStyle.Render("  Q = Cancel (keep current values)"))
	b.WriteString("\n\n")

	b.WriteString(promptStyle.Render(fmt.Sprintf("Select band (1-%d/Q): ", bandCount)))
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

	componentName := "capacitor"
	if m.componentType == ComponentResistor {
		componentName = "resistor"
	}
	b.WriteString(valueStyle.Render(fmt.Sprintf("Add a note to this %s reading (max 200 characters):", componentName)))
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
