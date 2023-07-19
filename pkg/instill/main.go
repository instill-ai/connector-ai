package instill

import (
	"bytes"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	venderName   = "instillModel"
	getModelPath = "/v1alpha/models/"
	reqTimeout   = time.Second * 60
)

var (
	//go:embed config/seed/definitions.json
	definitionJSON    []byte
	once              sync.Once
	connector         base.IConnector
	connectorStateMap = map[string]connectorPB.Connector_State{
		"STATE_UNSPECIFIED": connectorPB.Connector_STATE_UNSPECIFIED,
		"STATE_OFFLINE":     connectorPB.Connector_STATE_DISCONNECTED,
		"STATE_ONLINE":      connectorPB.Connector_STATE_CONNECTED,
		"STATE_ERROR":       connectorPB.Connector_STATE_ERROR,
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
	Model *Model `json:"model"`
}

type Model struct {
	Name            string    `json:"name"`
	UID             string    `json:"uid"`
	ID              string    `json:"id"`
	Description     string    `json:"description"`
	ModelDefinition string    `json:"model_definition"`
	Configuration   any       `json:"configuration"`
	Task            string    `json:"task"`
	State           string    `json:"state"`
	Visibility      string    `json:"visibility"`
	User            string    `json:"user"`
	CreateTime      time.Time `json:"create_time"`
	UpdateTime      time.Time `json:"update_time"`
}

// Client represents an Instill Model client
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

// NewClient initializes a new Instill model client
func (c *Connection) NewClient() (*Client, error) {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	return &Client{APIKey: c.getAPIKey(), HTTPClient: &http.Client{Timeout: reqTimeout, Transport: tr}}, nil
}

// sendReq is responsible for making the http request with to given URL, method, and params and unmarshalling the response into given object.
func (c *Client) sendReq(reqURL, method string, params interface{}, respObj interface{}) (err error) {

	var req *http.Request
	data := []byte{}
	if params == nil {
		req, err = http.NewRequest(method, reqURL, nil)
		if err != nil {
			return err
		}
	} else {
		data, err = json.Marshal(params)
		if err != nil {
			return err
		}

		req, err = http.NewRequest(method, reqURL, bytes.NewBuffer(data))
		if err != nil {
			return err
		}
	}

	if c.APIKey != "" {
		req.Header.Add("Authorization", "Bearer "+c.APIKey)
	}
	http.DefaultClient.Timeout = reqTimeout
	res, err := c.HTTPClient.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil || res == nil {
		err = fmt.Errorf("error occurred: %v, while calling URL: %s, request body: %s", err, reqURL, data)
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
	return c.config.GetFields()["api_token"].GetStringValue()
}

func (c *Connection) getServerURL() string {
	serverUrl := c.config.GetFields()["server_url"].GetStringValue()
	if strings.HasPrefix(serverUrl, "https://") {
		if len(strings.Split(serverUrl, ":")) == 2 {
			serverUrl = serverUrl + ":443"
		}
	} else if strings.HasPrefix(serverUrl, "http://") {
		if len(strings.Split(serverUrl, ":")) == 2 {
			serverUrl = serverUrl + ":80"
		}
	}
	return serverUrl
}

func (c *Connection) getModelID() string {
	return c.config.GetFields()["model_id"].GetStringValue()
}

func (c *Connection) getModel() (res *GetModelRes, err error) {
	modelID := c.getModelID()
	serverURL := c.getServerURL()
	c.client, err = c.NewClient()
	if err != nil {
		return res, err
	}
	reqURL := serverURL + getModelPath + modelID
	err = c.client.sendReq(reqURL, http.MethodGet, nil, &res)
	return res, err
}

func (c *Connection) Execute(inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	res, err := c.getModel()
	if err != nil || res == nil || res.Model == nil {
		return inputs, err
	}
	if len(inputs) <= 0 || inputs[0] == nil {
		return inputs, fmt.Errorf("invalid input: %v for model: %s", inputs, res.Model.Name)
	}

	gRPCCLient, gRPCCLientConn := initModelPublicServiceClient(c.getServerURL())
	if gRPCCLientConn != nil {
		defer gRPCCLientConn.Close()
	}
	var result []*connectorPB.DataPayload
	switch res.Model.Task {
	case connectorPB.Task_TASK_UNSPECIFIED.String():
		result, err = c.executeUnspecified(gRPCCLient, res.Model, inputs)
	case connectorPB.Task_TASK_CLASSIFICATION.String():
		result, err = c.executeImageClassification(gRPCCLient, res.Model, inputs)
	case connectorPB.Task_TASK_DETECTION.String():
		result, err = c.executeObjectDetection(gRPCCLient, res.Model, inputs)
	case connectorPB.Task_TASK_KEYPOINT.String():
		result, err = c.executeKeyPointDetection(gRPCCLient, res.Model, inputs)
	case connectorPB.Task_TASK_OCR.String():
		result, err = c.executeOCR(gRPCCLient, res.Model, inputs)
	case connectorPB.Task_TASK_INSTANCE_SEGMENTATION.String():
		result, err = c.executeInstanceSegmentation(gRPCCLient, res.Model, inputs)
	case connectorPB.Task_TASK_SEMANTIC_SEGMENTATION.String():
		result, err = c.executeSemanticSegmentation(gRPCCLient, res.Model, inputs)
	case connectorPB.Task_TASK_TEXT_TO_IMAGE.String():
		result, err = c.executeTextToImage(gRPCCLient, res.Model, inputs)
	case connectorPB.Task_TASK_TEXT_GENERATION.String():
		result, err = c.executeTextGeneration(gRPCCLient, res.Model, inputs)
	default:
		return inputs, fmt.Errorf("unsupported task: %s", res.Model.Task)
	}
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

func (c *Connection) GetTask() (connectorPB.Task, error) {
	res, err := c.getModel()
	if err != nil || res == nil || res.Model == nil {
		return connectorPB.Task_TASK_UNSPECIFIED, err
	}
	task, ok := connectorPB.Task_value[res.Model.Task]
	if !ok {
		return connectorPB.Task_TASK_UNSPECIFIED, fmt.Errorf("mapping not found for: %s", res.Model.Task)
	}
	return connectorPB.Task(task), nil
}
