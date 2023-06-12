package stabilityai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	host         = "https://api.stability.ai"
	jsonMimeType = "application/json"
	reqTimeout   = time.Second * 60
)

// makeReq is responsible for making the http request with to given URL, method, and params and unmarshalling the response into given object.
func makeReq(reqURL, method, apiKey string, params interface{}, respObj interface{}) (err error) {
	data, _ := json.Marshal(params)
	req, _ := http.NewRequest(method, reqURL, bytes.NewBuffer(data))
	req.Header.Add("Content-Type", jsonMimeType)
	req.Header.Add("Accept", jsonMimeType)
	req.Header.Add("Authorization", "Bearer "+apiKey)
	http.DefaultClient.Timeout = reqTimeout
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var body interface{}
		_ = json.NewDecoder(res.Body).Decode(&body)
		err = fmt.Errorf("non-200 response: %s", body)
		return
	}
	if err = json.NewDecoder(res.Body).Decode(&respObj); err != nil {
		err = fmt.Errorf("error in json decode: %s, invalid response: %v", err, res)
	}
	return
}
