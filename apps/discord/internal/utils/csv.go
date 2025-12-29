package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gocarina/gocsv"
)

// HandleCSV takes a generic struct with CSV tags and a URL to the
// CSV download, and returns the unmarshalled data.
func HandleCSV[T any](URL string) ([]*T, error) {
	// Create temp file
	filename := fmt.Sprintf("download-%s.csv", time.Now())
	temp, err := os.CreateTemp("/tmp", filename)
	if err != nil {
		return nil, err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()

	// Download file
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting file")
	}
	defer resp.Body.Close()

	// Copy file content to temp file
	_, err = io.Copy(temp, resp.Body)
	if err != nil {
		return nil, err
	}

	// Set offset to beginning of file
	if _, err := temp.Seek(0, 0); err != nil {
		return nil, err
	}

	// Unmarshal file based on Entry struct
	var data []*T
	if unmarshalError := gocsv.UnmarshalFile(temp, &data); unmarshalError != nil {
		panic(unmarshalError)
	}

	return data, nil
}
