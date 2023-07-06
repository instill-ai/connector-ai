package instill

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeImageClassification(model *Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, model.Name)
	}
	dataPayload := inputs[0]
	if len(dataPayload.Images) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", *dataPayload, model.Name)
	}
	base64Str, err := encodeToBase64(dataPayload.Images[0])
	if err != nil {
		return nil, fmt.Errorf("invalid image string: %v for model: %s", dataPayload.Images[0], model.Name)
	}
	modelInput := &modelPB.TaskInput_Classification{
		Classification: &modelPB.ClassificationInput{
			Type: &modelPB.ClassificationInput_ImageBase64{ImageBase64: base64Str},
		},
	}
	req := modelPB.TriggerModelRequest{
		Name:       model.Name,
		TaskInputs: []*modelPB.TaskInput{{Input: modelInput}},
	}
	if c.client == nil || c.client.GRPCClient == nil {
		return nil, fmt.Errorf("client not setup: %v", c.client)
	}
	res, err := c.client.GRPCClient.TriggerModel(context.Background(), &req)
	if err != nil || res == nil {
		return nil, err
	}
	output := res.GetTaskOutputs()
	if len(output) <= 0 {
		return nil, fmt.Errorf("invalid output: %v for model: %s", output, model.Name)
	}
	imgClassificationOp := output[0].GetClassification()
	if imgClassificationOp == nil {
		return nil, fmt.Errorf("invalid output: %v for model: %s", imgClassificationOp, model.Name)
	}
	inputs[0].StructuredData = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"category": {Kind: &structpb.Value_StringValue{StringValue: imgClassificationOp.Category}},
			"score":    {Kind: &structpb.Value_NumberValue{NumberValue: float64(imgClassificationOp.Score)}},
		},
	}
	return inputs, nil
}

// encode the bytes to base64 string if not already encoded
func encodeToBase64(input []byte) (string, error) {
	if len(input) <= 0 {
		return "", fmt.Errorf("invalid byte value :%v", input)
	}
	return base64.StdEncoding.EncodeToString(input), nil
}
