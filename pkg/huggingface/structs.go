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

// Request structure for the Fill Mask endpoint
type FillMaskRequest struct {
	// (Required) a string to be filled from, must contain the [MASK] token (check model card for exact name of the mask)
	Inputs  string  `json:"inputs,omitempty"`
	Options Options `json:"options,omitempty"`
}

// Used in the FillMaskResponse struct
type FillMaskResponseEntry struct {
	// The actual sequence of tokens that ran against the model (may contain special tokens)
	Sequence string `json:"sequence,omitempty"`

	// The probability for this token.
	Score float64 `json:"score,omitempty"`

	// The id of the token
	Token int `json:"token,omitempty"`

	// The string representation of the token
	TokenStr string `json:"token_str,omitempty"`
}

// Request structure for the summarization endpoint
type SummarizationRequest struct {
	// String to be summarized
	Inputs     string                  `json:"inputs"`
	Parameters SummarizationParameters `json:"parameters,omitempty"`
	Options    Options                 `json:"options,omitempty"`
}

// Used with SummarizationRequest
type SummarizationParameters struct {
	// (Default: None). Integer to define the minimum length in tokens of the output summary.
	MinLength *int `json:"min_length,omitempty"`

	// (Default: None). Integer to define the maximum length in tokens of the output summary.
	MaxLength *int `json:"max_length,omitempty"`

	// (Default: None). Integer to define the top tokens considered within the sample operation to create
	// new text.
	TopK *int `json:"top_k,omitempty"`

	// (Default: None). Float to define the tokens that are within the sample` operation of text generation.
	// Add tokens in the sample for more probable to least probable until the sum of the probabilities is
	// greater than top_p.
	TopP *float64 `json:"top_p,omitempty"`

	// (Default: 1.0). Float (0.0-100.0). The temperature of the sampling operation. 1 means regular sampling,
	// 0 mens top_k=1, 100.0 is getting closer to uniform probability.
	Temperature *float64 `json:"temperature,omitempty"`

	// (Default: None). Float (0.0-100.0). The more a token is used within generation the more it is penalized
	// to not be picked in successive generation passes.
	RepetitionPenalty *float64 `json:"repetitionpenalty,omitempty"`

	// (Default: None). Float (0-120.0). The amount of time in seconds that the query should take maximum.
	// Network can cause some overhead so it will be a soft limit.
	MaxTime *float64 `json:"maxtime,omitempty"`
}

// Response structure for the summarization endpoint
type SummarizationResponse struct {
	// The summarized input string
	SummaryText string `json:"summary_text,omitempty"`
}

// Request structure for the Text classification endpoint
type TextClassificationRequest struct {
	//String to be classified
	Inputs  string  `json:"inputs"`
	Options Options `json:"options,omitempty"`
}

// Used in TextClassificationResponse
type TextClassificationResponseLabel struct {
	// The label for the class (model specific)
	Label string `json:"label,omitempty"`

	// A float that represents how likely is that the text belongs in this class.
	Score float64 `json:"score,omitempty"`
}
