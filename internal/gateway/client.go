package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// A more specific error to return when the API gives a non-200 response
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: status %d, message: %s", e.StatusCode, e.Message)
}

type baseClient struct {
	baseURL    string
	httpClient *http.Client
}

// doProtoRequest is an helper method to create and execute an HTTP request for the gateway.
// If the user specifies a res, the value is populated with the unmarshaled response.
func (c *baseClient) doProtoRequest(ctx context.Context, method, path string, req, res proto.Message) error {
	var reqBody io.Reader
	if req != nil {
		bodyBytes, err := protojson.Marshal(req)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(bodyBytes)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "An unexpected error occurred",
		}
	}

	if res != nil {
		if err := protojson.Unmarshal(body, res); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
