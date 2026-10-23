package alsocloud

import (
	"net/http"
	"time"
)

// Options configures the ALSO Cloud client.
// All fields are optional; sensible defaults are applied in NewClient.
type Options struct {
	// APIPrefix is the path prefix for API endpoints.
	// Default: /SimpleAPI/SimpleAPIService.svc/rest
	APIPrefix string

	// LoginEndpoint is the endpoint used to obtain a session token.
	// Default: GetSessionToken
	LoginEndpoint string

	// UserAgent is sent with every HTTP request.
	// Default: "go-also-cloud-wrapper <Version>"
	UserAgent string

	// Timeout for the HTTP client.
	// Values smaller than one millisecond are treated as seconds for
	// backward compatibility (e.g. Timeout: 30 means 30 seconds).
	// Prefer explicit durations: Timeout: 30 * time.Second.
	Timeout time.Duration

	// VerifySSL is retained for API compatibility. TLS verification is
	// always enabled for the default client. To customize TLS (including
	// insecure skip-verify for local debugging), set HTTPClient.
	VerifySSL bool

	// Batchsize is reserved for future batch helpers.
	// Default: 200
	Batchsize int

	// Log enables debug logging via the standard library logger.
	Log bool

	// HTTPClient allows injecting a custom HTTP client (e.g. App Engine urlfetch).
	// When set, Timeout is not applied to that client.
	HTTPClient *http.Client
}
