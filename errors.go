package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError represents an error returned by the API.
// It contains error details including status code, message,
// and optional documentation URL for further information.
type APIError struct {
	// Message contains the error message returned by the API
	Message string `json:"message"`

	// DocumentationUrl provides a link to the API documentation
	// related to this error, if available
	DocumentationUrl string `json:"documentation_url,omitempty"`

	// StatusCode contains the HTTP status code of the response
	StatusCode int

	// Errors contains detailed error information when multiple
	// errors are returned by the API
	Errors []APIErrorDetail `json:"errors,omitempty"`
}

// APIErrorDetail represents detailed information about a specific error.
// It provides additional context about what went wrong during API requests.
type APIErrorDetail struct {
	// Code represents the error code (e.g., "401", "404", "500")
	Code string `json:"code,omitempty"`

	// Resource indicates the type of resource that caused the error
	// (e.g., "Issue", "Repository", "User")
	Resource string `json:"resource,omitempty"`

	// Field specifies which field caused the error (e.g., "title", "body", "name")
	Field string `json:"field,omitempty"`
}

func newAPIError(res *http.Response) error {
	apiErr := &APIError{
		StatusCode: res.StatusCode,
	}

	err := json.NewDecoder(res.Body).Decode(apiErr)
	if err != nil {
		apiErr.Message = fmt.Sprintf("request failed with status %d", res.StatusCode)
	}

	return apiErr
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error: %d - %s", e.StatusCode, e.Message)
}
