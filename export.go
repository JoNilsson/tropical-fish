package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

// ComponentEntry represents a decoded component (capacitor or resistor) with notes
type ComponentEntry struct {
	ComponentType   ComponentType
	CapacitorResult *CalculationResult
	ResistorResult  *ResistorResult
	Note            string
}

// ExportToCSV exports the component history to a CSV file
func ExportToCSV(history []ComponentEntry, filename string) error {
	if len(history) == 0 {
		return fmt.Errorf("no component data to export")
	}

	// Create or overwrite the CSV file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{
		"Timestamp",
		"Component Type",
		"Cap Type",
		"Band Count",
		"Band 1",
		"Band 2",
		"Band 3",
		"Band 4",
		"Band 5",
		"Band 6",
		"Value",
		"Unit",
		"Tolerance (%)",
		"Min Value",
		"Max Value",
		"Voltage (V)",
		"Temp Coefficient",
		"Note",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write each component entry
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	for _, entry := range history {
		var record []string

		if entry.ComponentType == ComponentCapacitor && entry.CapacitorResult != nil {
			result := entry.CapacitorResult

			// Get color names for bands
			band1Name := GetColorInfo(result.Reading.Band1).Name
			band2Name := GetColorInfo(result.Reading.Band2).Name
			band3Name := GetColorInfo(result.Reading.Band3).Name
			band4Name := GetColorInfo(result.Reading.Band4).Name
			band5Name := ""
			if result.Reading.BandCount >= 4 {
				band5Name = GetColorInfo(result.Reading.Band5).Name
			}

			// Format tolerance
			tolerancePercent := ""
			if result.ToleranceType == "percentage" {
				tolerancePercent = fmt.Sprintf("%.1f", result.TolerancePercent)
			}

			// Format voltage
			voltage := ""
			if result.VoltageValid {
				voltage = fmt.Sprintf("%.1f", result.VoltageRating)
			}

			// Format temperature coefficient
			tempCoeff := ""
			if result.TempCoeffValid {
				tempCoeff = fmt.Sprintf("%d", result.TempCoefficient)
			}

			// Format min/max values
			minVal := FormatCapacitance(result.MinValue, result.MinUnit)
			maxVal := FormatCapacitance(result.MaxValue, result.MaxUnit)

			record = []string{
				timestamp,
				"Capacitor",
				string(result.Reading.CapType),
				fmt.Sprintf("%d", result.Reading.BandCount),
				band1Name,
				band2Name,
				band3Name,
				band4Name,
				band5Name,
				"",
				fmt.Sprintf("%.3f", result.CapacitanceValue),
				result.CapacitanceUnit,
				tolerancePercent,
				minVal,
				maxVal,
				voltage,
				tempCoeff,
				entry.Note,
			}

		} else if entry.ComponentType == ComponentResistor && entry.ResistorResult != nil {
			result := entry.ResistorResult

			// Get color names for bands
			band1Name := GetColorInfo(result.Reading.Band1).Name
			band2Name := GetColorInfo(result.Reading.Band2).Name
			band3Name := GetColorInfo(result.Reading.Band3).Name
			band4Name := GetColorInfo(result.Reading.Band4).Name
			band5Name := ""
			if result.Reading.BandCount >= 5 {
				band5Name = GetColorInfo(result.Reading.Band5).Name
			}
			band6Name := ""
			if result.Reading.BandCount == 6 {
				band6Name = GetColorInfo(result.Reading.Band6).Name
			}

			// Format tolerance
			tolerancePercent := fmt.Sprintf("%.2f", result.TolerancePercent)

			// Format temperature coefficient
			tempCoeff := ""
			if result.TempCoeffValid {
				tempCoeff = fmt.Sprintf("%d ppm/°C", result.TempCoefficient)
			}

			// Format min/max values
			minVal := FormatResistance(result.MinValue, result.MinUnit)
			maxVal := FormatResistance(result.MaxValue, result.MaxUnit)

			record = []string{
				timestamp,
				"Resistor",
				"",
				fmt.Sprintf("%d", result.Reading.BandCount),
				band1Name,
				band2Name,
				band3Name,
				band4Name,
				band5Name,
				band6Name,
				fmt.Sprintf("%.3f", result.ResistanceValue),
				result.ResistanceUnit,
				tolerancePercent,
				minVal,
				maxVal,
				"",
				tempCoeff,
				entry.Note,
			}
		} else {
			continue
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}
