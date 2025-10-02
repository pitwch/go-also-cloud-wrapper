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
	"time"
)

// const Version shows the version of the wrapper
const (
	Version = "1.0.0"
)

// DefaultHTTPTransport is an http.RoundTripper that has DisableKeepAlives set true.
var DefaultHTTPTransport = &http.Transport{
	DisableKeepAlives: true,
}

// DefaultHTTPClient is an http.Client with the DefaultHTTPTransport and (Cookie) Jar set nil.
var DefaultHTTPClient = &http.Client{
	Jar:       nil,
	Transport: DefaultHTTPTransport,
}

// Client represents the struct for building the client
type Client struct {
	marketplace  *url.URL
	username     string
	password     string
	option       *Options
	client       *http.Client
	sessiontoken *string
}

// loginStruct represents the struct  for building the login request
type loginStruct struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewClient creates a new instance of client
func NewClient(marketplace string, apiUser string, apiPassword string, options *Options) (*Client, error) {

	// Check URL, else exit
	mpurl, err := url.ParseRequestURI(marketplace)
	if err != nil {
		return nil, fmt.Errorf("URL in wrong format: %s", err)
	}

	if options == nil {
		options = &Options{}
	}

	if options.APIPrefix == "" {
		options.APIPrefix = "/SimpleAPI/SimpleAPIService.svc/rest"
	}

	if options.LoginEndpoint == "" {
		options.LoginEndpoint = "GetSessionToken"
	}

	if options.UserAgent == "" {
		options.UserAgent = "go-also-cloud-wrapper " + Version
	}

	//Set default batchsize for batch requests
	if options.Batchsize == 0 {
		options.Batchsize = 200
	}

	// Configure HTTP client with timeout
	httpClient := DefaultHTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	// Apply timeout if specified
	if options.Timeout > 0 {
		httpClient = &http.Client{
			Timeout:   options.Timeout * time.Second,
			Transport: DefaultHTTPTransport,
		}
	}

	path := options.APIPrefix + "/"
	mpurl.Path = path

	return &Client{
		marketplace:  mpurl,
		username:     apiUser,
		password:     apiPassword,
		option:       options,
		client:       httpClient,
		sessiontoken: nil,
	}, nil
}

// createNewToken creates a new session token
func (c *Client) createNewToken(ctx context.Context) (token string, err error) {

	// Build Login URL
	urlstr := c.marketplace.String() + c.option.LoginEndpoint
	logDebug(ctx, c, fmt.Sprintf("Login on %v with user %v", urlstr, c.username))

	// Create Login Body
	data := loginStruct{c.username, c.password}

	// Encode Login Body to JSON
	body := new(bytes.Buffer)
	encoder := json.NewEncoder(body)
	if err := encoder.Encode(data); err != nil {
		return "", err
	}

	// Build Login Request
	req, err := http.NewRequest("POST", urlstr, body)
	if err != nil {
		return "", err
	}

	// Set Login Header
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	// If Response gives errors print also Body
	if resp.StatusCode == http.StatusBadRequest ||
		resp.StatusCode == http.StatusNotFound ||
		resp.StatusCode == http.StatusMethodNotAllowed ||
		resp.StatusCode == http.StatusInternalServerError {

		return "", errorFormatterPx(ctx, c, resp.StatusCode, resp.Body)
	}

	// Read response body
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", fmt.Errorf("failed to read login response: %w", err)
	}
	tk := buf.String()

	//Trim Token (remove quotes)
	tk = strings.Trim(tk, `"`)
	//If everything OK  return nil
	return tk, nil

}

// Login generates a new session token.
// Its a helper Function for Login to ALSO Cloud REST-API.
// Usually login is automatically done in context.
func (c *Client) login(ctx context.Context) error {

	// If sessiontoken doesnt yet exists create a new one
	if c.sessiontoken == nil || *c.sessiontoken == "" {
		tk, err := c.createNewToken(ctx)

		if tk != "" {
			c.sessiontoken = &tk
		}

		return err
		// If sessiontoken already exists return stored value
	}
	return nil
}

// GetSessionToken returns the current session token or creates a new one if needed.
// This is useful for debugging or when you need to inspect the token.
func (c *Client) GetSessionToken(ctx context.Context) (string, error) {
	if err := c.login(ctx); err != nil {
		return "", err
	}
	if c.sessiontoken == nil {
		return "", fmt.Errorf("session token is nil after login")
	}
	return *c.sessiontoken, nil
}

// SetSessionToken manually sets a session token.
// This is useful for reusing an existing valid session token.
func (c *Client) SetSessionToken(token string) {
	c.sessiontoken = &token
}

// ClearSessionToken clears the current session token, forcing a new login on next request.
func (c *Client) ClearSessionToken() {
	c.sessiontoken = nil
}

