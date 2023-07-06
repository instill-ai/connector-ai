package instill

import (
	"context"
	"fmt"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeTextToImage(model *Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, model.Name)
	}
	dataPayload := inputs[0]
	if len(dataPayload.Texts) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", *dataPayload, model.Name)
	}
	steps := int64(dataPayload.GetMetadata().GetFields()["steps"].GetNumberValue())
	cfgScale := float32(dataPayload.GetMetadata().GetFields()["cfg_scale"].GetNumberValue())
	seed := int64(dataPayload.GetMetadata().GetFields()["seed"].GetNumberValue())
	samples := int64(dataPayload.GetMetadata().GetFields()["samples"].GetNumberValue())

	modelInput := &modelPB.TaskInput_TextToImage{
		TextToImage: &modelPB.TextToImageInput{
			Prompt:   inputs[0].Texts[0],
			Steps:    &steps,
			CfgScale: &cfgScale,
			Seed:     &seed,
			Samples:  &samples,
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
	textToImgOutput := output[0].GetTextToImage()
	if textToImgOutput == nil || len(textToImgOutput.Images) <= 0 {
		return nil, fmt.Errorf("invalid output: %v for model: %s", textToImgOutput, model.Name)
	}
	for _, img := range textToImgOutput.Images {
		inputs[0].Images = append(inputs[0].Images, []byte(img))
	}
	return inputs, nil
}
