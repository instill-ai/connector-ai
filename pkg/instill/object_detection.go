package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (c *Connection) executeObjectDetection(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	taskInputs := []*modelPB.TaskInput{}
	for _, input := range inputs {
		inputJson, err := protojson.Marshal(input)
		if err != nil {
			return nil, err
		}

		detectionInput := &modelPB.DetectionInput{}
		err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(inputJson, detectionInput)
		if err != nil {
			return nil, err
		}

		modelInput := &modelPB.TaskInput_Detection{
			Detection: detectionInput,
		}
		taskInputs = append(taskInputs, &modelPB.TaskInput{Input: modelInput})
	}

	req := modelPB.TriggerUserModelRequest{
		Name:       modelName,
		TaskInputs: taskInputs,
	}
	if c.client == nil || grpcClient == nil {
		return nil, fmt.Errorf("client not setup: %v", c.client)
	}

	md := metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", c.getAPIKey()))
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	res, err := grpcClient.TriggerUserModel(ctx, &req)
	if err != nil || res == nil {
		return nil, err
	}

	taskOutputs := res.GetTaskOutputs()
	if len(taskOutputs) <= 0 {
		return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, modelName)
	}

	outputs := []*structpb.Struct{}
	for idx := range inputs {
		objDetectionOutput := taskOutputs[idx].GetDetection()
		if objDetectionOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", objDetectionOutput, modelName)
		}
		outputJson, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(objDetectionOutput)
		if err != nil {
			return nil, err
		}
		output := &structpb.Struct{}
		err = protojson.Unmarshal(outputJson, output)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)

	}

	return outputs, nil
}
