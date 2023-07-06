package instill

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeInstanceSegmentation(model *Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
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
		taskInput := &modelPB.TaskInput_InstanceSegmentation{
			InstanceSegmentation: &modelPB.InstanceSegmentationInput{
				Type: &modelPB.InstanceSegmentationInput_ImageBase64{ImageBase64: base64Str},
			},
		}
		tasklInputs = append(tasklInputs, &modelPB.TaskInput{Input: taskInput})
	}
	req := modelPB.TriggerModelRequest{
		Name:       model.Name,
		TaskInputs: tasklInputs,
	}
	if c.client == nil || c.client.GRPCClient == nil {
		return nil, fmt.Errorf("client not setup: %v", c.client)
	}
	res, err := c.client.GRPCClient.TriggerModel(context.Background(), &req)
	if err != nil || res == nil {
		return nil, err
	}
	taskOutputs := res.GetTaskOutputs()
	if len(taskOutputs) <= 0 {
		return nil, fmt.Errorf("invalid output: %v for model: %s", taskOutputs, model.Name)
	}

	outputs := []*connectorPB.DataPayload{}
	for idx := range inputs {
		instanceSegmentationOp := taskOutputs[idx].GetInstanceSegmentation()
		if instanceSegmentationOp == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", instanceSegmentationOp, model.Name)
		}
		values := make([]*structpb.Value, 0, len(instanceSegmentationOp.Objects))
		for _, o := range instanceSegmentationOp.Objects {
			obj := &structpb.Value{
				Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"rle":   {Kind: &structpb.Value_StringValue{StringValue: o.Rle}},
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
