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
func NewClient(RestURL string, apiUser string, apiPassword string, options *Options) (*Client, error) {
	restURL, err := url.Parse(RestURL)
	if err != nil {
		return nil, err
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

	if DefaultHTTPClient == nil {

		DefaultHTTPClient = http.DefaultClient
	}

	path := options.APIPrefix + "/"
	restURL.Path = path

	return &Client{
		marketplace:  restURL,
		username:     apiUser,
		password:     apiPassword,
		option:       options,
		client:       DefaultHTTPClient,
		sessiontoken: nil,
	}, nil
}

// createNewToken creates a new session token
func (c *Client) createNewToken(ctx context.Context) (token string, err error) {

	// Check URL, else exit
	_, err = url.ParseRequestURI(c.marketplace.String())
	if err != nil {
		return "", fmt.Errorf("URL in wrong format: %s", err)
	}

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
		return "", errorFormatterPx(ctx, c, resp.StatusCode, resp.Body)
	}

	// If Response gives errors print also Body
	if resp.StatusCode == http.StatusBadRequest ||
		resp.StatusCode == http.StatusNotFound ||
		resp.StatusCode == http.StatusMethodNotAllowed ||
		resp.StatusCode == http.StatusInternalServerError {

		return "", errorFormatterPx(ctx, c, resp.StatusCode, resp.Body)
	}

	respbody := resp.Body
	buf := new(bytes.Buffer)
	buf.ReadFrom(respbody)
	tk := buf.String()
	defer respbody.Close()

	//Trim Token
	tk = strings.Replace(tk, "\"", "", -1)
	//If everything OK  return nil
	return tk, nil

}

// Login generates a session token
// Helper Function for Login to ALSO Cloud REST-API
// Login is automatically done
func (c *Client) Login(ctx context.Context) error {

	// If sessiontoken doesnt yet exists create a new one
	if c.sessiontoken == nil || *c.sessiontoken == "" {
		tk, err := c.createNewToken(ctx)
		c.sessiontoken = &tk

		return err
		// If sessiontoken already exists return stored value
	}
	return nil
}

// request does basic requests (POST/PUT/PATCH/GET/DELETE)
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
	case http.MethodPost, http.MethodPut:
	case http.MethodDelete, http.MethodGet, http.MethodOptions:
	default:
		return nil, nil, 0, fmt.Errorf("Method is not recognised: %s", method)
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

	token := "CCPSessionId " + *c.sessiontoken
	req.Header.Set("Authenticate", token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, resp.StatusCode, err
	}

	return resp.Body, resp.Header, resp.StatusCode, err

}

// Post does a POST request to API
// Accepts Context, Endpoint and Data as Input
// Returns io.ReadCloser,http.Header,Statuscode,error
func (c *Client) Post(ctx context.Context, endpoint string, data interface{}) (io.ReadCloser, http.Header, int, error) {
	err := c.Login(ctx)
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

// errorFormatterPX prettifies the XML-errors from API
// Returns Error, Statuscode and detailed error message
func errorFormatterPx(ctx context.Context, c *Client, statuscode int, request io.Reader) (err error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(request)
	errbyte := buf.Bytes()

	//Define Error Struct
	parsedError := cloudfault{}

	//Try to parse JSON in XML ErrorStruct
	err = xml.Unmarshal(errbyte, &parsedError)

	if err != nil {
		fmt.Print(err)
	}
	return fmt.Errorf("Error: %v, Statuscode: %v, Message: %v", parsedError.Reason.CText, statuscode, parsedError.Detail.ServiceException.Message.String)
}

// logDebug does check if Log is enabled in Options or not
// Returns bool
func logDebug(ctx context.Context, c *Client, logtext string) {
	//If Log enabled in options
	if c.option.Log == true {
		log.Print(logtext)
	}
}
