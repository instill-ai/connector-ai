package openai

import "net/http"

const (
	embeddingsURL = host + "/v1/embeddings"
)

type TextEmbeddingsReq struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type TextEmbeddingsResp struct {
	Object string `json:"object"`
	Data   []Data `json:"data"`
	Model  string `json:"model"`
	Usage  Usage  `json:"usage"`
}

type Data struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// GenerateTextEmbeddings makes a call to the embeddings API from OpenAI.
// https://platform.openai.com/docs/api-reference/embeddings
func (c *Client) GenerateTextEmbeddings(req TextEmbeddingsReq) (result TextEmbeddingsResp, err error) {
	err = c.sendReq(embeddingsURL, http.MethodPost, req, &result)
	return result, err
}