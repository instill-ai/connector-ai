package instill

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeImageClassification(grpcClient modelPB.ModelPublicServiceClient, model *Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, model.Name)
	}

	tasklInputs := []*modelPB.TaskInput{}
	for idx := range inputs {
		dataPayload := inputs[idx]
		if len(dataPayload.Images) <= 0 {
			return nil, fmt.Errorf("invalid input: %v for model: %s", *dataPayload, model.Name)
		}
		base64Str, err := encodeToBase64(dataPayload.Images[0])
		if err != nil {
			return nil, fmt.Errorf("invalid image string: %v for model: %s", dataPayload.Images[0], model.Name)
		}
		taskInput := &modelPB.TaskInput_Classification{
			Classification: &modelPB.ClassificationInput{
				Type: &modelPB.ClassificationInput_ImageBase64{ImageBase64: base64Str},
			},
		}
		tasklInputs = append(tasklInputs, &modelPB.TaskInput{Input: taskInput})
	}

	req := modelPB.TriggerModelRequest{
		Name:       model.Name,
		TaskInputs: tasklInputs,
	}
	if c.client == nil || grpcClient == nil {
		return nil, fmt.Errorf("client not setup: %v", c.client)
	}
	md := metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", c.getAPIKey()))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	res, err := grpcClient.TriggerModel(ctx, &req)
	if err != nil || res == nil {
		return nil, err
	}
	taskOutputs := res.GetTaskOutputs()
	if len(taskOutputs) <= 0 {
		return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, model.Name)
	}
	outputs := []*connectorPB.DataPayload{}
	for idx := range inputs {
		imgClassificationOp := taskOutputs[idx].GetClassification()
		if imgClassificationOp == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", imgClassificationOp, model.Name)
		}
		outputs = append(outputs, &connectorPB.DataPayload{
			DataMappingIndex: inputs[idx].DataMappingIndex,
			StructuredData: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"category": {Kind: &structpb.Value_StringValue{StringValue: imgClassificationOp.Category}},
					"score":    {Kind: &structpb.Value_NumberValue{NumberValue: float64(imgClassificationOp.Score)}},
				},
			},
		})
	}
	return outputs, nil
}

// encode the bytes to base64 string if not already encoded
func encodeToBase64(input []byte) (string, error) {
	if len(input) <= 0 {
		return "", fmt.Errorf("invalid byte value :%v", input)
	}
	return base64.StdEncoding.EncodeToString(input), nil
}

// decode the base64 string to bytesp[]
func decodeFromBase64(b64str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(b64str)
}
