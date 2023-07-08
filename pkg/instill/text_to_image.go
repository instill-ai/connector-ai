package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeTextToImage(model *Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, model.Name)
	}

	outputs := []*connectorPB.DataPayload{}
	for idx := range inputs {
		dataPayload := inputs[idx]
		if len(dataPayload.Texts) <= 0 {
			return nil, fmt.Errorf("invalid input: %v for model: %s", *dataPayload, model.Name)
		}
		steps := int64(dataPayload.GetMetadata().GetFields()["steps"].GetNumberValue())
		cfgScale := float32(dataPayload.GetMetadata().GetFields()["cfg_scale"].GetNumberValue())
		seed := int64(dataPayload.GetMetadata().GetFields()["seed"].GetNumberValue())
		samples := int64(dataPayload.GetMetadata().GetFields()["samples"].GetNumberValue())

		taskInput := &modelPB.TaskInput_TextToImage{
			TextToImage: &modelPB.TextToImageInput{
				Prompt:   inputs[idx].Texts[0],
				Steps:    &steps,
				CfgScale: &cfgScale,
				Seed:     &seed,
				Samples:  &samples,
			},
		}

		// only support batch 1
		req := modelPB.TriggerModelRequest{
			Name:       model.Name,
			TaskInputs: []*modelPB.TaskInput{{Input: taskInput}},
		}
		if c.client == nil || c.client.GRPCClient == nil {
			return nil, fmt.Errorf("client not setup: %v", c.client)
		}
		md := metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", c.getAPIKey()))
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		res, err := c.client.GRPCClient.TriggerModel(ctx, &req)
		if err != nil || res == nil {
			return nil, err
		}
		taskOutputs := res.GetTaskOutputs()
		if len(taskOutputs) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, model.Name)
		}

		textToImgOutput := taskOutputs[0].GetTextToImage()
		if textToImgOutput == nil || len(textToImgOutput.Images) <= 0 {
			return nil, fmt.Errorf("invalid output: %v for model: %s", textToImgOutput, model.Name)
		}

		images := [][]byte{}

		for idx := range textToImgOutput.Images {
			image, err := decodeFromBase64(textToImgOutput.Images[idx])
			if err != nil {
				return nil, fmt.Errorf("invalid output: %v for model: %s", textToImgOutput, model.Name)
			}
			images = append(images, image)
		}

		outputs = append(outputs, &connectorPB.DataPayload{
			DataMappingIndex: inputs[idx].DataMappingIndex,
			Images:           images,
		})
	}
	return outputs, nil
}
