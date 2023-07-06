package instill

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"google.golang.org/protobuf/types/known/structpb"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

func (c *Connection) executeImageClassification(model *Model, inputs []*connectorPB.DataPayload) ([]*connectorPB.DataPayload, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", inputs, model.Name)
	}
	dataPayload := inputs[0]
	if len(dataPayload.Images) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s", *dataPayload, model.Name)
	}
	base64Str, err := fetchImageFromURL(dataPayload.Images[0])
	fmt.Println("base64Str", base64Str)
	if err != nil {
		return nil, fmt.Errorf("invalid image string: %v for model: %s", dataPayload.Images[0], model.Name)
	}
	modelInput := &modelPB.TaskInput_Classification{
		Classification: &modelPB.ClassificationInput{
			Type: &modelPB.ClassificationInput_ImageBase64{ImageBase64: base64Str},
		},
	}
	req := modelPB.TriggerModelRequest{
		Name:       model.Name,
		TaskInputs: []*modelPB.TaskInput{{Input: modelInput}},
	}
	if c.client == nil || c.client.GRPCClient == nil {
		return nil, fmt.Errorf("client not setup: %v", c.client)
	}
	fmt.Println("before TriggerModel")
	res, err := c.client.GRPCClient.TriggerModel(context.Background(), &req)
	fmt.Printf("\n\n after TriggerModel res:%v, err:%v \n\n", res, err)
	if err != nil || res == nil {
		return nil, err
	}
	output := res.GetTaskOutputs()
	fmt.Printf("\n\n after TriggerModel output:%v \n\n", output)
	if len(output) <= 0 {
		return nil, fmt.Errorf("invalid output: %v for model: %s", output, model.Name)
	}
	imgClassificationOp := output[0].GetClassification()
	fmt.Printf("\n\n after TriggerModel imgClassificationOp:%v \n\n", imgClassificationOp)
	if imgClassificationOp == nil {
		return nil, fmt.Errorf("invalid output: %v for model: %s", imgClassificationOp, model.Name)
	}
	inputs[0].StructuredData = &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"category": {Kind: &structpb.Value_StringValue{StringValue: imgClassificationOp.Category}},
			"score":    {Kind: &structpb.Value_NumberValue{NumberValue: float64(imgClassificationOp.Score)}},
		},
	}
	fmt.Printf("\n\n after TriggerModel inputs[0].StructuredData:%v \n\n", inputs[0].StructuredData)
	return inputs, nil
}

func fetchImageFromURL(input []byte) (string, error) {
	// Check if the input is a URL
	s := string(input)
	_, err := url.ParseRequestURI(s)
	if err != nil {
		// Input is not a URL, return it as it is
		return s, nil
	}
	response, err := http.Get(s)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	imageData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	// Encode the image data to base64
	return base64.StdEncoding.EncodeToString(imageData), nil
}
