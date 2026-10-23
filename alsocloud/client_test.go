package alsocloud

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

const testParentAccountID = 523525

type testCredentials struct {
	username string
	password string
}

func getTestCredentials(t *testing.T) testCredentials {
	t.Helper()
	return testCredentials{
		username: os.Getenv("ALSO_CLOUDUSER"),
		password: os.Getenv("ALSO_CLOUDPW"),
	}
}

func (tc testCredentials) hasCredentials() bool {
	return tc.username != "" && tc.password != ""
}

func setupTestClient(t *testing.T) *Client {
	t.Helper()

	creds := getTestCredentials(t)
	if !creds.hasCredentials() {
		t.Skip("Skipping test - ALSO_CLOUDUSER and ALSO_CLOUDPW environment variables not set")
	}

	client, err := NewClient(
		Debug,
		creds.username,
		creds.password,
		&Options{Timeout: 30 * time.Second},
	)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	return client
}

// newTestServerClient wires a Client to an httptest.Server for unit tests.
func newTestServerClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, "user@example.com", "secret", &Options{
		APIPrefix:     "/api",
		LoginEndpoint: "GetSessionToken",
		HTTPClient:    server.Client(),
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

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
			options:     &Options{Timeout: 1 * time.Second},
			wantErr:     false,
		},
		{
			name:        "valid_url_legacy_timeout_seconds",
			marketplace: Debug,
			apiUser:     "test@example.com",
			apiPassword: "password",
			options:     &Options{Timeout: 30}, // legacy: raw seconds
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
				return
			}
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if client == nil {
				t.Error("Expected client to be non-nil")
			}
		})
	}
}

func TestNormalizeTimeout(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want time.Duration
	}{
		{0, 0},
		{30, 30 * time.Second},               // legacy raw seconds
		{30 * time.Second, 30 * time.Second}, // modern duration
		{500 * time.Millisecond, 500 * time.Millisecond},
	}
	for _, tt := range tests {
		if got := normalizeTimeout(tt.in); got != tt.want {
			t.Errorf("normalizeTimeout(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestClient_LoginAndPost(t *testing.T) {
	var gotAuth string
	var gotUA string
	var gotMethod string
	var gotBody string

	client := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "GetSessionToken"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`"session-token-abc123"`))
		case strings.HasSuffix(r.URL.Path, "GetCompany"):
			gotAuth = r.Header.Get("Authenticate")
			gotUA = r.Header.Get("User-Agent")
			gotMethod = r.Method
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			http.NotFound(w, r)
		}
	})

	ctx := context.Background()
	resp, _, status, err := client.Post(ctx, "GetCompany", map[string]int{"parentAccountId": 1})
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	defer resp.Close()

	if status != http.StatusOK {
		t.Errorf("status = %d, want 200", status)
	}
	if gotAuth != "CCPSessionId session-token-abc123" {
		t.Errorf("Authenticate = %q", gotAuth)
	}
	if !strings.Contains(gotUA, "go-also-cloud-wrapper") {
		t.Errorf("User-Agent = %q, want wrapper UA", gotUA)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if !strings.Contains(gotBody, "parentAccountId") {
		t.Errorf("body = %q, want parentAccountId", gotBody)
	}
}

func TestClient_GetDoesNotSendBody(t *testing.T) {
	var gotBody []byte
	client := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "GetSessionToken"):
			_, _ = w.Write([]byte(`"tok"`))
		default:
			gotBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
		}
	})

	resp, _, _, err := client.Get(context.Background(), "List", url.Values{"q": {"x"}})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Close()

	if len(gotBody) != 0 {
		t.Errorf("GET body = %q, want empty", gotBody)
	}
}

func TestClient_APIErrorStatus(t *testing.T) {
	xmlFault := `<?xml version="1.0"?>
<Fault xmlns="http://schemas.microsoft.com/ws/2005/05/envelope/none">
  <Reason><Text xml:lang="en-US">Bad Request</Text></Reason>
  <Detail>
    <ServiceException xmlns="http://schemas.datacontract.org/2004/07/Nervogrid.Platform.API">
      <Message>invalid</Message>
    </ServiceException>
  </Detail>
</Fault>`

	client := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "GetSessionToken"):
			_, _ = w.Write([]byte(`"tok"`))
		default:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(xmlFault))
		}
	})

	resp, _, status, err := client.Post(context.Background(), "Broken", nil)
	if resp != nil {
		t.Error("expected nil body on API error")
	}
	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", status)
	}
	if err == nil {
		t.Fatal("expected API error")
	}
	if !strings.Contains(err.Error(), "Bad Request") && !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error = %v, want fault details", err)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	client := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, _, err := client.Post(ctx, "Slow", nil)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestClient_SessionTokenHelpers(t *testing.T) {
	client, err := NewClient(Debug, "u", "p", nil)
	if err != nil {
		t.Fatal(err)
	}

	client.SetSessionToken("manual-token")
	tok, err := client.GetSessionToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "manual-token" {
		t.Errorf("token = %q", tok)
	}

	client.ClearSessionToken()
	client.mu.RLock()
	empty := client.sessiontoken == ""
	client.mu.RUnlock()
	if !empty {
		t.Error("expected cleared token")
	}
}

