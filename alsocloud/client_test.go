package alsocloud

import (
	"context"
	"os"
	"strings"
	"testing"
)

// Test fixtures and helper structures
const (
	testParentAccountID = 523525 // Default test account ID
)

// testCredentials holds the test credentials from environment
type testCredentials struct {
	username string
	password string
}

// getTestCredentials retrieves credentials from environment variables
func getTestCredentials(t *testing.T) testCredentials {
	t.Helper()
	return testCredentials{
		username: os.Getenv("ALSO_CLOUDUSER"),
		password: os.Getenv("ALSO_CLOUDPW"),
	}
}

// hasCredentials checks if test credentials are available
func (tc testCredentials) hasCredentials() bool {
	return tc.username != "" && tc.password != ""
}

// setupTestClient creates a test client with credentials from environment
func setupTestClient(t *testing.T, ctx context.Context) *Client {
	t.Helper()

	creds := getTestCredentials(t)
	if !creds.hasCredentials() {
		t.Skip("Skipping test - ALSO_CLOUDUSER and ALSO_CLOUDPW environment variables not set")
	}

	client, err := NewClient(
		Debug,
		creds.username,
		creds.password,
		&Options{Timeout: 30},
	)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	return client
}

// TestNewClient tests client creation with various inputs
func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		marketplace string
		apiUser     string
		apiPassword string
		options     *Options
		wantErr     bool
		errContains string
	}{
		{
			name:        "invalid_url_no_scheme",
			marketplace: "google.ch",
			apiUser:     "demouser",
			apiPassword: "1234",
			options:     nil,
			wantErr:     true,
			errContains: "URL in wrong format",
		},
		{
			name:        "valid_url_with_timeout",
			marketplace: Debug,
			apiUser:     "test@example.com",
			apiPassword: "password",
			options:     &Options{Timeout: 1},
			wantErr:     false,
		},
		{
			name:        "valid_url_no_options",
			marketplace: Debug,
			apiUser:     "test@example.com",
			apiPassword: "password",
			options:     nil,
			wantErr:     false,
		},
		{
			name:        "switzerland_marketplace",
			marketplace: Switzerland,
			apiUser:     "test@example.com",
			apiPassword: "password",
			options:     &Options{Log: true},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.marketplace, tt.apiUser, tt.apiPassword, tt.options)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain %q, got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if client == nil {
					t.Error("Expected client to be non-nil")
				}
			}
		})
	}
}
func TestClient_GetCompanies(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, ctx)

	t.Run("post_get_company", func(t *testing.T) {
		data := map[string]interface{}{
			"parentAccountId": testParentAccountID,
		}

		resp, _, statuscode, err := client.Post(ctx, "GetCompany", data)
		if resp != nil {
			defer resp.Close()
		}

		if statuscode != 200 {
			t.Errorf("Expected HTTP status code 200, got %d", statuscode)
		}

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("validate_session", func(t *testing.T) {
		valid := client.Validate(ctx)
		if !valid {
			t.Error("Expected session to be valid")
		}
	})
}

func TestErrorFormatter(t *testing.T) {
	ctx := context.Background()
	client, err := NewClient(Debug, "demo", "1234", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("xml_error_file", func(t *testing.T) {
		xmlError, err := os.Open("sampleError.xml")
		if err != nil {
			t.Skipf("Skipping test - sampleError.xml not found: %v", err)
		}
		defer xmlError.Close()

		err = errorFormatterPx(ctx, client, 400, xmlError)
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err != nil && err.Error() == "" {
			t.Error("Expected non-empty error message")
		}
	})

	t.Run("nil_request", func(t *testing.T) {
		err := errorFormatterPx(ctx, client, 500, nil)
		if err == nil {
			t.Error("Expected error for nil request, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "no response body") {
			t.Errorf("Expected error to mention 'no response body', got: %v", err)
		}
	})
}

func TestClient_Timeout(t *testing.T) {
	t.Run("create_client_with_timeout", func(t *testing.T) {
		client, err := NewClient(
			Debug,
			"test@example.com",
			"password",
			&Options{Timeout: 1},
		)

		if err != nil {
			t.Errorf("Expected no error creating client with timeout, got: %v", err)
		}
		if client == nil {
			t.Error("Expected client to be non-nil")
		}
	})

	t.Run("create_client_with_long_timeout", func(t *testing.T) {
		client, err := NewClient(
			Debug,
			"test@example.com",
			"password",
			&Options{Timeout: 60},
		)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if client == nil {
			t.Error("Expected client to be non-nil")
		}
	})
}

func TestClient_SessionManagement(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, ctx)

	t.Run("get_session_token", func(t *testing.T) {
		token, err := client.GetSessionToken(ctx)
		if err != nil {
			t.Errorf("Expected no error getting session token, got: %v", err)
		}
		if token == "" {
			t.Error("Expected non-empty session token")
		}
	})

	t.Run("set_and_clear_session_token", func(t *testing.T) {
		testToken := "test-token-12345"
		client.SetSessionToken(testToken)

		token, err := client.GetSessionToken(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if token != testToken {
			t.Errorf("Expected token %q, got %q", testToken, token)
		}

		client.ClearSessionToken()
		// After clearing, getting token should create a new one (requires valid creds)
	})
}
