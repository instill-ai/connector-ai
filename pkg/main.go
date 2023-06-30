package pkg

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/connector-ai/pkg/instill"
	"github.com/instill-ai/connector-ai/pkg/stabilityai"
	"github.com/instill-ai/connector/pkg/base"
)

var once sync.Once
var connector base.IConnector

type Connector struct {
	base.BaseConnector
	stabilityAIConnector base.IConnector
	instillConnector     base.IConnector
}

type ConnectorOptions struct {
	StabilityAI stabilityai.ConnectorOptions
	Instill     instill.ConnectorOptions
}

func Init(logger *zap.Logger, options ConnectorOptions) base.IConnector {
	once.Do(func() {
		stabilityAIConnector := stabilityai.Init(logger, options.StabilityAI)
		instillConnector := instill.Init(logger, options.Instill)
		connector = &Connector{
			BaseConnector:        base.BaseConnector{Logger: logger},
			stabilityAIConnector: stabilityAIConnector,
			instillConnector:     instillConnector,
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
		for _, uid := range instillConnector.ListConnectorDefinitionUids() {
			def, err := instillConnector.GetConnectorDefinitionByUid(uid)
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
	case c.instillConnector.HasUid(defUid):
		return c.instillConnector.CreateConnection(defUid, config, logger)
	default:
		return nil, fmt.Errorf("no aiConnector uid: %s", defUid)
	}
}