func TestMaskToken(t *testing.T) {
	if got := maskToken(""); got != "none" {
		t.Errorf("empty = %q", got)
	}
	if got := maskToken("short"); got != "***" {
		t.Errorf("short = %q", got)
	}
	if got := maskToken("abcdefghijklmnop"); got != "abcd…mnop" {
		t.Errorf("long = %q", got)
	}
}

func TestHelpers(t *testing.T) {
	t.Run("ReadResponseBody", func(t *testing.T) {
		body := io.NopCloser(strings.NewReader(`{"a":1}`))
		s, err := ReadResponseBody(body)
		if err != nil {
			t.Fatal(err)
		}
		if s != `{"a":1}` {
			t.Errorf("got %q", s)
		}
	})

	t.Run("ReadResponseJSON", func(t *testing.T) {
		body := io.NopCloser(strings.NewReader(`{"a":1}`))
		var target map[string]int
		if err := ReadResponseJSON(body, &target); err != nil {
			t.Fatal(err)
		}
		if target["a"] != 1 {
			t.Errorf("got %#v", target)
		}
	})

	t.Run("IsValidSessionToken", func(t *testing.T) {
		tok := "x"
		if !IsValidSessionToken(&tok) {
			t.Error("expected valid")
		}
		if IsValidSessionToken(nil) {
			t.Error("expected invalid nil")
		}
	})
}

func TestClient_GetCompanies(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t)

	t.Run("post_get_company", func(t *testing.T) {
		data := map[string]interface{}{
			"parentAccountId": testParentAccountID,
		}

		resp, _, statuscode, err := client.Post(ctx, "GetCompany", data)
		if resp != nil {
			defer resp.Close()
		}

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if statuscode != 200 {
			t.Errorf("Expected HTTP status code 200, got %d", statuscode)
		}
	})

	t.Run("validate_session", func(t *testing.T) {
		if !client.Validate(ctx) {
			t.Error("Expected session to be valid")
		}
	})
}

func TestErrorFormatter(t *testing.T) {
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

		err = errorFormatterPx(client, 400, xmlError)
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "TEST OK") {
			t.Errorf("Expected TEST OK in error, got: %v", err)
		}
	})

	t.Run("nil_request", func(t *testing.T) {
		err := errorFormatterPx(client, 500, nil)
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
		client, err := NewClient(Debug, "test@example.com", "password", &Options{Timeout: time.Second})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if client.client.Timeout != time.Second {
			t.Errorf("Timeout = %v, want 1s", client.client.Timeout)
		}
	})

	t.Run("legacy_timeout_seconds", func(t *testing.T) {
		client, err := NewClient(Debug, "test@example.com", "password", &Options{Timeout: 60})
		if err != nil {
			t.Fatal(err)
		}
		if client.client.Timeout != 60*time.Second {
			t.Errorf("Timeout = %v, want 60s", client.client.Timeout)
		}
	})
}

func TestClient_SessionManagement(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t)

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
	})
}

func TestClient_CustomHTTPClient(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, _ = w.Write([]byte(`"tok"`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "u", "p", &Options{
		APIPrefix:     "/",
		LoginEndpoint: "login",
		HTTPClient:    server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.GetSessionToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("custom HTTP client was not used")
	}
}

func TestRequest_UnrecognizedMethod(t *testing.T) {
	client, err := NewClient(Debug, "u", "p", nil)
	if err != nil {
		t.Fatal(err)
	}
	client.SetSessionToken("tok")

	_, _, _, err = client.request(context.Background(), "TRACE", "x", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "not recognised") {
		t.Errorf("expected unrecognized method error, got %v", err)
	}
}

func TestPost_NilDataSendsEmptyObject(t *testing.T) {
	var raw json.RawMessage
	client := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "GetSessionToken"):
			_, _ = w.Write([]byte(`"tok"`))
		default:
			_ = json.NewDecoder(r.Body).Decode(&raw)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`null`))
		}
	})

	resp, _, _, err := client.Post(context.Background(), "PingPong", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()

	if string(raw) != "{}" {
		t.Errorf("body = %s, want {}", raw)
	}
}
