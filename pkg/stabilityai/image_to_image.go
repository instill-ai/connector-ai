package stabilityai

// ImageToImageReq represents the request body for image-to-image API
type ImageToImageReq struct {
	ImageTaskCommon
	InitImage     string  `json:"init_image"`
	InitImageMode string  `json:"init_image_mode,omitempty"`
	ImageStrength float64 `json:"image_strength,omitempty"`
}

// GenerateImageFromImage makes a call to the image-to-image API from Stability AI.
// https://platform.stability.ai/rest-api#tag/v1generation/operation/imageToImage
func (c *Client) GenerateImageFromImage(params ImageToImageReq, engine string) (results []Image, err error) {
	var resp ImageTaskRes
	// default engine
	if engine == "" {
		engine = "stable-diffusion-v1-5"
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
