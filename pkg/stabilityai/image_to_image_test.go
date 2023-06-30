package stabilityai

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateImageFromImage(t *testing.T) {
	//setup mocks
	mockClient := &MockHTTPClient{}
	stabilityClient := Client{APIKey: "test_key", HTTPClient: mockClient}

	tests := []struct {
		name        string
		req         ImageToImageReq
		expectedRes []Image
		expectedErr error
		setMocks    func()
	}{
		{
			name:        "error in sending request",
			req:         ImageToImageReq{InitImage: "a", TextPrompts: []TextPrompt{{Text: "a cat and a dog", Weight: 0.5}}},
			expectedErr: errors.New("error occurred: call failed, while calling URL: https://api.stability.ai/v1/generation/stable-diffusion-v1-5/image-to-image, request body: {\"text_prompts\":[{\"text\":\"a cat and a dog\",\"weight\":0.5}],\"init_image\":\"a\"}"),
			setMocks: func() {
				mockResponse = func() (*http.Response, error) {
					return nil, errors.New("call failed")
				}
			},
		},
		{
			name:        "nil response",
			req:         ImageToImageReq{InitImage: "a", TextPrompts: []TextPrompt{{Text: "a cat and a dog", Weight: 0.5}}},
			expectedErr: errors.New("error occurred: <nil>, while calling URL: https://api.stability.ai/v1/generation/stable-diffusion-v1-5/image-to-image, request body: {\"text_prompts\":[{\"text\":\"a cat and a dog\",\"weight\":0.5}],\"init_image\":\"a\"}"),
			setMocks: func() {
				mockResponse = func() (*http.Response, error) {
					return nil, nil
				}
			},
		},
		{
			name:        "non 200 status code",
			req:         ImageToImageReq{InitImage: "a", TextPrompts: []TextPrompt{{Text: "a cat and a dog", Weight: 0.5}}},
			expectedErr: errors.New("non-200 status code: 401, while calling URL: https://api.stability.ai/v1/generation/stable-diffusion-v1-5/image-to-image, response body: {\"id\": \"9160aa70-222f-4a36-9eb7-475e2668362a\",\"name\": \"unauthorized\",\"message\": \"missing authorization header\"}"),
			setMocks: func() {
				mockResponse = func() (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusUnauthorized, Body: io.NopCloser(bytes.NewBuffer([]byte(`{"id": "9160aa70-222f-4a36-9eb7-475e2668362a","name": "unauthorized","message": "missing authorization header"}`)))}, nil
				}
			},
		},
		{
			name:        "json unmarshal error invalid response",
			req:         ImageToImageReq{InitImage: "a", TextPrompts: []TextPrompt{{Text: "a cat and a dog", Weight: 0.5}}},
			expectedErr: errors.New("error in json decode: unexpected end of JSON input, while calling URL: https://api.stability.ai/v1/generation/stable-diffusion-v1-5/image-to-image, response body: {"),
			setMocks: func() {
				mockResponse = func() (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer([]byte(`{`)))}, nil
				}
			},
		},
		{
			name:        "valid response",
			req:         ImageToImageReq{InitImage: "a", TextPrompts: []TextPrompt{{Text: "a cat and a dog", Weight: 0.5}}},
			expectedRes: []Image{{Base64: "a", Seed: 1234, FinishReason: "SUCCESS"}},
			setMocks: func() {
				mockResponse = func() (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer([]byte(`{"artifacts":[{"base64":"a","seed":1234,"finishReason":"SUCCESS"}]}`)))}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setMocks != nil {
				tt.setMocks()
			}
			res, err := stabilityClient.GenerateImageFromImage(tt.req, "stable-diffusion-v1-5")
			assert.Equal(t, res, tt.expectedRes)
			assert.Equal(t, err, tt.expectedErr)
		})
	}
}

func TestTemp(t *testing.T) {
	uuid.NewV4()
}
