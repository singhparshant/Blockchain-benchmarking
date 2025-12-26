package client

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"fmt"
)

func InitializeCSVWriter(contractID string) (*csv.Writer, *os.File, error) {
	// Construct the file name using the contractID
	fileName := fmt.Sprintf("%s.csv", contractID)
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}

	// Create a directory to store the results, if it doesn't already exist
	resultsDir := filepath.Join(homeDir, "results")
	if _, err := os.Stat(resultsDir); os.IsNotExist(err) {
		if err := os.Mkdir(resultsDir, 0755); err != nil {
			return nil, nil, err
		}
	}

	// Create the CSV file
	file, err := os.Create(filepath.Join(resultsDir, fileName))
	if err != nil {
		return nil, nil, err
	}

	// Initialize the CSV writer
	writer := csv.NewWriter(file)

	// Write the header row to the CSV file
	header := []string{"Block_Height", "TxnHash", "GasFee", "GasBurnt", "SentAt", "FinalisedAt"}
	if err := writer.Write(header); err != nil {
		file.Close() // Attempt to close the file on error
		return nil, nil, err
	}

	// Return the CSV writer and file
	return writer, file, nil
}

func WriteToCSV(writer *csv.Writer, result *ExperimentResult) error {
	// Convert result fields to strings
	record := []string{
		strconv.FormatUint(result.Block_Height, 10),
		result.TxnHash,
		result.GasFee.String(), // Assuming GasFee has a String method for formatting
		strconv.FormatUint(result.GasBurnt, 10),
		strconv.FormatInt(result.SentAt, 10),
		strconv.FormatUint(result.FinalisedAt, 10),
	}

	// Write the record to the CSV file
	if err := writer.Write(record); err != nil {
		return err
	}

	// Flush the writer to ensure the record is written to the file
	writer.Flush()

	// Check for errors that may have occurred during flushing
	if err := writer.Error(); err != nil {
		return err
	}

	return nil
}
