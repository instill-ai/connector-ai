package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeTextGeneration(grpcClient modelPB.ModelPublicServiceClient, model *Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, model.Name)
	}

	outputs := []*connectorPB.DataPayload{}

	for idx := range inputs {
		dataPayload := inputs[idx]
		if len(dataPayload.Texts) <= 0 {
			return nil, fmt.Errorf("invalid input: %v for model: %s", dataPayload, model.Name)
		}
		seed := int64(dataPayload.GetMetadata().GetFields()["seed"].GetNumberValue())
		outputLen := int64(dataPayload.GetMetadata().GetFields()["output_len"].GetNumberValue())
		badWords := dataPayload.GetMetadata().GetFields()["bad_words"].GetStringValue()
		stopWords := dataPayload.GetMetadata().GetFields()["stop_words"].GetStringValue()
		topK := int64(dataPayload.GetMetadata().GetFields()["top_k"].GetNumberValue())

		taskInput := &modelPB.TaskInput_TextGeneration{
			TextGeneration: &modelPB.TextGenerationInput{
				Prompt:        dataPayload.Texts[0],
				OutputLen:     &outputLen,
				BadWordsList:  &badWords,
				StopWordsList: &stopWords,
				Topk:          &topK,
				Seed:          &seed,
			},
		}

		// only support batch 1
		req := modelPB.TriggerModelRequest{
			Name:       model.Name,
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
			return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, model.Name)
		}

		textGenOutput := taskOutputs[0].GetTextGeneration()
		if textGenOutput == nil || len(textGenOutput.GetText()) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s", textGenOutput, model.Name)
		}
		outputs = append(outputs, &connectorPB.DataPayload{
			DataMappingIndex: inputs[idx].DataMappingIndex,
			Texts:            []string{textGenOutput.GetText()},
		})
	}
	return outputs, nil
}
