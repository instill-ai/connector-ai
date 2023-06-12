package stabilityai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	host         = "https://api.stability.ai"
	jsonMimeType = "application/json"
	reqTimeout   = time.Second * 60
)

// Client represents a Stability AI client
type Client struct {
	APIKey string
}

// NewClient initializes a new Stability AI client
func NewClient(apiKey string) Client {
	return Client{APIKey: apiKey}
}

// makeReq is responsible for making the http request with to given URL, method, and params and unmarshalling the response into given object.
func (c *Client) makeReq(reqURL, method string, params interface{}, respObj interface{}) (err error) {
	data, _ := json.Marshal(params)
	req, _ := http.NewRequest(method, reqURL, bytes.NewBuffer(data))
	req.Header.Add("Content-Type", jsonMimeType)
	req.Header.Add("Accept", jsonMimeType)
	req.Header.Add("Authorization", "Bearer "+c.APIKey)
	http.DefaultClient.Timeout = reqTimeout
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	bytes, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("non-200 status code: %d, while calling URL: %s, response body: %#v", res.StatusCode, reqURL, bytes)
		return
	}
	if err = json.Unmarshal(bytes, &respObj); err != nil {
		err = fmt.Errorf("error in json decode: %s, while calling URL: %s, response body: %#v", err, reqURL, bytes)
	}
	return
}
