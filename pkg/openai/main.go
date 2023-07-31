package openai

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/connector/pkg/base"
	"github.com/instill-ai/connector/pkg/configLoader"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

const (
	venderName            = "openAI"
	host                  = "https://api.openai.com"
	jsonMimeType          = "application/json"
	reqTimeout            = time.Second * 60 * 5
	textGenerationTask    = "Text Generation"
	textEmbeddingsTask    = "Text Embeddings"
	speechRecognitionTask = "Speech Recognition"
)

var (
	//go:embed config/seed/definitions.json
	definitionJSON []byte
	once           sync.Once
	connector      base.IConnector
	taskToNameMap  = map[string]connectorPB.Task{
		textGenerationTask: connectorPB.Task_TASK_TEXT_GENERATION,
		textEmbeddingsTask: connectorPB.Task_TASK_TEXT_EMBEDDINGS,
		// TODO: update this once speech recognition task is added to connectorPB
		speechRecognitionTask: connectorPB.Task_TASK_UNSPECIFIED,
	}
)

type ConnectorOptions struct{}

type Connector struct {
	base.BaseConnector
	options ConnectorOptions
}

type Connection struct {
	base.BaseConnection
	connector *Connector
	defUid    uuid.UUID
	config    *structpb.Struct
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
	return &Connection{
		BaseConnection: base.BaseConnection{Logger: logger},
		connector:      c,
		defUid:         defUid,
		config:         config,
	}, nil
}

// NewClient initializes a new Stability AI client
func NewClient(apiKey string) Client {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	return Client{APIKey: apiKey, HTTPClient: &http.Client{Timeout: reqTimeout, Transport: tr}}
}

// sendReq is responsible for making the http request with to given URL, method, and params and unmarshalling the response into given object.
// func (c *Client) sendReq(reqURL, method string, params interface{}, respObj interface{}) (err error) {
func (c *Client) sendReq(reqURL, method, contentType string, data io.Reader, respObj interface{}) (err error) {
	req, _ := http.NewRequest(method, reqURL, data)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", jsonMimeType)
	req.Header.Add("Authorization", "Bearer "+c.APIKey)
	http.DefaultClient.Timeout = reqTimeout
	res, err := c.HTTPClient.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil || res == nil {
		err = fmt.Errorf("error occurred: %v, while calling URL: %s", err, reqURL)
		return
	}
	bytes, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("non-200 status code: %d, while calling URL: %s, response body: %s", res.StatusCode, reqURL, bytes)
		return
	}
	if err = json.Unmarshal(bytes, &respObj); err != nil {
		err = fmt.Errorf("error in json decode: %s, while calling URL: %s, response body: %s", err, reqURL, bytes)
	}
	return
}

func (c *Connection) getAPIKey() string {
	return c.config.GetFields()["api_key"].GetStringValue()
}

func (c *Connection) getTask() string {
	return c.config.GetFields()["task"].GetStringValue()
}

func (c *Connection) getModel() string {
	return c.config.GetFields()["model"].GetStringValue()
}

