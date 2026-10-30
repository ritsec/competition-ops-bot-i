package utils

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"

	"github.com/gocarina/gocsv"
)

// maxCSVSize limits how much of an uploaded CSV is read into memory
const maxCSVSize = 10 << 20 // 10 MiB

// downloadCSV fetches the CSV at URL and returns its contents
func downloadCSV(URL string) ([]byte, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting file: %s", resp.Status)
	}

	return io.ReadAll(io.LimitReader(resp.Body, maxCSVSize))
}

// HandleCSV takes a generic struct with CSV tags and a URL to the
// CSV download, and returns the unmarshalled data.
func HandleCSV[T any](URL string) ([]*T, error) {
	content, err := downloadCSV(URL)
	if err != nil {
		return nil, err
	}

	// Unmarshal file based on Entry struct
	var data []*T
	if err := gocsv.UnmarshalBytes(content, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// ReadCSV returns the raw rows of the CSV at URL. Use this instead of
// HandleCSV when the columns can't be matched by header name.
func ReadCSV(URL string) ([][]string, error) {
	content, err := downloadCSV(URL)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(content))
	reader.FieldsPerRecord = -1 // Allow rows of different lengths
	return reader.ReadAll()
}
