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

type ClassificationResponse struct {
	// The label for the class (model specific)
	Label string `json:"label,omitempty"`

	// A float that represents how likely is that the text belongs in this class.
	Score float64 `json:"score,omitempty"`
}

type TextGenerationRequest struct {
	// (Required) a string to be generated from
	Inputs     string                   `json:"inputs"`
	Parameters TextGenerationParameters `json:"parameters,omitempty"`
	Options    Options                  `json:"options,omitempty"`
}

type TextGenerationResponse struct {
	GeneratedText string `json:"generated_text,omitempty"`
}

type TextGenerationParameters struct {
	// (Default: None). Integer to define the top tokens considered within the sample operation to create new text.
	TopK *int `json:"top_k,omitempty"`

	// (Default: None). Float to define the tokens that are within the sample` operation of text generation. Add
	// tokens in the sample for more probable to least probable until the sum of the probabilities is greater
	// than top_p.
	TopP *float64 `json:"top_p,omitempty"`

	// (Default: 1.0). Float (0.0-100.0). The temperature of the sampling operation. 1 means regular sampling,
	// 0 means top_k=1, 100.0 is getting closer to uniform probability.
	Temperature *float64 `json:"temperature,omitempty"`

	// (Default: None). Float (0.0-100.0). The more a token is used within generation the more it is penalized
	// to not be picked in successive generation passes.
	RepetitionPenalty *float64 `json:"repetition_penalty,omitempty"`

	// (Default: None). Int (0-250). The amount of new tokens to be generated, this does not include the input
	// length it is a estimate of the size of generated text you want. Each new tokens slows down the request,
	// so look for balance between response times and length of text generated.
	MaxNewTokens *int `json:"max_new_tokens,omitempty"`

	// (Default: None). Float (0-120.0). The amount of time in seconds that the query should take maximum.
	// Network can cause some overhead so it will be a soft limit. Use that in combination with max_new_tokens
	// for best results.
	MaxTime *float64 `json:"max_time,omitempty"`

	// (Default: True). Bool. If set to False, the return results will not contain the original query making it
	// easier for prompting.
	ReturnFullText *bool `json:"return_full_text,omitempty"`

	// (Default: 1). Integer. The number of proposition you want to be returned.
	NumReturnSequences *int `json:"num_return_sequences,omitempty"`
}

// Request structure for the token classification endpoint
type TokenClassificationRequest struct {
	// (Required) strings to be classified
	Inputs     string                        `json:"inputs"`
	Parameters TokenClassificationParameters `json:"parameters,omitempty"`
	Options    Options                       `json:"options,omitempty"`
}

type TokenClassificationParameters struct {
	// (Default: simple)
	AggregationStrategy string `json:"aggregation_strategy,omitempty"`
}

type TokenClassificationResponseEntity struct {
	// The type for the entity being recognized (model specific).
	EntityGroup string `json:"entity_group,omitempty"`

	// How likely the entity was recognized.
	Score float64 `json:"score,omitempty"`

	// The string that was captured
	Word string `json:"word,omitempty"`

	// The offset stringwise where the answer is located. Useful to disambiguate if Entity occurs multiple times.
	Start int `json:"start,omitempty"`

	// The offset stringwise where the answer is located. Useful to disambiguate if Entity occurs multiple times.
	End int `json:"end,omitempty"`
}

// Request structure for the Translation endpoint
type TranslationRequest struct {
	// (Required) a string to be translated in the original languages
	Inputs string `json:"inputs"`

	Options Options `json:"options,omitempty"`
}

// Response structure from the Translation endpoint
type TranslationResponse struct {
	// The translated Input string
	TranslationText string `json:"translation_text,omitempty"`
}

type ZeroShotRequest struct {
	// (Required)
	Inputs string `json:"inputs"`

	// (Required)
	Parameters ZeroShotParameters `json:"parameters"`

	Options Options `json:"options,omitempty"`
}

// Used with ZeroShotRequest
type ZeroShotParameters struct {
	// (Required) A list of strings that are potential classes for inputs. Max 10 candidate_labels,
	// for more, simply run multiple requests, results are going to be misleading if using
	// too many candidate_labels anyway. If you want to keep the exact same, you can
	// simply run multi_label=True and do the scaling on your end.
	CandidateLabels []string `json:"candidate_labels"`

	// (Default: false) Boolean that is set to True if classes can overlap
	MultiLabel *bool `json:"multi_label,omitempty"`
}

// Response structure from the Zero-shot classification endpoint.
type ZeroShotResponse struct {
	// The string sent as an input
	Sequence string `json:"sequence,omitempty"`

	// The list of labels sent in the request, sorted in descending order
	// by probability that the input corresponds to the to the label.
	Labels []string `json:"labels,omitempty"`

	// a list of floats that correspond the the probability of label, in the same order as labels.
	Scores []float64 `json:"scores,omitempty"`
}

type FeatureExtractionRequest struct {
	// (Required)
	Inputs string `json:"inputs"`

	Options Options `json:"options,omitempty"`
}

// Request structure for question answering model
type QuestionAnsweringRequest struct {
	// (Required)
	Inputs  QuestionAnsweringInputs `json:"inputs"`
	Options Options                 `json:"options,omitempty"`
}

