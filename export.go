package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

// ExportToCSV exports the capacitor history to a CSV file
func ExportToCSV(history []CapacitorEntry, filename string) error {
	if len(history) == 0 {
		return fmt.Errorf("no capacitor data to export")
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
		"Type",
		"Band Count",
		"Band 1",
		"Band 2",
		"Band 3",
		"Band 4",
		"Band 5",
		"Capacitance (pF)",
		"Capacitance (Value)",
		"Capacitance (Unit)",
		"Capacitance (µF)",
		"Tolerance Type",
		"Tolerance (%)",
		"Tolerance (pF)",
		"Min Value",
		"Max Value",
		"Voltage (V)",
		"Temp Coefficient",
		"Note",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write each capacitor entry
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	for _, entry := range history {
		if entry.Result == nil {
			continue
		}

		result := entry.Result

		// Get color names for bands
		band1Name := GetColorInfo(result.Reading.Band1).Name
		band2Name := GetColorInfo(result.Reading.Band2).Name
		band3Name := GetColorInfo(result.Reading.Band3).Name
		band4Name := GetColorInfo(result.Reading.Band4).Name
		band5Name := ""
		if result.Reading.BandCount >= 4 {
			band5Name = GetColorInfo(result.Reading.Band5).Name
		}

		// Calculate µF value
		ufValue := result.CapacitancePF / 1000000.0

		// Format tolerance
		tolerancePercent := ""
		tolerancePF := ""
		if result.ToleranceType == "percentage" {
			tolerancePercent = fmt.Sprintf("%.1f", result.TolerancePercent)
		} else {
			tolerancePF = fmt.Sprintf("%.2f", result.ToleranceAbsolutePF)
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

		record := []string{
			timestamp,
			string(result.Reading.CapType),
			fmt.Sprintf("%d", result.Reading.BandCount),
			band1Name,
			band2Name,
			band3Name,
			band4Name,
			band5Name,
			fmt.Sprintf("%.3f", result.CapacitancePF),
			fmt.Sprintf("%.3f", result.CapacitanceValue),
			result.CapacitanceUnit,
			fmt.Sprintf("%.9f", ufValue),
			result.ToleranceType,
			tolerancePercent,
			tolerancePF,
			minVal,
			maxVal,
			voltage,
			tempCoeff,
			entry.Note,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}
