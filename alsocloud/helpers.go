package alsocloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// ReadResponseBody is a helper function to read and close response body
// It returns the response as a string and handles cleanup
func ReadResponseBody(body io.ReadCloser) (string, error) {
	if body == nil {
		return "", fmt.Errorf("response body is nil")
	}
	defer body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	return buf.String(), nil
}

// ReadResponseJSON is a helper function to read response body and unmarshal JSON
// It automatically closes the response body
func ReadResponseJSON(body io.ReadCloser, target interface{}) error {
	if body == nil {
		return fmt.Errorf("response body is nil")
	}
	defer body.Close()

	decoder := json.NewDecoder(body)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}
	return nil
}

// IsValidSessionToken checks if a session token is valid (not nil and not empty)
func IsValidSessionToken(token *string) bool {
	return token != nil && *token != ""
}
