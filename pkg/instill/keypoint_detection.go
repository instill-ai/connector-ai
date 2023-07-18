package instill

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeKeyPointDetection(grpcClient modelPB.ModelPublicServiceClient, model *Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, model.Name)
	}
	tasklInputs := []*modelPB.TaskInput{}
	for idx := range inputs {
		dataPayload := inputs[idx]
		if len(dataPayload.Images) <= 0 {
			return nil, fmt.Errorf("invalid input: %v for model: %s", dataPayload, model.Name)
		}
		base64Str, err := encodeToBase64(dataPayload.Images[0])
		if err != nil {
			return nil, fmt.Errorf("invalid image string: %v for model: %s", dataPayload.Images[0], model.Name)
		}
		taskInput := &modelPB.TaskInput_Keypoint{
			Keypoint: &modelPB.KeypointInput{Type: &modelPB.KeypointInput_ImageBase64{ImageBase64: base64Str}},
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
		keyPointOutput := taskOutputs[idx].GetKeypoint()
		if keyPointOutput == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s", keyPointOutput, model.Name)
		}
		values := make([]*structpb.Value, 0, len(keyPointOutput.Objects))
		for _, o := range keyPointOutput.Objects {
			keyPoints := make([]*structpb.Value, 0, len(o.Keypoints))
			for _, k := range o.Keypoints {
				kp := &structpb.Value{
					Kind: &structpb.Value_StructValue{
						StructValue: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"v": {Kind: &structpb.Value_NumberValue{NumberValue: float64(k.V)}},
								"x": {Kind: &structpb.Value_NumberValue{NumberValue: float64(k.X)}},
								"y": {Kind: &structpb.Value_NumberValue{NumberValue: float64(k.Y)}},
							},
						},
					},
				}
				keyPoints = append(keyPoints, kp)
			}
			obj := &structpb.Value{
				Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"score":     {Kind: &structpb.Value_NumberValue{NumberValue: float64(o.Score)}},
							"keypoints": {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: keyPoints}}},
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
