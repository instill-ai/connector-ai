package stabilityai

const (
	successFinishReason = "SUCCESS"
)

// TextToImageReq represents the request body for text-to-image API
type TextToImageReq struct {
	//required params
	TextPrompts []TextPrompt `json:"text_prompts"`
	//optional params
	Height             uint32  `json:"height,omitempty"`
	Width              uint32  `json:"width,omitempty"`
	CFGScale           float32 `json:"cfg_scale,omitempty"`
	ClipGuidancePreset string  `json:"clip_guidance_preset,omitempty"`
	Sampler            string  `json:"sampler,omitempty"`
	Samples            uint32  `json:"samples,omitempty"`
	Seed               uint32  `json:"seed,omitempty"`
	Steps              uint32  `json:"steps,omitempty"`
	StylePreset        string  `json:"style_preset,omitempty"`
}

// TextPrompt holds a prompt's text and its weight.
type TextPrompt struct {
	Text   string  `json:"text"`
	Weight float32 `json:"weight"`
}

// TextToImage represents a single image response.
type TextToImage struct {
	Base64       string `json:"base64"`
	Seed         uint32 `json:"seed"`
	FinishReason string `json:"finishReason"`
}

// TextToImageRes represents the response body for text-to-image API
type TextToImageRes struct {
	Images []TextToImage `json:"artifacts"`
}

// GenerateImageFromText makes a call to the text-to-image API from Stability AI.
// https://platform.stability.ai/rest-api#tag/v1generation/operation/textToImage
func (c *Client) GenerateImageFromText(params TextToImageReq, engine string) (results []TextToImage, err error) {
	var resp TextToImageRes
	// default engine
	if engine == "" {
		engine = "stable-diffusion-v1-5"
	}
	textToImageURL := host + "/v1/generation/" + engine + "/text-to-image"
	err = c.sendReq(textToImageURL, "POST", params, &resp)
	for _, i := range resp.Images {
		if i.FinishReason == successFinishReason {
			results = append(results, i)
		}
	}
	return
}
