package alsocloud

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Version is the library version reported in the default User-Agent.
const Version = "1.1.0"

// DefaultHTTPTransport disables keep-alives to avoid stale connections
// against marketplace endpoints that close idle sockets aggressively.
var DefaultHTTPTransport = &http.Transport{
	DisableKeepAlives: true,
}

// DefaultHTTPClient is used when Options.HTTPClient is nil and no timeout is set.
var DefaultHTTPClient = &http.Client{
	Jar:       nil,
	Transport: DefaultHTTPTransport,
}

// Client is the ALSO Cloud Marketplace API client.
type Client struct {
	marketplace  *url.URL
	username     string
	password     string
	option       *Options
	client       *http.Client
	mu           sync.RWMutex
	sessiontoken string
}

type loginStruct struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewClient creates a new ALSO Cloud API client.
func NewClient(marketplace string, apiUser string, apiPassword string, options *Options) (*Client, error) {
	mpurl, err := url.ParseRequestURI(marketplace)
	if err != nil {
		return nil, fmt.Errorf("URL in wrong format: %w", err)
	}

	if options == nil {
		options = &Options{}
	}
	applyDefaults(options)

	mpurl.Path = options.APIPrefix + "/"

	return &Client{
		marketplace: mpurl,
		username:    apiUser,
		password:    apiPassword,
		option:      options,
		client:      buildHTTPClient(options),
	}, nil
}

func applyDefaults(options *Options) {
	if options.APIPrefix == "" {
		options.APIPrefix = "/SimpleAPI/SimpleAPIService.svc/rest"
	}
	if options.LoginEndpoint == "" {
		options.LoginEndpoint = "GetSessionToken"
	}
	if options.UserAgent == "" {
		options.UserAgent = "go-also-cloud-wrapper " + Version
	}
	if options.Batchsize == 0 {
		options.Batchsize = 200
	}
}

func buildHTTPClient(options *Options) *http.Client {
	if options.HTTPClient != nil {
		return options.HTTPClient
	}

	if timeout := normalizeTimeout(options.Timeout); timeout > 0 {
		return &http.Client{
			Timeout:   timeout,
			Transport: DefaultHTTPTransport,
		}
	}

	if DefaultHTTPClient != nil {
		return DefaultHTTPClient
	}
	return http.DefaultClient
}

// normalizeTimeout accepts both modern durations (30*time.Second) and the
// legacy "raw seconds" form (Timeout: 30) used in earlier releases.
func normalizeTimeout(d time.Duration) time.Duration {
	if d > 0 && d < time.Millisecond {
		return d * time.Second
	}
	return d
}

