package alsocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Get does a GET request to the API.
func (c *Client) Get(ctx context.Context, endpoint string, params url.Values) (io.ReadCloser, http.Header, int, error) {
	if params == nil {
		params = url.Values{}
	}
	logDebug(c, fmt.Sprintf("GET request to endpoint: %v with params: %v", endpoint, params))
	return c.doAuthenticatedRequest(ctx, http.MethodGet, endpoint, params, nil)
}

// Put does a PUT request to the API.
func (c *Client) Put(ctx context.Context, endpoint string, data interface{}) (io.ReadCloser, http.Header, int, error) {
	if data == nil {
		data = json.RawMessage(`{}`)
	}
	logDebug(c, fmt.Sprintf("Sent data in PUT-Request: %v", data))
	return c.doAuthenticatedRequest(ctx, http.MethodPut, endpoint, url.Values{}, data)
}

// Delete does a DELETE request to the API.
func (c *Client) Delete(ctx context.Context, endpoint string, data interface{}) (io.ReadCloser, http.Header, int, error) {
	if data == nil {
		data = json.RawMessage(`{}`)
	}
	logDebug(c, fmt.Sprintf("DELETE request to endpoint: %v", endpoint))
	return c.doAuthenticatedRequest(ctx, http.MethodDelete, endpoint, url.Values{}, data)
}

// Patch does a PATCH request to the API.
func (c *Client) Patch(ctx context.Context, endpoint string, data interface{}) (io.ReadCloser, http.Header, int, error) {
	if data == nil {
		data = json.RawMessage(`{}`)
	}
	logDebug(c, fmt.Sprintf("Sent data in PATCH-Request: %v", data))
	return c.doAuthenticatedRequest(ctx, http.MethodPatch, endpoint, url.Values{}, data)
}
