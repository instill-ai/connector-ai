package huggingface

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/connector/pkg/base"
	"github.com/instill-ai/connector/pkg/configLoader"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

const (
	venderName   = "huggingface"
	baseURL      = "https://api-inference.huggingface.co/models/"
	jsonMimeType = "application/json"
	reqTimeout   = time.Second * 60 * 5
	//tasks
	textGenerationTask         = "TEXT_GENERATION"
	textToImageTask            = "TEXT_TO_IMAGE"
	fillMaskTask               = "FILL_MASK"
	summarizationTask          = "SUMMARIZATION"
	textClassificationTask     = "TEXT_CLASSIFICATION"
	tokenClassificationTask    = "TOKEN_CLASSIFICATION"
	translationTask            = "TRANSLATION"
	zeroShotClassificationTask = "ZERO_SHOT_CLASSIFICATION"
	featureExtractionTask      = "FEATURE_EXTRACTION"
	questionAnsweringTask      = "QUESTION_ANSWERING"
	tableQuestionAnsweringTask = "TABLE_QUESTION_ANSWERING"
	sentenceSimilarityTask     = "SENTENCE_SIMILARITY"
	conversationalTask         = "CONVERSATIONAL"
)

var (
	//go:embed config/definitions.json
	definitionJSON []byte
	once           sync.Once
	connector      base.IConnector
)

type ConnectorOptions struct{}

type Connector struct {
	base.BaseConnector
	options ConnectorOptions
}

type Connection struct {
	base.BaseConnection
	connector *Connector
}

// Client represents a OpenAI client
type Client struct {
	APIKey     string
	HTTPClient HTTPClient
}

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func Init(logger *zap.Logger, options ConnectorOptions) base.IConnector {
	once.Do(func() {
		loader := configLoader.InitJSONSchema(logger)
		connDefs, err := loader.Load(venderName, connectorPB.ConnectorType_CONNECTOR_TYPE_AI, definitionJSON)
		if err != nil {
			panic(err)
		}
		connector = &Connector{
			BaseConnector: base.BaseConnector{Logger: logger},
			options:       options,
		}
		for idx := range connDefs {
			err := connector.AddConnectorDefinition(uuid.FromStringOrNil(connDefs[idx].GetUid()), connDefs[idx].GetId(), connDefs[idx])
			if err != nil {
				logger.Warn(err.Error())
			}
		}
	})
	return connector
}

