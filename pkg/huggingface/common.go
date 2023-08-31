package huggingface

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

const (
	AuthHeaderKey    = "Authorization"
	AuthHeaderPrefix = "Bearer "
	modelsPath       = "/models/"
)

// MakeHFAPIRequest builds and sends an HTTP POST request to the given model
// using the provided JSON body. If the request is successful, returns the
// response JSON and a nil error. If the request fails, returns an empty slice
// and an error describing the failure.
func (c *Client) MakeHFAPIRequest(body []byte, model string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, baseURL+modelsPath+model, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.New("nil request created")
	}
	req.Header.Set(AuthHeaderKey, AuthHeaderPrefix+c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = checkRespForError(respBody)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

type apiError struct {
	Error string `json:"error,omitempty"`
}

type apiErrors struct {
	Errors []string `json:"error,omitempty"`
}

// Checks for errors in the API response and returns them if
// found.
func checkRespForError(respJSON []byte) error {
	// Check for single error
	{
		buf := make([]byte, len(respJSON))
		copy(buf, respJSON)
		apiErr := apiError{}
		json.Unmarshal(buf, &apiErr)
		if apiErr.Error != "" {
			return errors.New(string(respJSON))
		}
	}
	// Check for multiple errors
	{
		buf := make([]byte, len(respJSON))
		copy(buf, respJSON)
		apiErrs := apiErrors{}
		json.Unmarshal(buf, &apiErrs)
		if apiErrs.Errors != nil {
			return errors.New(string(respJSON))
		}
	}
	return nil
}

func (c *Client) GetConnectionState() (connectorPB.ConnectorResource_State, error) {
	req, _ := http.NewRequest(http.MethodGet, baseURL, nil)
	req.Header.Set("Content-Type", jsonMimeType)
	req.Header.Set(AuthHeaderKey, AuthHeaderPrefix+c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return connectorPB.ConnectorResource_STATE_ERROR, err
	}
	if resp != nil && resp.StatusCode == http.StatusOK {
		return connectorPB.ConnectorResource_STATE_CONNECTED, nil
	}
	return connectorPB.ConnectorResource_STATE_DISCONNECTED, nil
}
