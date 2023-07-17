package openai

import (
	"net/http"
)

const (
	completionsURL = host + "/v1/chat/completions"
)

type TextCompletionReq struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Temperature      float32   `json:"temperature,omitempty"`
	TopP             float32   `json:"top_p,omitempty"`
	N                int       `json:"n,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
	Stop             string    `json:"stop,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	PresencePenalty  float32   `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32   `json:"frequency_penalty,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TextCompletionResp struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int       `json:"created"`
	Choices []Choices `json:"choices"`
	Usage   Usage     `json:"usage"`
}

type Choices struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

/*
{
  "id": "cmpl-uqkvlQyYK7bGYrRHQ0eXlWi7",
  "object": "text_completion",
  "created": 1589478378,
  "model": "text-davinci-003",
  "choices": [
    {
      "text": "\n\nThis is indeed a test",
      "index": 0,
      "logprobs": null,
      "finish_reason": "length"
    }
  ],
  "usage": {
    "prompt_tokens": 5,
    "completion_tokens": 7,
    "total_tokens": 12
  }
}
*/

// GenerateTextCompletion makes a call to the completions API from OpenAI.
// https://platform.openai.com/docs/api-reference/completions
func (c *Client) GenerateTextCompletion(req TextCompletionReq) (result TextCompletionResp, err error) {
	err = c.sendReq(completionsURL, http.MethodPost, req, &result)
	return result, err
}
