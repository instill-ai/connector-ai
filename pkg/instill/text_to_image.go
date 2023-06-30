package instill_model

import (
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeTextToImage(model *modelPB.Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	//TODO: execute model, fetch output and convert to standard format
	return inputs, nil
}