// createNewToken obtains a new session token from the login endpoint.
func (c *Client) createNewToken(ctx context.Context) (string, error) {
	urlstr := c.marketplace.String() + c.option.LoginEndpoint
	logDebug(c, fmt.Sprintf("Login on %s with user %s", urlstr, c.username))

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(loginStruct{c.username, c.password}); err != nil {
		return "", fmt.Errorf("failed to encode login body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlstr, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.option.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}
	defer closeBody(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errorFormatterPx(c, resp.StatusCode, resp.Body)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", fmt.Errorf("failed to read login response: %w", err)
	}

	return strings.Trim(buf.String(), `"`), nil
}

// login ensures a session token is present, creating one if needed.
func (c *Client) login(ctx context.Context) error {
	c.mu.RLock()
	hasToken := c.sessiontoken != ""
	c.mu.RUnlock()
	if hasToken {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock.
	if c.sessiontoken != "" {
		return nil
	}

	tk, err := c.createNewToken(ctx)
	if err != nil {
		return err
	}
	c.sessiontoken = tk
	return nil
}

// GetSessionToken returns the current session token, creating one if needed.
func (c *Client) GetSessionToken(ctx context.Context) (string, error) {
	if err := c.login(ctx); err != nil {
		return "", err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.sessiontoken == "" {
		return "", fmt.Errorf("session token is empty after login")
	}
	return c.sessiontoken, nil
}

// SetSessionToken manually sets a session token for reuse.
func (c *Client) SetSessionToken(token string) {
	c.mu.Lock()
	c.sessiontoken = token
	c.mu.Unlock()
}

// ClearSessionToken clears the session token, forcing a new login on the next request.
func (c *Client) ClearSessionToken() {
	c.mu.Lock()
	c.sessiontoken = ""
	c.mu.Unlock()
}

// request performs an HTTP request against the marketplace API.
func (c *Client) request(ctx context.Context, method, endpoint string, params url.Values, data interface{}) (io.ReadCloser, http.Header, int, error) {
	urlstr := c.marketplace.String() + endpoint
	if encoded := params.Encode(); encoded != "" {
		urlstr += "?" + encoded
	}

	c.mu.RLock()
	token := c.sessiontoken
	c.mu.RUnlock()

	logDebug(c, fmt.Sprintf("Request-URL: %s, Method: %s, Token: %s", urlstr, method, maskToken(token)))

	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodGet, http.MethodOptions:
	default:
		return nil, nil, 0, fmt.Errorf("method is not recognised: %s", method)
	}

	var bodyReader io.Reader
	// Avoid sending a JSON body for GET/OPTIONS, and when data is nil.
	if data != nil && method != http.MethodGet && method != http.MethodOptions {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(data); err != nil {
			return nil, nil, 0, fmt.Errorf("JSON encoding failed: %w", err)
		}
		bodyReader = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, urlstr, bodyReader)
	if err != nil {
		return nil, nil, 0, err
	}

	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", c.option.UserAgent)

	if token != "" {
		req.Header.Set("Authenticate", "CCPSessionId "+token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, 0, err
	}

	return resp.Body, resp.Header, resp.StatusCode, nil
}

// doAuthenticatedRequest logs in, performs the request, and maps API error responses.
func (c *Client) doAuthenticatedRequest(ctx context.Context, method, endpoint string, params url.Values, data interface{}) (io.ReadCloser, http.Header, int, error) {
	if err := c.login(ctx); err != nil {
		return nil, nil, 0, err
	}

	body, header, statusCode, err := c.request(ctx, method, endpoint, params, data)
	if err != nil {
		return nil, nil, 0, err
	}

	if statusCode < 200 || statusCode >= 300 {
		apiErr := errorFormatterPx(c, statusCode, body)
		closeBody(body)
		return nil, header, statusCode, apiErr
	}

	return body, header, statusCode, nil
}

// Post sends a POST request to the API.
func (c *Client) Post(ctx context.Context, endpoint string, data interface{}) (io.ReadCloser, http.Header, int, error) {
	if data == nil {
		data = json.RawMessage(`{}`)
	}
	logDebug(c, fmt.Sprintf("Sent data in POST-Request: %v", data))
	return c.doAuthenticatedRequest(ctx, http.MethodPost, endpoint, url.Values{}, data)
}

// Validate checks whether the current session is still active.
func (c *Client) Validate(ctx context.Context) bool {
	val, _, _, err := c.Post(ctx, "PingPong", nil)
	if err != nil {
		logDebug(c, fmt.Sprintf("Validation failed: %v", err))
		return false
	}
	defer closeBody(val)

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(val); err != nil {
		logDebug(c, fmt.Sprintf("Failed to read validation response: %v", err))
		return false
	}

	valbyte := buf.String()
	logDebug(c, fmt.Sprintf("Validation response: %s", valbyte))
	return valbyte == "null"
}

// errorFormatterPx prettifies XML fault responses from the API.
func errorFormatterPx(c *Client, statuscode int, request io.Reader) error {
	if request == nil {
		return fmt.Errorf("API error: statuscode %d, no response body", statuscode)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(request); err != nil {
		return fmt.Errorf("failed to read error response: %w", err)
	}
	errbyte := buf.Bytes()

	parsedError := cloudfault{}
	if err := xml.Unmarshal(errbyte, &parsedError); err != nil {
		logDebug(c, fmt.Sprintf("Failed to parse error XML: %v", err))
		return fmt.Errorf("API error: statuscode %d, body: %s", statuscode, string(errbyte))
	}

	reason := ""
	if parsedError.Reason != nil && parsedError.Reason.CText != nil {
		reason = parsedError.Reason.CText.String
	}
	message := ""
	if parsedError.Detail != nil &&
		parsedError.Detail.ServiceException != nil &&
		parsedError.Detail.ServiceException.Message != nil {
		message = parsedError.Detail.ServiceException.Message.String
	}

	return fmt.Errorf("API error: reason=%s, statuscode=%d, message=%s", reason, statuscode, message)
}

func logDebug(c *Client, logtext string) {
	if c.option != nil && c.option.Log {
		log.Printf("[go-also-cloud-wrapper] %s", logtext)
	}
}

func closeBody(body io.Closer) {
	if body == nil {
		return
	}
	if err := body.Close(); err != nil {
		log.Printf("[go-also-cloud-wrapper] failed to close response body: %v", err)
	}
}

func maskToken(token string) string {
	if token == "" {
		return "none"
	}
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "…" + token[len(token)-4:]
}