func (c *Connection) Execute(inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	task := c.getTask()
	client := NewClient(c.getAPIKey())

	outputs := []*connectorPB.DataPayload{}
	switch task {
	case textGenerationTask:
		for i, dataPayload := range inputs {
			noOfPrompts := len(dataPayload.Texts)
			if noOfPrompts <= 0 {
				return inputs, fmt.Errorf("no text prompts given")
			}
			messages := make([]Message, 0, noOfPrompts)
			for _, t := range dataPayload.Texts {
				messages = append(messages, Message{Role: "user", Content: t})
			}
			req := TextCompletionReq{
				Messages:         messages,
				Model:            c.getModel(),
				MaxTokens:        int(dataPayload.GetMetadata().GetFields()["max_tokens"].GetNumberValue()),
				Temperature:      float32(dataPayload.GetMetadata().GetFields()["temperature"].GetNumberValue()),
				TopP:             float32(dataPayload.GetMetadata().GetFields()["top_p"].GetNumberValue()),
				N:                int(dataPayload.GetMetadata().GetFields()["n"].GetNumberValue()),
				Stream:           dataPayload.GetMetadata().GetFields()["stream"].GetBoolValue(),
				Stop:             dataPayload.GetMetadata().GetFields()["stop"].GetStringValue(),
				PresencePenalty:  float32(dataPayload.GetMetadata().GetFields()["presence_penalty"].GetNumberValue()),
				FrequencyPenalty: float32(dataPayload.GetMetadata().GetFields()["frequency_penalty"].GetNumberValue()),
			}
			resp, err := client.GenerateTextCompletion(req)
			if err != nil {
				return inputs, err
			}
			outputTexts := make([]string, 0, len(resp.Choices))
			for _, c := range resp.Choices {
				outputTexts = append(outputTexts, c.Message.Content)
			}
			outputs = append(outputs, &connectorPB.DataPayload{
				DataMappingIndex: inputs[i].DataMappingIndex,
				Texts:            outputTexts,
			})
		}
	case textEmbeddingsTask:
		for i, dataPayload := range inputs {
			noOfPrompts := len(dataPayload.Texts)
			if noOfPrompts <= 0 {
				return inputs, fmt.Errorf("no text prompts given")
			}
			req := TextEmbeddingsReq{
				Model: c.getModel(),
				Input: dataPayload.Texts,
			}
			resp, err := client.GenerateTextEmbeddings(req)
			if err != nil {
				return inputs, err
			}
			values := make([]*structpb.Value, 0, len(resp.Data))
			for _, em := range resp.Data {
				embeddingValues := make([]*structpb.Value, 0, len(em.Embedding))
				for _, v := range em.Embedding {
					embeddingValues = append(embeddingValues, &structpb.Value{Kind: &structpb.Value_NumberValue{NumberValue: v}})
				}
				obj := &structpb.Value{
					Kind: &structpb.Value_StructValue{
						StructValue: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"index":     {Kind: &structpb.Value_NumberValue{NumberValue: float64(em.Index)}},
								"object":    {Kind: &structpb.Value_StringValue{StringValue: em.Object}},
								"embedding": {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: embeddingValues}}},
							},
						},
					},
				}
				values = append(values, obj)
			}
			outputs = append(outputs, &connectorPB.DataPayload{
				DataMappingIndex: inputs[i].DataMappingIndex,
				StructuredData: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"embeddings": {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: values}}},
					},
				},
			})
		}
	case speechRecognitionTask:
		for i, dataPayload := range inputs {
			noOfAudios := len(dataPayload.Audios)
			if noOfAudios <= 0 {
				return inputs, fmt.Errorf("no audios given")
			}
			req := AudioTranscriptionReq{
				File:        dataPayload.Audios[0],
				Model:       c.getModel(),
				Language:    dataPayload.GetMetadata().GetFields()["language"].GetStringValue(),
				Temperature: dataPayload.GetMetadata().GetFields()["temperature"].GetNumberValue(),
			}
			if len(dataPayload.Texts) > 0 {
				req.Prompt = dataPayload.Texts[0]
			}
			resp, err := client.GenerateAudioTranscriptions(req)
			if err != nil {
				return inputs, err
			}
			outputs = append(outputs, &connectorPB.DataPayload{
				DataMappingIndex: inputs[i].DataMappingIndex,
				Texts:            []string{resp.Text},
			})
		}
	default:
		return nil, fmt.Errorf("not supported task: %s", task)
	}
	return outputs, nil
}

func (c *Connection) Test() (connectorPB.Connector_State, error) {
	client := NewClient(c.getAPIKey())
	models, err := client.ListModels()
	if err != nil {
		return connectorPB.Connector_STATE_ERROR, err
	}
	if len(models.Data) == 0 {
		return connectorPB.Connector_STATE_DISCONNECTED, nil
	}
	return connectorPB.Connector_STATE_CONNECTED, nil
}

func (c *Connection) GetTask() (connectorPB.Task, error) {
	name, ok := taskToNameMap[c.getTask()]
	if !ok {
		name = connectorPB.Task_TASK_UNSPECIFIED
	}
	return name, nil
}