// request does basic requests (POST/PUT/PATCH/GET/DELETE).
// Building the Request Method for Client
// Includes Params for possible further use
func (c *Client) request(ctx context.Context, method, endpoint string, params url.Values, data interface{}) (io.ReadCloser, http.Header, int, error) {

	var urlstr string

	if params.Encode() != "" {
		urlstr = c.marketplace.String() + endpoint + "?" + params.Encode()

	} else {
		urlstr = c.marketplace.String() + endpoint
	}

	//If Log enabled log URL
	if c.sessiontoken != nil {
		logDebug(ctx, c, fmt.Sprintf("Request-URL: %v, Method: %v, Token: %v", urlstr, method, *c.sessiontoken))
	} else {
		logDebug(ctx, c, fmt.Sprintf("Request-URL: %v, Method: %v, Token: none", urlstr, method))
	}

	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
	case http.MethodDelete, http.MethodGet, http.MethodOptions:
	default:
		return nil, nil, 0, fmt.Errorf("method is not recognised: %s", method)
	}

	body := new(bytes.Buffer)

	encoder := json.NewEncoder(body)

	if err := encoder.Encode(data); err != nil {
		return nil, nil, 0, fmt.Errorf("JSON Encoding failed: %s", err)
	}

	req, err := http.NewRequest(method, urlstr, body)
	req = req.WithContext(ctx)

	if err != nil {
		return nil, nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication header only if session token exists
	if c.sessiontoken != nil && *c.sessiontoken != "" {
		token := "CCPSessionId " + *c.sessiontoken
		req.Header.Set("Authenticate", token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, 0, err
	}

	return resp.Body, resp.Header, resp.StatusCode, nil

}

// Post does a POST request to API.
// It Accepts Context, Endpoint and Data as Input.
// Returns io.ReadCloser,http.Header,Statuscode,error.
func (c *Client) Post(ctx context.Context, endpoint string, data interface{}) (io.ReadCloser, http.Header, int, error) {
	err := c.login(ctx)
	if err != nil {
		return nil, nil, 0, err
	}

	//Fix for sending empty json body if data is nil
	if data == nil {
		data = json.RawMessage(`{}`)

	}
	request, header, statuscode, err := c.request(ctx, "POST", endpoint, url.Values{}, data)

	//If Log enabled in options log data
	logDebug(ctx, c, fmt.Sprintf("Sent data in POST-Request: %v", data))

	//If error return nicely formatted error
	if err != nil {
		return request, header, statuscode, errorFormatterPx(ctx, c, statuscode, request)
		//If no error return nil
	}
	return request, header, statuscode, nil

}

// Validate validates the session and checks if session is still active.
// If still active returns true else returns false
func (c *Client) Validate(ctx context.Context) bool {
	val, _, _, err := c.Post(ctx, "PingPong", nil)
	if err != nil {
		logDebug(ctx, c, fmt.Sprintf("Validation failed: %v", err))
		return false
	}
	defer val.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(val); err != nil {
		logDebug(ctx, c, fmt.Sprintf("Failed to read validation response: %v", err))
		return false
	}
	valbyte := buf.String()

	logDebug(ctx, c, fmt.Sprintf("Validation response: %s", valbyte))
	return valbyte == "null"
}

// errorFormatterPX prettifies the XML-errors from API
// Returns Error, Statuscode and detailed error message
func errorFormatterPx(ctx context.Context, c *Client, statuscode int, request io.Reader) (err error) {
	if request == nil {
		return fmt.Errorf("API error: statuscode %d, no response body", statuscode)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(request); err != nil {
		return fmt.Errorf("failed to read error response: %w", err)
	}
	errbyte := buf.Bytes()

	//Define Error Struct
	parsedError := cloudfault{}

	//Try to parse XML ErrorStruct
	if err := xml.Unmarshal(errbyte, &parsedError); err != nil {
		logDebug(ctx, c, fmt.Sprintf("Failed to parse error XML: %v", err))
		// Return raw error body if parsing fails
		return fmt.Errorf("API error: statuscode %d, body: %s", statuscode, string(errbyte))
	}

	// Build error message from parsed XML
	reason := ""
	if parsedError.Reason != nil && parsedError.Reason.CText != nil {
		reason = parsedError.Reason.CText.String
	}
	message := ""
	if parsedError.Detail != nil && parsedError.Detail.ServiceException != nil && parsedError.Detail.ServiceException.Message != nil {
		message = parsedError.Detail.ServiceException.Message.String
	}

	return fmt.Errorf("API error: reason=%s, statuscode=%d, message=%s", reason, statuscode, message)
}

// logDebug does check if Log is enabled in Options or not
// ctx parameter reserved for future context-aware logging enhancements
func logDebug(_ context.Context, c *Client, logtext string) {
	//If Log enabled in options
	if c.option.Log {
		log.Printf("[go-also-cloud-wrapper] %s", logtext)
	}
}