type QuestionAnsweringInputs struct {
	// (Required) The question as a string that has an answer within Context.
	Question string `json:"question"`

	// (Required) A string that contains the answer to the question
	Context string `json:"context"`
}

// Response structure for question answering model
type QuestionAnsweringResponse struct {
	// A string that’s the answer within the Context text.
	Answer string `json:"answer,omitempty"`

	// A float that represents how likely that the answer is correct.
	Score float64 `json:"score,omitempty"`

	// The string index of the start of the answer within Context.
	Start int `json:"start,omitempty"`

	// The string index of the stop of the answer within Context.
	End int `json:"end,omitempty"`
}

// Request structure for table question answering model
type TableQuestionAnsweringRequest struct {
	Inputs  TableQuestionAnsweringInputs `json:"inputs"`
	Options Options                      `json:"options,omitempty"`
}

type TableQuestionAnsweringInputs struct {
	// (Required) The query in plain text that you want to ask the table
	Query string `json:"query"`

	// (Required) A table of data represented as a dict of list where entries
	// are headers and the lists are all the values, all lists must
	// have the same size.
	Table map[string][]string `json:"table"`
}

// Response structure for table question answering model
type TableQuestionAnsweringResponse struct {
	// The plaintext answer
	Answer string `json:"answer,omitempty"`

	// A list of coordinates of the cells references in the answer
	Coordinates [][]int `json:"coordinates,omitempty"`

	// A list of coordinates of the cells contents
	Cells []string `json:"cells,omitempty"`

	// The aggregator used to get the answer
	Aggregator string `json:"aggregator,omitempty"`
}

// Request structure for the Sentence Similarity endpoint.
type SentenceSimilarityRequest struct {
	// (Required) Inputs for the request.
	Inputs  SentenceSimilarityInputs `json:"inputs"`
	Options Options                  `json:"options,omitempty"`
}

type SentenceSimilarityInputs struct {
	// (Required) The string that you wish to compare the other strings with.
	// This can be a phrase, sentence, or longer passage, depending on the
	// model being used.
	SourceSentence string `json:"source_sentence"`

	// A list of strings which will be compared against the source_sentence.
	Sentences []string `json:"sentences"`
}

// Request structure for the conversational endpoint
type ConversationalRequest struct {
	// (Required)
	Inputs ConverstationalInputs `json:"inputs"`

	Parameters ConversationalParameters `json:"parameters,omitempty"`
	Options    Options                  `json:"options,omitempty"`
}

// Used with ConversationalRequest
type ConverstationalInputs struct {
	// (Required) The last input from the user in the conversation.
	Text string `json:"text"`

	// A list of strings corresponding to the earlier replies from the model.
	GeneratedResponses []string `json:"generated_responses,omitempty"`

	// A list of strings corresponding to the earlier replies from the user.
	// Should be of the same length of GeneratedResponses.
	PastUserInputs []string `json:"past_user_inputs,omitempty"`
}

// Used with ConversationalRequest
type ConversationalParameters struct {
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
	RepetitionPenalty *float64 `json:"repetition_penalty,omitempty"`

	// (Default: None). Float (0-120.0). The amount of time in seconds that the query should take maximum.
	// Network can cause some overhead so it will be a soft limit.
	MaxTime *float64 `json:"maxtime,omitempty"`
}

// Response structure for the conversational endpoint
type ConversationalResponse struct {
	// The answer of the model
	GeneratedText string `json:"generated_text,omitempty"`

	// A facility dictionary to send back for the next input (with the new user input addition).
	Conversation Conversation `json:"conversation,omitempty"`
}

// Used with ConversationalResponse
type Conversation struct {
	// The last outputs from the model in the conversation, after the model has run.
	GeneratedResponses []string `json:"generated_responses,omitempty"`

	// The last inputs from the user in the conversation, after the model has run.
	PastUserInputs []string `json:"past_user_inputs,omitempty"`
}

type ImageRequest struct {
	Image string `json:"image"`
}

type ImageSegmentationResponse struct {
	// The label for the class (model specific) of a segment.
	Label string `json:"label,omitempty"`

	// A float that represents how likely it is that the segment belongs to the given class.
	Score float64 `json:"score,omitempty"`

	// A str (base64 str of a single channel black-and-white img) representing the mask of a segment.
	Mask string `json:"mask,omitempty"`
}

type ObjectDetectionResponse struct {
	// The label for the class (model specific) of a detected object.
	Label string `json:"label,omitempty"`

	// A float that represents how likely it is that the detected object belongs to the given class.
	Score float64 `json:"score,omitempty"`

	// Bounding box of the detected object
	Box ObjectBox
}

type ObjectBox struct {
	XMin int `json:"xmin,omitempty"`
	YMin int `json:"ymin,omitempty"`
	XMax int `json:"xmax,omitempty"`
	YMax int `json:"ymax,omitempty"`
}

type ImageToTextResponse struct {
	// The generated caption
	GeneratedText string `json:"generated_text"`
}

type AudioRequest struct {
	Audio string `json:"audio"`
}

type SpeechRecognitionResponse struct {
	// The string that was recognized within the audio file.
	Text string `json:"text,omitempty"`
}