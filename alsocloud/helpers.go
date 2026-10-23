package alsocloud

import (
	"encoding/json"
	"fmt"
	"io"
)

// ReadResponseBody reads and closes a response body, returning it as a string.
func ReadResponseBody(body io.ReadCloser) (string, error) {
	if body == nil {
		return "", fmt.Errorf("response body is nil")
	}
	defer closeBody(body)

	data, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	return string(data), nil
}

// ReadResponseJSON reads a response body, unmarshals JSON into target, and closes the body.
func ReadResponseJSON(body io.ReadCloser, target interface{}) error {
	if body == nil {
		return fmt.Errorf("response body is nil")
	}
	defer closeBody(body)

	if err := json.NewDecoder(body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}
	return nil
}

// IsValidSessionToken reports whether a session token pointer is non-nil and non-empty.
func IsValidSessionToken(token *string) bool {
	return token != nil && *token != ""
}
