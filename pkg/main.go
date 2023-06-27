package pkg

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/connector-model/pkg/stabilityai"
	"github.com/instill-ai/connector/pkg/base"
)

var once sync.Once
var connector base.IConnector

type Connector struct {
	base.BaseConnector
	stabilityAIConnector base.IConnector
}

type ConnectorOptions struct {
	StabilityAI stabilityai.ConnectorOptions
}

func Init(logger *zap.Logger, options ConnectorOptions) base.IConnector {
	once.Do(func() {
		stabilityAIConnector := stabilityai.Init(logger, options.StabilityAI)
		connector = &Connector{
			BaseConnector:        base.BaseConnector{Logger: logger},
			stabilityAIConnector: stabilityAIConnector,
		}

		for _, uid := range stabilityAIConnector.ListConnectorDefinitionUids() {
			def, err := stabilityAIConnector.GetConnectorDefinitionByUid(uid)
			if err != nil {
				logger.Error(err.Error())
			}
			err = connector.AddConnectorDefinition(uid, def.GetId(), def)
			if err != nil {
				logger.Warn(err.Error())
			}
		}
	})
	return connector
}

func (c *Connector) CreateConnection(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (base.IConnection, error) {
	switch {
	case c.stabilityAIConnector.HasUid(defUid):
		return c.stabilityAIConnector.CreateConnection(defUid, config, logger)
	default:
		return nil, fmt.Errorf("no destinationConnector uid: %s", defUid)
	}
}
