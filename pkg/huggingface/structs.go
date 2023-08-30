package huggingface

// Request structure for text-to-image model
type TextToImageRequest struct {
	// The prompt or prompts to guide the image generation.
	Inputs     string                       `json:"inputs"`
	Options    Options                      `json:"options,omitempty"`
	Parameters TextToImageRequestParameters `json:"parameters,omitempty"`
}

// Request structure for text-to-image model
type TextToImageResponse struct {
	Image string `json:"image"`
}

type Options struct {
	// (Default: false). Boolean to use GPU instead of CPU for inference.
	// Requires Startup plan at least.
	UseGPU *bool `json:"use_gpu,omitempty"`
	// (Default: true). There is a cache layer on the inference API to speedup
	// requests we have already seen. Most models can use those results as is
	// as models are deterministic (meaning the results will be the same anyway).
	// However if you use a non deterministic model, you can set this parameter
	// to prevent the caching mechanism from being used resulting in a real new query.
	UseCache *bool `json:"use_cache,omitempty"`
	// (Default: false) If the model is not ready, wait for it instead of receiving 503.
	// It limits the number of requests required to get your inference done. It is advised
	// to only set this flag to true after receiving a 503 error as it will limit hanging
	// in your application to known places.
	WaitForModel *bool `json:"wait_for_model,omitempty"`
}

type TextToImageRequestParameters struct {
	// The prompt or prompts not to guide the image generation.
	// Ignored when not using guidance (i.e., ignored if guidance_scale is less than 1).
	NegativePrompt string `json:"negative_prompt,omitempty"`
	// The height in pixels of the generated image.
	Height int64 `json:"height,omitempty"`
	// The width in pixels of the generated image.
	Width int64 `json:"width,omitempty"`
	// The number of denoising steps. More denoising steps usually lead to a higher quality
	// image at the expense of slower inference. Defaults to 50.
	NumInferenceSteps int64 `json:"num_inference_steps,omitempty"`
	// Higher guidance scale encourages to generate images that are closely linked to the text
	// input, usually at the expense of lower image quality. Defaults to 7.5.
	GuidanceScale float64 `json:"guidance_scale,omitempty"`
}
