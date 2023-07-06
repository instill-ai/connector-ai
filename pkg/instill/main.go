package instill

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

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

const (
	venderName       = "instillModel"
	getModelPath     = "/v1alpha/models/"
	instillCloudHost = "https://api.instill.tech/model"
	instillCloudPort = 443
	reqTimeout       = time.Second * 60
)

var (
	//go:embed config/seed/definitions.json
	definitionJSON    []byte
	once              sync.Once
	connector         base.IConnector
	connectorStateMap = map[modelPB.Model_State]connectorPB.Connector_State{
		modelPB.Model_STATE_UNSPECIFIED: connectorPB.Connector_STATE_UNSPECIFIED,
		modelPB.Model_STATE_OFFLINE:     connectorPB.Connector_STATE_DISCONNECTED,
		modelPB.Model_STATE_ONLINE:      connectorPB.Connector_STATE_CONNECTED,
		modelPB.Model_STATE_ERROR:       connectorPB.Connector_STATE_ERROR,
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
	client    *Client
}

type GetModelRes struct {
	Model *modelPB.Model `json:"model"`
}

// Client represents an Instill Model client
type Client struct {
	APIKey     string
	HTTPClient HTTPClient
	GRPCClient modelPB.ModelPublicServiceClient
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

// NewClient initializes a new Instill model client
func (c *Connection) NewClient() (*Client, error) {
	apiKey := ""
	serverURL := c.getServerURL()
	if serverURL == instillCloudHost {
		apiKey = c.getAPIKey()
		if apiKey == "" {
			return nil, fmt.Errorf("api key cannot be empty for instill cloud")
		}
	}
	gRPCCLient, _ := initModelPublicServiceClient(serverURL)
	return &Client{APIKey: apiKey, HTTPClient: &http.Client{Timeout: reqTimeout}, GRPCClient: gRPCCLient}, nil
}

// sendReq is responsible for making the http request with to given URL, method, and params and unmarshalling the response into given object.
func (c *Client) sendReq(reqURL, method string, params interface{}, respObj interface{}) (err error) {
	data, _ := json.Marshal(params)
	req, _ := http.NewRequest(method, reqURL, bytes.NewBuffer(data))
	if c.APIKey != "" {
		req.Header.Add("Authorization", "Bearer "+c.APIKey)
	}
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

func (c *Connection) getAPIKey() string {
	return fmt.Sprintf("%s", c.config.GetFields()["api_key"].GetStringValue())
}

func (c *Connection) getServerURL() string {
	return fmt.Sprintf("%s", c.config.GetFields()["server_url"].GetStringValue())
}

func (c *Connection) getModelID() string {
	return fmt.Sprintf("%s", c.config.GetFields()["model_id"].GetStringValue())
}

func (c *Connection) getModel() (res *GetModelRes, err error) {
	modelID := c.getModelID()
	serverURL := c.getServerURL()
	c.client, err = c.NewClient()
	if err != nil {
		return res, err
	}
	reqURL := serverURL + getModelPath + modelID
	err = c.client.sendReq(reqURL, http.MethodGet, nil, res)
	return res, err
}

func (c *Connection) Execute(inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	fmt.Printf("inputs: %v", inputs)
	res, err := c.getModel()
	fmt.Printf("after getModel, res: %v, err: %v", res, err)
	if err != nil || res == nil || res.Model == nil {
		return inputs, err
	}
	if len(inputs) <= 0 || inputs[0] == nil {
		return inputs, fmt.Errorf("invalid input: %v for model: %s", inputs, res.Model.Name)
	}
	fmt.Printf("input[0]: %v", *inputs[0])
	fmt.Printf("res.Model.Task: %v", res.Model.Task)
	var result []*connectorPB.DataPayload
	switch res.Model.Task {
	case modelPB.Model_TASK_UNSPECIFIED:
		result, err = c.executeUnspecified(res.Model, inputs)
	case modelPB.Model_TASK_CLASSIFICATION:
		result, err = c.executeImageClassification(res.Model, inputs)
	case modelPB.Model_TASK_DETECTION:
		result, err = c.executeObjectDetection(res.Model, inputs)
	case modelPB.Model_TASK_KEYPOINT:
		result, err = c.executeKeyPointDetection(res.Model, inputs)
	case modelPB.Model_TASK_OCR:
		result, err = c.executeOCR(res.Model, inputs)
	case modelPB.Model_TASK_INSTANCE_SEGMENTATION:
		result, err = c.executeInstanceSegmentation(res.Model, inputs)
	case modelPB.Model_TASK_SEMANTIC_SEGMENTATION:
		result, err = c.executeSemanticSegmentation(res.Model, inputs)
	case modelPB.Model_TASK_TEXT_TO_IMAGE:
		result, err = c.executeTextToImage(res.Model, inputs)
	case modelPB.Model_TASK_TEXT_GENERATION:
		result, err = c.executeTextGeneration(res.Model, inputs)
	default:
		return inputs, fmt.Errorf("unsupported task: %s", res.Model.Task)
	}
	fmt.Printf("result: %v, err:%v", result, err)
	fmt.Printf("results[0]: %v", *result[0])
	return result, err
}

func (c *Connection) Test() (connectorPB.Connector_State, error) {
	res, err := c.getModel()
	if err != nil || res == nil || res.Model == nil {
		return connectorPB.Connector_STATE_UNSPECIFIED, err
	}
	st, ok := connectorStateMap[res.Model.State]
	if !ok {
		return connectorPB.Connector_STATE_UNSPECIFIED, fmt.Errorf("mapping not found for: %v", res.Model.State)
	}
	return st, nil
}

func (c *Connection) GetTaskName() (string, error) {
	res, err := c.getModel()
	if err != nil || res == nil || res.Model == nil {
		return modelPB.Model_TASK_UNSPECIFIED.String(), err
	}
	return res.Model.Task.String(), nil
}
