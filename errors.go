package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type APIError struct {
	*http.Response
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url,omitempty"`

	Errors []struct {
		Code     string
		Resource string
		Field    string
	} `json:"errors,omitempty"`
}

func newApiError(res *http.Response) error {
	if res == nil {
		return errors.New("received nil response")
	}

	errResponse := &APIError{
		Response: res,
	}

	if err := json.NewDecoder(res.Body).Decode(errResponse); err != nil {
		errResponse.Message = fmt.Sprintf("Request failed with status %d", res.StatusCode)
	}

	return errResponse
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error: %d - %s\n", e.StatusCode, e.Message)
}
