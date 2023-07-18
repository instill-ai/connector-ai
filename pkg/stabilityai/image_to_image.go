package stabilityai

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/http"
)

// ImageToImageReq represents the request body for image-to-image API
type ImageToImageReq struct {
	TextPrompts        []TextPrompt `json:"text_prompts" om:"texts[:]"`
	CFGScale           float64      `json:"cfg_scale,omitempty" om:"metadata.cfg_scale"`
	ClipGuidancePreset string       `json:"clip_guidance_preset,omitempty" om:"metadata.clip_guidance_preset"`
	Sampler            string       `json:"sampler,omitempty" om:"metadata.sampler"`
	Samples            uint32       `json:"samples,omitempty" om:"metadata.samples"`
	Seed               uint32       `json:"seed,omitempty" om:"metadata.seed"`
	Steps              uint32       `json:"steps,omitempty" om:"metadata.steps"`
	StylePreset        string       `json:"style_preset,omitempty" om:"metadata.style_preset"`
	InitImage          string       `json:"init_image" om:"images[0]"`
	InitImageMode      string       `json:"init_image_mode,omitempty" om:"metadata.init_image_mode"`
	ImageStrength      float64      `json:"image_strength,omitempty" om:"metadata.image_strength"`
}

// GenerateImageFromImage makes a call to the image-to-image API from Stability AI.
// https://platform.stability.ai/rest-api#tag/v1generation/operation/imageToImage
func (c *Client) GenerateImageFromImage(req ImageToImageReq, engine string) (results []Image, err error) {
	var resp ImageTaskRes
	if engine == "" {
		return nil, fmt.Errorf("no engine selected")
	}
	imageToImageURL := host + "/v1/generation/" + engine + "/image-to-image"
	formData, contentType, err := getBytes(req)
	if err != nil {
		return nil, err
	}
	err = c.sendReq(imageToImageURL, http.MethodPost, contentType, formData, &resp)
	for _, i := range resp.Images {
		if i.FinishReason == successFinishReason {
			results = append(results, i)
		}
	}
	return
}

func getBytes(req ImageToImageReq) (*bytes.Reader, string, error) {
	data := &bytes.Buffer{}
	writer := multipart.NewWriter(data)
	err := writeImage(writer, req.InitImage)
	if err != nil {
		return nil, "", err
	}
	writeField(writer, "cfg_scale", fmt.Sprintf("%f", req.CFGScale))
	writeField(writer, "clip_guidance_preset", req.ClipGuidancePreset)
	writeField(writer, "sampler", req.Sampler)
	writeField(writer, "seed", fmt.Sprintf("%d", req.Seed))
	writeField(writer, "style_preset", req.StylePreset)
	writeField(writer, "init_image_mode", req.InitImageMode)
	writeField(writer, "image_strength", fmt.Sprintf("%f", req.ImageStrength))
	if req.Samples != 0 {
		writeField(writer, "samples", fmt.Sprintf("%d", req.Samples))
	}
	if req.Steps != 0 {
		writeField(writer, "steps", fmt.Sprintf("%d", req.Steps))
	}

	i := 0
	for _, t := range req.TextPrompts {
		if t.Text == "" {
			continue
		}
		_ = writer.WriteField(fmt.Sprintf("text_prompts[%d][text]", i), t.Text)
		_ = writer.WriteField(fmt.Sprintf("text_prompts[%d][weight]", i), fmt.Sprintf("%f", t.Weight))
		i++
	}
	writer.Close()
	return bytes.NewReader(data.Bytes()), writer.FormDataContentType(), nil
}

func writeField(writer *multipart.Writer, key string, value string) {
	if key != "" && value != "" {
		_ = writer.WriteField(key, value)
	}
}

func writeImage(writer *multipart.Writer, imgBase64 string) error {
	part, err := writer.CreateFormFile("init_image", "")
	if err != nil {
		return err
	}
	imageData, _ := base64.StdEncoding.DecodeString(imgBase64)
	_, err = part.Write(imageData)
	return err
}
