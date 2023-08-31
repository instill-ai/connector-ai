package pkg

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/connector-ai/pkg/huggingface"
	"github.com/instill-ai/connector-ai/pkg/instill"
	"github.com/instill-ai/connector-ai/pkg/openai"
	"github.com/instill-ai/connector-ai/pkg/stabilityai"
	"github.com/instill-ai/connector/pkg/base"
)

var once sync.Once
var connector base.IConnector

type Connector struct {
	base.BaseConnector
	stabilityAIConnector base.IConnector
	instillConnector     base.IConnector
	openAIConnector      base.IConnector
	huggingFaceConnector base.IConnector
}

type ConnectorOptions struct {
	StabilityAI stabilityai.ConnectorOptions
	Instill     instill.ConnectorOptions
	OpenAI      openai.ConnectorOptions
	HuggingFace huggingface.ConnectorOptions
}

func Init(logger *zap.Logger, options ConnectorOptions) base.IConnector {
	once.Do(func() {
		stabilityAIConnector := stabilityai.Init(logger, options.StabilityAI)
		instillConnector := instill.Init(logger, options.Instill)
		openAIConnector := openai.Init(logger, options.OpenAI)
		huggingFaceConnector := huggingface.Init(logger, options.HuggingFace)
		connector = &Connector{
			BaseConnector:        base.BaseConnector{Logger: logger},
			stabilityAIConnector: stabilityAIConnector,
			instillConnector:     instillConnector,
			openAIConnector:      openAIConnector,
			huggingFaceConnector: huggingFaceConnector,
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
		for _, uid := range openAIConnector.ListConnectorDefinitionUids() {
			def, err := openAIConnector.GetConnectorDefinitionByUid(uid)
			if err != nil {
				logger.Error(err.Error())
			}
			err = connector.AddConnectorDefinition(uid, def.GetId(), def)
			if err != nil {
				logger.Warn(err.Error())
			}
		}
		for _, uid := range huggingFaceConnector.ListConnectorDefinitionUids() {
			def, err := huggingFaceConnector.GetConnectorDefinitionByUid(uid)
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
	case c.openAIConnector.HasUid(defUid):
		return c.openAIConnector.CreateConnection(defUid, config, logger)
	case c.huggingFaceConnector.HasUid(defUid):
		return c.huggingFaceConnector.CreateConnection(defUid, config, logger)
	default:
		return nil, fmt.Errorf("no aiConnector uid: %s", defUid)
	}
}
