package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (c *Connection) executeTextGeneration(grpcClient modelPB.ModelPublicServiceClient, modelName string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, modelName)
	}

	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		inputJson, err := protojson.Marshal(input)
		if err != nil {
			return nil, err
		}
		textGenerationInput := &modelPB.TextGenerationInput{}
		protojson.Unmarshal(inputJson, textGenerationInput)

		taskInput := &modelPB.TaskInput_TextGeneration{
			TextGeneration: textGenerationInput,
		}

		// only support batch 1
		req := modelPB.TriggerModelRequest{
			Name:       modelName,
			TaskInputs: []*modelPB.TaskInput{{Input: taskInput}},
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

		textGenOutput := taskOutputs[0].GetTextGeneration()
		if textGenOutput == nil || len(textGenOutput.GetText()) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s", textGenOutput, modelName)
		}
		outputJson, err := protojson.Marshal(textGenOutput)
		if err != nil {
			return nil, err
		}
		output := &structpb.Struct{}
		protojson.Unmarshal(outputJson, output)
		outputs = append(outputs, output)

	}
	return outputs, nil
}
