package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeOCR(model *Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, model.Name)
	}

	outputs := []*connectorPB.DataPayload{}
	for idx := range inputs {
		dataPayload := inputs[idx]
		if len(dataPayload.Images) <= 0 {
			return nil, fmt.Errorf("invalid input: %v for model: %s", *dataPayload, model.Name)
		}
		base64Str, err := encodeToBase64(dataPayload.Images[0])
		if err != nil {
			return nil, fmt.Errorf("invalid image string: %v for model: %s", dataPayload.Images[0], model.Name)
		}
		taskInput := &modelPB.TaskInput_Ocr{
			Ocr: &modelPB.OcrInput{
				Type: &modelPB.OcrInput_ImageBase64{ImageBase64: base64Str},
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

		ocrOutput := taskOutputs[0].GetOcr()
		if ocrOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", ocrOutput, model.Name)
		}
		values := make([]*structpb.Value, 0, len(ocrOutput.Objects))
		for _, o := range ocrOutput.Objects {
			obj := &structpb.Value{
				Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"text":  {Kind: &structpb.Value_StringValue{StringValue: o.Text}},
							"score": {Kind: &structpb.Value_NumberValue{NumberValue: float64(o.Score)}},
							"bounding_box": {Kind: &structpb.Value_StructValue{
								StructValue: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"top":    {Kind: &structpb.Value_NumberValue{NumberValue: float64(o.BoundingBox.Top)}},
										"left":   {Kind: &structpb.Value_NumberValue{NumberValue: float64(o.BoundingBox.Left)}},
										"width":  {Kind: &structpb.Value_NumberValue{NumberValue: float64(o.BoundingBox.Width)}},
										"height": {Kind: &structpb.Value_NumberValue{NumberValue: float64(o.BoundingBox.Height)}},
									},
								},
							},
							},
						},
					},
				},
			}
			values = append(values, obj)
		}
		outputs = append(outputs, &connectorPB.DataPayload{
			DataMappingIndex: inputs[idx].DataMappingIndex,
			StructuredData: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"objects": {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: values}}},
				},
			},
		})
	}
	return outputs, nil
}
