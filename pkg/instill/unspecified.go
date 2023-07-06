package instill

import (
	"fmt"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeUnspecified(model *Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, model.Name)
	}
	//TODO: figure out what to do here?
	/*
		modelInput := &modelPB.TaskInput_Unspecified{
			Unspecified: &modelPB.UnspecifiedInput{
				RawInputs: []*structpb.Struct{},
			},
		}
	*/
	return inputs, nil
}
