package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (c *Connection) executeImageClassification(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	tasklInputs := []*modelPB.TaskInput{}
	for _, input := range inputs {

		inputJson, err := protojson.Marshal(input)
		if err != nil {
			return nil, err
		}
		classificationInput := &modelPB.ClassificationInput{}
		protojson.Unmarshal(inputJson, classificationInput)

		taskInput := &modelPB.TaskInput_Classification{
			Classification: classificationInput,
		}
		tasklInputs = append(tasklInputs, &modelPB.TaskInput{Input: taskInput})
	}

	req := modelPB.TriggerModelRequest{
		Name:       modelName,
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
		return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, modelName)
	}
	outputs := []*structpb.Struct{}
	for idx := range inputs {
		imgClassificationOp := taskOutputs[idx].GetClassification()
		if imgClassificationOp == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", imgClassificationOp, modelName)
		}
		outputJson, err := protojson.Marshal(imgClassificationOp)
		if err != nil {
			return nil, err
		}
		output := &structpb.Struct{}
		protojson.Unmarshal(outputJson, output)
		outputs = append(outputs, output)
	}
	return outputs, nil
}
