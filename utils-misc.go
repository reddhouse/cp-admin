package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Returns a custom timestamp (cts) for time.Now() as Day/HH:MM:SS
func cts() string {
	t := time.Now()
	return fmt.Sprintf("%02d/%02d%02d%02d", t.Day(), t.Hour(), t.Minute(), t.Second())
}

// Decodes and unmarshals the JSON response body into the provided destination,
// or fatally exit upon error.
func unmarshalOrExit(body io.Reader, dst interface{}) {
	// User a decoder to read from the response body, and unmarshal the JSON.
	decoder := json.NewDecoder(body)
	// Ensure JSON doesn't contain unexpected fields (not present responseBody).
	decoder.DisallowUnknownFields()
	err := decoder.Decode(dst)
	if err != nil {
		fmt.Printf("[err][admin] decoding & unmarshaling JSON: %v [%s]\n", err, cts())
		os.Exit(1)
	}
}
