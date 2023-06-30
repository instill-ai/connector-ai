package instill

import (
	"sync"

	_ "embed"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/connector/pkg/base"
	"github.com/instill-ai/connector/pkg/configLoader"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

// Note: this is a dummy connector

const vendorName = "instill"

//go:embed config/seed/definitions.json
var definitionJson []byte

var once sync.Once
var connector base.IConnector

type Connector struct {
	base.BaseConnector
	options ConnectorOptions
}

type Connection struct {
	base.BaseConnection
	config *structpb.Struct
}

type ConnectorOptions struct{}

func Init(logger *zap.Logger, options ConnectorOptions) base.IConnector {
	once.Do(func() {
		loader := configLoader.InitJSONSchema(logger)
		connDefs, err := loader.Load(vendorName, connectorPB.ConnectorType_CONNECTOR_TYPE_AI, definitionJson)

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
		config:         config,
	}, nil
}

func (con *Connection) Execute(inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {

	mock_data := []byte(`
	{
		"detection": {
			"objects": [
				{
					"score": 0.9474398,
					"category": "bear",
					"boundingBox": {
						"top": 455,
						"left": 1372,
						"width": 1300,
						"height": 2178
					}
				}
			]
	  	}
	}
	`)

	structuredData := &structpb.Struct{}
	protojson.Unmarshal(mock_data, structuredData)
	outputs := []*connectorPB.DataPayload{}
	for idx := range inputs {
		output := &connectorPB.DataPayload{
			DataMappingIndex: inputs[idx].DataMappingIndex,
			StructuredData:   structuredData,
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func (con *Connection) Test() (connectorPB.Connector_State, error) {
	return connectorPB.Connector_STATE_CONNECTED, nil
}

func (con *Connection) GetTaskName() (string, error) {
	return "TASK_DETECTION", nil
}
