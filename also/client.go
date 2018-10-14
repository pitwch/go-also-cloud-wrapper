package also

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

//Version of Wrapper
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

//Client Struct
type Client struct {
	restURL      *url.URL
	Username     string
	Password     string
	option       *Options
	client       *http.Client
	sessiontoken *string
}

//Login Struct
type LoginStruct struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//Error Struct
type ErrorStruct struct {
	Type    string `json:"Type"`
	Message string `json:"Message"`
	Fields  []struct {
		Reason  string `json:"Reason"`
		Name    string `json:"Name"`
		Message string `json:"Message"`
	} `json:"Fields"`
}

//Building new Client
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
		restURL:      restURL,
		Username:     apiUser,
		Password:     apiPassword,
		option:       options,
		client:       DefaultHTTPClient,
		sessiontoken: nil,
	}, nil
}

//Function for creating new sessiontoken
func (c *Client) createNewToken(ctx context.Context) (token string, err error) {

	//Check URL, else exit
	_, err = url.ParseRequestURI(c.restURL.String())
	if err != nil {
		return "", fmt.Errorf("URL in wrong format: %s", err)
	}

	//Build Login URL
	urlstr := c.restURL.String() + c.option.LoginEndpoint
	logDebug(ctx, c, fmt.Sprintf("Login on %v with user %v", urlstr, c.Username))
	//Create Login Body
	data := LoginStruct{c.Username, c.Password}

	//Encode Login Body to JSON
	body := new(bytes.Buffer)
	encoder := json.NewEncoder(body)
	if err := encoder.Encode(data); err != nil {
		return "", err
	}

	//Build Login Request
	req, err := http.NewRequest("POST", urlstr, body)
	if err != nil {
		return "", err
	}

	//Set Login Header
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	//If Response gives errors print also Body
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

//Function for Login
//Helper Function for Login to PROFFIX REST-API
//Login is automatically done
func (c *Client) Login(ctx context.Context) error {

	// If sessiontoken doesnt yet exists create a new one
	if c.sessiontoken == nil || *c.sessiontoken == "" {
		tk, err := c.createNewToken(ctx)
		c.sessiontoken = &tk

		return err
		// If sessiontoken already exists return stored value
	} else {
		return nil
	}
}

//Request Method
//Building the Request Method for Client
func (c *Client) request(ctx context.Context, method, endpoint string, params url.Values, data interface{}) (io.ReadCloser, http.Header, int, error) {

	var urlstr string

	if params.Encode() != "" {
		urlstr = c.restURL.String() + endpoint + "?" + params.Encode()

	} else {
		urlstr = c.restURL.String() + endpoint
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

	if c.sessiontoken != nil {
		token := "CCPSessionId " + *c.sessiontoken
		req.Header.Set("authenticate", token)

	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, resp.StatusCode, err
	}

	return resp.Body, resp.Header, resp.StatusCode, err

}

//POST Request for REST-API
//Accepts Context, Endpoint and Data as Input
//Returns io.ReadCloser,http.Header,Statuscode,error
func (c *Client) Post(ctx context.Context, endpoint string, data interface{}) (io.ReadCloser, http.Header, int, error) {
	err := c.Login(ctx)
	if err != nil {
		return nil, nil, 0, err
	}
	request, header, statuscode, err := c.request(ctx, "POST", endpoint, url.Values{}, data)

	//If Log enabled in options log data
	logDebug(ctx, c, fmt.Sprintf("Sent data in POST-Request: %v", data))

	//If Login is invalid - try again
	if statuscode == 401 {
		//Get new sessiontoken and write to var
		err := c.Login(ctx)
		request, header, statuscode, err := c.request(ctx, "POST", endpoint, url.Values{}, data)

		return request, header, statuscode, err
	}

	//If Statuscode not 201
	if statuscode != 201 {
		return request, header, statuscode, errorFormatterPx(ctx, c, statuscode, request)
	}

	return request, header, statuscode, err
}

//PUT Request for REST-API
//Accepts Endpoint and Data as Input
//Returns io.ReadCloser,http.Header,Statuscode,error
func (c *Client) Put(ctx context.Context, endpoint string, data interface{}) (io.ReadCloser, http.Header, int, error) {
	err := c.Login(ctx)
	if err != nil {
		return nil, nil, 0, err
	}
	request, header, statuscode, err := c.request(ctx, "PUT", endpoint, url.Values{}, data)

	//If Log enabled in options log data
	logDebug(ctx, c, fmt.Sprintf("Sent data in PUT-Request: %v", data))

	//If Login is invalid - try again
	if statuscode == 401 {
		//Get new sessiontoken and write to var
		err := c.Login(ctx)

		request, header, statuscode, err := c.request(ctx, "PUT", endpoint, url.Values{}, data)

		return request, header, statuscode, err
	}

	//If Statuscode not 204
	if statuscode != 204 {
		return request, header, statuscode, errorFormatterPx(ctx, c, statuscode, request)
	}

	//Write the latest sessiontoken back to var
	*c.sessiontoken = header.Get("sessiontoken")
	//c.Logout(sessiontoken)

	return request, header, statuscode, err
}

//GET Request for REST-API
//Accepts Endpoint and url.Values as Input
//Returns io.ReadCloser,http.Header,Statuscode,error
func (c *Client) Get(ctx context.Context, endpoint string, params url.Values) (io.ReadCloser, http.Header, int, error) {

	err := c.Login(ctx)

	if err != nil {
		return nil, nil, 0, err
	}

	request, header, statuscode, err := c.request(ctx, "GET", endpoint, params, nil)

	//If Login is invalid - try again
	if statuscode == 401 {
		//Get new sessiontoken and write to var
		err := c.Login(ctx)

		request, header, statuscode, err := c.request(ctx, "GET", endpoint, params, nil)

		return request, header, statuscode, err
	}

	//If Statuscode not 200
	if statuscode != 200 {
		return request, header, statuscode, errorFormatterPx(ctx, c, statuscode, request)
	}

	//Write the latest sessiontoken back to var
	*c.sessiontoken = header.Get("sessiontoken")

	return request, header, statuscode, err
}

//DELETE Request for REST-API
//Accepts Endpoint and url.Values as Input
//Returns io.ReadCloser,http.Header,Statuscode,error
func (c *Client) Delete(ctx context.Context, endpoint string) (io.ReadCloser, http.Header, int, error) {
	err := c.Login(ctx)
	if err != nil {
		return nil, nil, 0, err
	}
	request, header, statuscode, err := c.request(ctx, "DELETE", endpoint, nil, nil)

	//If Login is invalid - try again
	if statuscode == 401 {
		//Get new sessiontoken and write to var
		err := c.Login(ctx)
		request, header, statuscode, err := c.request(ctx, "DELETE", endpoint, nil, nil)

		return request, header, statuscode, err
	}

	//If Statuscode not 204
	if statuscode != 204 {
		return request, header, statuscode, errorFormatterPx(ctx, c, statuscode, request)
	}

	//Write the latest sessiontoken back to var
	*c.sessiontoken = header.Get("sessiontoken")

	return request, header, statuscode, err
}

func errorFormatterPx(ctx context.Context, c *Client, statuscode int, request io.Reader) (err error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(request)
	errbyte := buf.Bytes()
	errstr := buf.String()

	//Define Error Struct
	parsedError := ErrorStruct{}

	//Try to parse JSON in ErrorStruct
	err = json.Unmarshal(errbyte, &parsedError)

	//If error on parse return plain text
	if err != nil {
		return fmt.Errorf("ERROR: %v", errstr)
	}
	if len(parsedError.Fields) > 0 {
		var errorFields []string
		for _, field := range parsedError.Fields {
			errorFields = append(errorFields, field.Name)
		}
		return fmt.Errorf("Status: %v \n Type %s: %s\n Fields: %v", statuscode, parsedError.Type, parsedError.Message, errorFields)

	}
	return fmt.Errorf("Status: Statuscode %v\n Type %s: %s", statuscode, parsedError.Type, parsedError.Message)
}

func logDebug(ctx context.Context, c *Client, logtext string) {
	//If Log enabled in options
	if c.option.Log == true {
		log.Print(logtext)
	}
}
