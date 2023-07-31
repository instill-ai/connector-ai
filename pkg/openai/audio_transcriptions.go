package openai

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/instill-ai/connector-ai/pkg/util"
)

const (
	transcriptionsURL = host + "/v1/audio/transcriptions"
)

type AudioTranscriptionReq struct {
	File        []byte  `json:"file"`
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Language    string  `json:"language"`
	Temperature float64 `json:"temperature"`
}

type AudioTranscriptionResp struct {
	Text string `json:"text"`
}

// GenerateAudioTranscriptions makes a call to the audio transcriptions API from OpenAI.
// https://platform.openai.com/docs/api-reference/audio/create-transcription
func (c *Client) GenerateAudioTranscriptions(req AudioTranscriptionReq) (result AudioTranscriptionResp, err error) {
	formData, contentType, err := getBytes(req)
	if err != nil {
		return result, err
	}
	err = c.sendReq(transcriptionsURL, http.MethodPost, contentType, formData, &result)
	return result, err
}

func getBytes(req AudioTranscriptionReq) (*bytes.Reader, string, error) {
	data := &bytes.Buffer{}
	writer := multipart.NewWriter(data)
	err := util.WriteFile(writer, "file", req.File)
	if err != nil {
		return nil, "", err
	}
	util.WriteField(writer, "model", req.Model)
	util.WriteField(writer, "prompt", req.Prompt)
	util.WriteField(writer, "language", req.Language)
	if req.Temperature != 0 {
		util.WriteField(writer, "temperature", fmt.Sprintf("%f", req.Temperature))
	}
	writer.Close()
	return bytes.NewReader(data.Bytes()), writer.FormDataContentType(), nil
}
