package stabilityai

import "fmt"

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
func (c *Client) GenerateImageFromImage(params ImageToImageReq, engine string) (results []Image, err error) {
	var resp ImageTaskRes
	if engine == "" {
		return nil, fmt.Errorf("no engine selected")
	}
	imageToImageURL := host + "/v1/generation/" + engine + "/image-to-image"
	err = c.sendReq(imageToImageURL, "POST", params, &resp)
	for _, i := range resp.Images {
		if i.FinishReason == successFinishReason {
			results = append(results, i)
		}
	}
	return
}
