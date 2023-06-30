package stabilityai

import (
	"bytes"
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
	venderName       = "stabilityAI"
	host             = "https://api.stability.ai"
	jsonMimeType     = "application/json"
	reqTimeout       = time.Second * 60
	textToImageTask  = "Text to Image"
	imageToImageTask = "Image to Image"
)

//go:embed config/definitions.json
var definitionJSON []byte

var (
	once      sync.Once
	connector base.IConnector
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

// Client represents a Stability AI client
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
	return Client{APIKey: apiKey, HTTPClient: &http.Client{Timeout: reqTimeout}}
}

// sendReq is responsible for making the http request with to given URL, method, and params and unmarshalling the response into given object.
func (c *Client) sendReq(reqURL, method string, params interface{}, respObj interface{}) (err error) {
	data, _ := json.Marshal(params)
	req, _ := http.NewRequest(method, reqURL, bytes.NewBuffer(data))
	req.Header.Add("Content-Type", jsonMimeType)
	req.Header.Add("Accept", jsonMimeType)
	req.Header.Add("Authorization", "Bearer "+c.APIKey)
	http.DefaultClient.Timeout = reqTimeout
	res, err := c.HTTPClient.Do(req)
	if err != nil || res == nil {
		err = fmt.Errorf("error occurred: %v, while calling URL: %s, request body: %s", err, reqURL, data)
		return
	}
	defer res.Body.Close()
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

func (con *Connection) getAPIKey() string {
	return fmt.Sprintf("%s", con.config.GetFields()["api_key"].GetStringValue())
}

func (con *Connection) getTask() string {
	return fmt.Sprintf("%s", con.config.GetFields()["task"].GetStringValue())
}

func (con *Connection) getEngine() string {
	return fmt.Sprintf("%s", con.config.GetFields()["engine"].GetStringValue())
}

func (c *Connection) Execute(inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	engine := c.getEngine()
	task := c.getTask()
	client := NewClient(c.getAPIKey())
	switch task {
	case textToImageTask:
		for i, dataPayload := range inputs {
			noOfPrompts := len(dataPayload.Texts)
			if noOfPrompts <= 0 {
				return inputs, fmt.Errorf("no text promts given")
			}
			req := TextToImageReq{
				CFGScale:           dataPayload.GetMetadata().GetFields()["cfg_scale"].GetNumberValue(),
				ClipGuidancePreset: dataPayload.GetMetadata().GetFields()["clip_guidance_preset"].GetStringValue(),
				Sampler:            dataPayload.GetMetadata().GetFields()["sampler"].GetStringValue(),
				Samples:            uint32(dataPayload.GetMetadata().GetFields()["samples"].GetNumberValue()),
				Seed:               uint32(dataPayload.GetMetadata().GetFields()["seed"].GetNumberValue()),
				Steps:              uint32(dataPayload.GetMetadata().GetFields()["steps"].GetNumberValue()),
				StylePreset:        dataPayload.GetMetadata().GetFields()["style_preset"].GetStringValue(),
				Height:             uint32(dataPayload.GetMetadata().GetFields()["height"].GetNumberValue()),
				Width:              uint32(dataPayload.GetMetadata().GetFields()["weight"].GetNumberValue()),
			}
			req.TextPrompts = make([]TextPrompt, 0, len(dataPayload.Texts))
			for _, t := range dataPayload.Texts {
				req.TextPrompts = append(req.TextPrompts, TextPrompt{Text: t})
			}
			images, err := client.GenerateImageFromText(req, engine)
			if err != nil {
				return inputs, err
			}
			// use inputs[i] instead of dataPayload to modify source data
			inputs[i].Images = make([][]byte, 0, len(images))
			for _, image := range images {
				inputs[i].Images = append(dataPayload.Images, []byte(image.Base64))
			}
		}
	case imageToImageTask:
		for i, dataPayload := range inputs {
			noOfPrompts := len(dataPayload.Texts)
			if noOfPrompts <= 0 {
				return inputs, fmt.Errorf("no text promts given")
			}
			noOfImages := len(dataPayload.Images)
			if noOfImages <= 0 {
				return inputs, fmt.Errorf("no initial images given")
			}
			req := ImageToImageReq{
				InitImage:          string(dataPayload.Images[0]),
				CFGScale:           dataPayload.GetMetadata().GetFields()["cfg_scale"].GetNumberValue(),
				ClipGuidancePreset: dataPayload.GetMetadata().GetFields()["clip_guidance_preset"].GetStringValue(),
				Sampler:            dataPayload.GetMetadata().GetFields()["sampler"].GetStringValue(),
				Samples:            uint32(dataPayload.GetMetadata().GetFields()["samples"].GetNumberValue()),
				Seed:               uint32(dataPayload.GetMetadata().GetFields()["seed"].GetNumberValue()),
				Steps:              uint32(dataPayload.GetMetadata().GetFields()["steps"].GetNumberValue()),
				StylePreset:        dataPayload.GetMetadata().GetFields()["style_preset"].GetStringValue(),
				InitImageMode:      dataPayload.GetMetadata().GetFields()["init_image_mode"].GetStringValue(),
				ImageStrength:      dataPayload.GetMetadata().GetFields()["image_strength"].GetNumberValue(),
			}
			req.TextPrompts = make([]TextPrompt, 0, len(dataPayload.Texts))
			for _, t := range dataPayload.Texts {
				req.TextPrompts = append(req.TextPrompts, TextPrompt{Text: t})
			}
			images, err := client.GenerateImageFromImage(req, engine)
			if err != nil {
				return inputs, err
			}
			// use inputs[i] instead of dataPayload to modify source data
			inputs[i].Images = make([][]byte, 0, len(images))
			for _, image := range images {
				inputs[i].Images = append(dataPayload.Images, []byte(image.Base64))
			}
		}
	default:
		return nil, fmt.Errorf("not supported task: %s", task)
	}
	return inputs, nil
}

func (c *Connection) Test() (connectorPB.Connector_State, error) {
	client := NewClient(c.getAPIKey())
	engines, err := client.ListEngines()
	if err != nil || len(engines) == 0 {
		return connectorPB.Connector_STATE_ERROR, err
	}
	return connectorPB.Connector_STATE_CONNECTED, nil
}

func (con *Connection) GetTaskName() (string, error) {
	// TODO: load from configuration
	return "TASK_TEXT_TO_IMAGE", nil
}
