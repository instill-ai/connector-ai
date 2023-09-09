package instill

import (
	"crypto/tls"
	_ "embed"
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

	commonPB "github.com/instill-ai/protogen-go/common/task/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

const (
	venderName   = "instillModel"
	getModelPath = "/v1alpha/models"
	reqTimeout   = time.Second * 60
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
	client    *Client
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

// NewClient initializes a new Instill model client
func (c *Connection) NewClient() (*Client, error) {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	return &Client{APIKey: c.getAPIKey(), HTTPClient: &http.Client{Timeout: reqTimeout, Transport: tr}}, nil
}

func (c *Connection) getAPIKey() string {
	return c.Config.GetFields()["api_token"].GetStringValue()
}

func (c *Connection) getServerURL() string {
	serverUrl := c.Config.GetFields()["server_url"].GetStringValue()
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

func (c *Connection) getModels() (err error) {
	serverURL := c.getServerURL()
	c.client, err = c.NewClient()
	if err != nil {
		return err
	}
	reqURL := serverURL + getModelPath
	err = c.client.sendReq(reqURL, http.MethodGet, nil)
	return err
}

// sendReq is responsible for making the http request with to given URL, method, and params and unmarshalling the response into given object.
func (c *Client) sendReq(reqURL, method string, params interface{}) (err error) {
	req, _ := http.NewRequest(method, reqURL, nil)
	if c.APIKey != "" {
		req.Header.Add("Authorization", "Bearer "+c.APIKey)
	}
	http.DefaultClient.Timeout = reqTimeout
	res, err := c.HTTPClient.Do(req)

	if err != nil || res == nil {
		err = fmt.Errorf("error occurred: %v, while calling URL: %s", err, reqURL)
		return
	}
	defer res.Body.Close()
	bytes, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("non-200 status code: %d, while calling URL: %s, response body: %s", res.StatusCode, reqURL, bytes)
		return
	}
	return
}

func (c *Connection) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	var err error
	c.client, err = c.NewClient()
	if err != nil {
		return nil, err
	}

	if len(inputs) <= 0 || inputs[0] == nil {
		return inputs, fmt.Errorf("invalid input")
	}

	gRPCCLient, gRPCCLientConn := initModelPublicServiceClient(c.getServerURL())
	if gRPCCLientConn != nil {
		defer gRPCCLientConn.Close()
	}

	task := inputs[0].GetFields()["task"].GetStringValue()
	for _, input := range inputs {
		if input.GetFields()["task"].GetStringValue() != task {
			return nil, fmt.Errorf("each input should be the same task")
		}
	}

	if err := c.ValidateInput(inputs, task); err != nil {
		return nil, err
	}

	modelName := inputs[0].GetFields()["model_name"].GetStringValue()

	var result []*structpb.Struct
	switch task {
	case commonPB.Task_TASK_UNSPECIFIED.String():
		result, err = c.executeUnspecified(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_CLASSIFICATION.String():
		result, err = c.executeImageClassification(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_DETECTION.String():
		result, err = c.executeObjectDetection(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_KEYPOINT.String():
		result, err = c.executeKeyPointDetection(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_OCR.String():
		result, err = c.executeOCR(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_INSTANCE_SEGMENTATION.String():
		result, err = c.executeInstanceSegmentation(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_SEMANTIC_SEGMENTATION.String():
		result, err = c.executeSemanticSegmentation(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_TEXT_TO_IMAGE.String():
		result, err = c.executeTextToImage(gRPCCLient, modelName, inputs)
	case commonPB.Task_TASK_TEXT_GENERATION.String():
		result, err = c.executeTextGeneration(gRPCCLient, modelName, inputs)
	default:
		return inputs, fmt.Errorf("unsupported task: %s", task)
	}
	if err := c.ValidateOutput(result, task); err != nil {
		return nil, err
	}
	return result, err
}

func (c *Connection) Test() (connectorPB.ConnectorResource_State, error) {
	err := c.getModels()
	if err != nil {
		return connectorPB.ConnectorResource_STATE_ERROR, err
	}
	return connectorPB.ConnectorResource_STATE_CONNECTED, nil
}
