package alsocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Get does a GET request to API.
// It Accepts Context, Endpoint and URL Parameters as Input.
// Returns io.ReadCloser, http.Header, Statuscode, error.
func (c *Client) Get(ctx context.Context, endpoint string, params url.Values) (io.ReadCloser, http.Header, int, error) {
	err := c.login(ctx)
	if err != nil {
		return nil, nil, 0, err
	}

	request, header, statuscode, err := c.request(ctx, http.MethodGet, endpoint, params, nil)

	// If Log enabled in options log data
	logDebug(ctx, c, fmt.Sprintf("GET request to endpoint: %v with params: %v", endpoint, params))

	// If error return nicely formatted error
	if err != nil {
		return request, header, statuscode, errorFormatterPx(ctx, c, statuscode, request)
	}
	return request, header, statuscode, nil
}

// Put does a PUT request to API.
// It Accepts Context, Endpoint and Data as Input.
// Returns io.ReadCloser, http.Header, Statuscode, error.
func (c *Client) Put(ctx context.Context, endpoint string, data interface{}) (io.ReadCloser, http.Header, int, error) {
	err := c.login(ctx)
	if err != nil {
		return nil, nil, 0, err
	}

	// Fix for sending empty json body if data is nil
	if data == nil {
		data = json.RawMessage(`{}`)
	}
	request, header, statuscode, err := c.request(ctx, http.MethodPut, endpoint, url.Values{}, data)

	// If Log enabled in options log data
	logDebug(ctx, c, fmt.Sprintf("Sent data in PUT-Request: %v", data))

	// If error return nicely formatted error
	if err != nil {
		return request, header, statuscode, errorFormatterPx(ctx, c, statuscode, request)
	}
	return request, header, statuscode, nil
}

// Delete does a DELETE request to API.
// It Accepts Context, Endpoint and optional Data as Input.
// Returns io.ReadCloser, http.Header, Statuscode, error.
func (c *Client) Delete(ctx context.Context, endpoint string, data interface{}) (io.ReadCloser, http.Header, int, error) {
	err := c.login(ctx)
	if err != nil {
		return nil, nil, 0, err
	}

	// Fix for sending empty json body if data is nil
	if data == nil {
		data = json.RawMessage(`{}`)
	}
	request, header, statuscode, err := c.request(ctx, http.MethodDelete, endpoint, url.Values{}, data)

	// If Log enabled in options log data
	logDebug(ctx, c, fmt.Sprintf("DELETE request to endpoint: %v", endpoint))

	// If error return nicely formatted error
	if err != nil {
		return request, header, statuscode, errorFormatterPx(ctx, c, statuscode, request)
	}
	return request, header, statuscode, nil
}

// Patch does a PATCH request to API.
// It Accepts Context, Endpoint and Data as Input.
// Returns io.ReadCloser, http.Header, Statuscode, error.
func (c *Client) Patch(ctx context.Context, endpoint string, data interface{}) (io.ReadCloser, http.Header, int, error) {
	err := c.login(ctx)
	if err != nil {
		return nil, nil, 0, err
	}

	// Fix for sending empty json body if data is nil
	if data == nil {
		data = json.RawMessage(`{}`)
	}
	request, header, statuscode, err := c.request(ctx, http.MethodPatch, endpoint, url.Values{}, data)

	// If Log enabled in options log data
	logDebug(ctx, c, fmt.Sprintf("Sent data in PATCH-Request: %v", data))

	// If error return nicely formatted error
	if err != nil {
		return request, header, statuscode, errorFormatterPx(ctx, c, statuscode, request)
	}
	return request, header, statuscode, nil
}
