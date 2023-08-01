package stabilityai

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/instill-ai/connector-ai/pkg/util"
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
	err := util.WriteFile(writer, "init_image", []byte(req.InitImage))
	if err != nil {
		return nil, "", err
	}
	util.WriteField(writer, "cfg_scale", fmt.Sprintf("%f", req.CFGScale))
	util.WriteField(writer, "clip_guidance_preset", req.ClipGuidancePreset)
	util.WriteField(writer, "sampler", req.Sampler)
	util.WriteField(writer, "seed", fmt.Sprintf("%d", req.Seed))
	util.WriteField(writer, "style_preset", req.StylePreset)
	util.WriteField(writer, "init_image_mode", req.InitImageMode)
	util.WriteField(writer, "image_strength", fmt.Sprintf("%f", req.ImageStrength))
	if req.Samples != 0 {
		util.WriteField(writer, "samples", fmt.Sprintf("%d", req.Samples))
	}
	if req.Steps != 0 {
		util.WriteField(writer, "steps", fmt.Sprintf("%d", req.Steps))
	}

	i := 0
	for _, t := range req.TextPrompts {
		if t.Text == "" {
			continue
		}
		util.WriteField(writer, fmt.Sprintf("text_prompts[%d][text]", i), t.Text)
		util.WriteField(writer, fmt.Sprintf("text_prompts[%d][weight]", i), fmt.Sprintf("%f", t.Weight))
		i++
	}
	writer.Close()
	return bytes.NewReader(data.Bytes()), writer.FormDataContentType(), nil
}