func (c *Connector) CreateConnection(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (base.IConnection, error) {
	def, err := c.GetConnectorDefinitionByUid(defUid)
	if err != nil {
		return nil, err
	}
	return &Connection{
		BaseConnection: base.BaseConnection{
			Logger: logger, DefUid: defUid,
			Config:     config,
			Definition: def,
		},
		connector: c,
	}, nil
}

// NewClient initializes a new Hugging Face client
func NewClient(apiKey string) Client {
	tr := &http.Transport{DisableKeepAlives: true}
	return Client{APIKey: apiKey, HTTPClient: &http.Client{Timeout: reqTimeout, Transport: tr}}
}

func (c *Connection) getAPIKey() string {
	return c.Config.GetFields()["api_key"].GetStringValue()
}

func (c *Connection) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	client := NewClient(c.getAPIKey())
	outputs := []*structpb.Struct{}
	task := inputs[0].GetFields()["task"].GetStringValue()
	model := inputs[0].GetFields()["model"].GetStringValue()
	for _, input := range inputs {
		if input.GetFields()["task"].GetStringValue() != task {
			return nil, fmt.Errorf("each input should be the same task")
		}
	}
	if err := c.ValidateInput(inputs, task); err != nil {
		return nil, err
	}

	for _, input := range inputs {
		switch task {
		case textGenerationTask:
			inputStruct := TextGenerationRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []TextGenerationResponse{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			generatedTexts := structpb.ListValue{}
			generatedTexts.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				generatedTexts.Values[i] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: outputArr[i].GeneratedText}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"texts": {Kind: &structpb.Value_ListValue{ListValue: &generatedTexts}}},
			}
			outputs = append(outputs, &output)
		case textToImageTask:
			inputStruct := TextToImageRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputStruct := TextToImageResponse{Image: base64.StdEncoding.EncodeToString(resp)}
			outputJson, err := json.Marshal(outputStruct)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(outputJson, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)
		case fillMaskTask:
			inputStruct := FillMaskRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []FillMaskResponseEntry{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			masks := structpb.ListValue{}
			masks.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				masks.Values[i] = &structpb.Value{Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"sequence":  {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].Sequence}},
							"score":     {Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i].Score}},
							"token":     {Kind: &structpb.Value_NumberValue{NumberValue: float64(outputArr[i].Token)}},
							"token_str": {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].TokenStr}},
						},
					},
				}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"masks": {Kind: &structpb.Value_ListValue{ListValue: &masks}}},
			}
			outputs = append(outputs, &output)
		case summarizationTask:
			inputStruct := SummarizationRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []SummarizationResponse{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			summaries := structpb.ListValue{}
			summaries.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				summaries.Values[i] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: outputArr[i].SummaryText}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"texts": {Kind: &structpb.Value_ListValue{ListValue: &summaries}}},
			}
			outputs = append(outputs, &output)
		case textClassificationTask:
			inputStruct := TextClassificationRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			nestedArr := [][]TextClassificationResponseLabel{}
			err = json.Unmarshal(resp, &nestedArr)
			if err != nil {
				return nil, err
			}
			if len(nestedArr) <= 0 {
				return nil, errors.New("invalid response")
			}
			outputArr := nestedArr[0]
			classes := structpb.ListValue{}
			classes.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				classes.Values[i] = &structpb.Value{Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"label": {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].Label}},
							"score": {Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i].Score}},
						},
					},
				}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"classes": {Kind: &structpb.Value_ListValue{ListValue: &classes}}},
			}
			outputs = append(outputs, &output)
		case tokenClassificationTask:
			inputStruct := TokenClassificationRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []TokenClassificationResponseEntity{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			classes := structpb.ListValue{}
			classes.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				classes.Values[i] = &structpb.Value{Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"entity_group": {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].EntityGroup}},
							"score":        {Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i].Score}},
							"word":         {Kind: &structpb.Value_StringValue{StringValue: outputArr[i].Word}},
							"start":        {Kind: &structpb.Value_NumberValue{NumberValue: float64(outputArr[i].Start)}},
							"end":          {Kind: &structpb.Value_NumberValue{NumberValue: float64(outputArr[i].End)}},
						},
					},
				}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"classes": {Kind: &structpb.Value_ListValue{ListValue: &classes}}},
			}
			outputs = append(outputs, &output)
		case translationTask:
			inputStruct := TranslationRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []TranslationResponse{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			translations := structpb.ListValue{}
			translations.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				translations.Values[i] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: outputArr[i].TranslationText}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"texts": {Kind: &structpb.Value_ListValue{ListValue: &translations}}},
			}
			outputs = append(outputs, &output)
		case zeroShotClassificationTask:
			inputStruct := ZeroShotRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(resp, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)
		case featureExtractionTask:
			inputStruct := FeatureExtractionRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			threeDArr := [][][]float64{}
			err = json.Unmarshal(resp, &threeDArr)
			if err != nil {
				return nil, err
			}
			if len(threeDArr) <= 0 {
				return nil, errors.New("invalid response")
			}
			nestedArr := threeDArr[0]
			features := structpb.ListValue{}
			features.Values = make([]*structpb.Value, len(nestedArr))
			for i, innerArr := range nestedArr {
				innerValues := make([]*structpb.Value, len(innerArr))
				for j := range innerArr {
					innerValues[j] = &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: innerArr[j]}}
				}
				features.Values[i] = &structpb.Value{Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: innerValues}}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"features": {Kind: &structpb.Value_ListValue{ListValue: &features}}},
			}
			outputs = append(outputs, &output)
		case questionAnsweringTask:
			inputStruct := QuestionAnsweringRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(resp, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)
		case tableQuestionAnsweringTask:
			inputStruct := TableQuestionAnsweringRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(resp, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)
		case sentenceSimilarityTask:
			inputStruct := SentenceSimilarityRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			outputArr := []float64{}
			err = json.Unmarshal(resp, &outputArr)
			if err != nil {
				return nil, err
			}
			scores := structpb.ListValue{}
			scores.Values = make([]*structpb.Value, len(outputArr))
			for i := range outputArr {
				scores.Values[i] = &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: outputArr[i]}}
			}
			output := structpb.Struct{
				Fields: map[string]*structpb.Value{"scores": {Kind: &structpb.Value_ListValue{ListValue: &scores}}},
			}
			outputs = append(outputs, &output)
		case conversationalTask:
			inputStruct := ConversationalRequest{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}
			jsonBody, _ := json.Marshal(inputStruct)
			resp, err := client.MakeHFAPIRequest(jsonBody, model)
			if err != nil {
				return nil, err
			}
			output := structpb.Struct{}
			err = protojson.Unmarshal(resp, &output)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &output)
		default:
			return nil, fmt.Errorf("not supported task: %s", task)
		}
	}
	if err := c.ValidateOutput(outputs, task); err != nil {
		return nil, err
	}
	return outputs, nil
}

func (c *Connection) Test() (connectorPB.ConnectorResource_State, error) {
	client := NewClient(c.getAPIKey())
	return client.GetConnectionState()
}
