//go:build integration
// +build integration

package pkg

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/connector-ai/pkg/openai"
	"github.com/instill-ai/connector-ai/pkg/stabilityai"

	connectorv1alpha "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

var (
	stabilityAIKey = "<valid api key>"
	openAIKey      = "<valid api key>"
)

func init() {
	b, _ := ioutil.ReadFile("test_artifacts/open_ai.txt")
	openAIKey = string(b)
	b, _ = ioutil.ReadFile("test_artifacts/stability_ai.txt")
	stabilityAIKey = string(b)
}

func TestStabilityAITextToImage(t *testing.T) {
	config := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"api_key": {Kind: &structpb.Value_StringValue{StringValue: stabilityAIKey}},
			"task":    {Kind: &structpb.Value_StringValue{StringValue: "Text to Image"}},
			"engine":  {Kind: &structpb.Value_StringValue{StringValue: "stable-diffusion-v1-5"}},
		},
	}
	in := []*connectorv1alpha.DataPayload{{
		Texts: []string{"dog", "black"},
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{},
		},
	}}
	c := Init(nil, ConnectorOptions{})
	con, err := c.CreateConnection(c.ListConnectorDefinitionUids()[0], config, nil)
	fmt.Printf("err:%s", err)
	op, err := con.Execute(in)
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func Test_ListEngines(t *testing.T) {
	client := stabilityai.NewClient(stabilityAIKey)
	engines, err := client.ListEngines()
	fmt.Printf("engines: %v, err: %v", engines, err)
}

func TestStabilityAIImageToImage(t *testing.T) {
	config := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"api_key": {Kind: &structpb.Value_StringValue{StringValue: stabilityAIKey}},
			"task":    {Kind: &structpb.Value_StringValue{StringValue: "Image to Image"}},
			"engine":  {Kind: &structpb.Value_StringValue{StringValue: "stable-diffusion-v1"}},
		},
	}
	b, _ := ioutil.ReadFile("test_artifacts/image.jpg")
	in := []*connectorv1alpha.DataPayload{{
		Texts:  []string{"invert colors"},
		Images: [][]byte{b},
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{},
		},
	}}
	c := Init(nil, ConnectorOptions{})
	con, err := c.CreateConnection(c.ListConnectorDefinitionUids()[0], config, nil)
	fmt.Printf("\n err: %s", err)
	op, err := con.Execute(in)
	fmt.Printf("\n op: %v, err: %s", op, err)
	err = ioutil.WriteFile("test_artifacts/image_op.png", op[0].Images[0], 0644)
}

func TestOpenAITextGeneration(t *testing.T) {
	config := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"api_key": {Kind: &structpb.Value_StringValue{StringValue: openAIKey}},
			"task":    {Kind: &structpb.Value_StringValue{StringValue: "Text Generation"}},
			"model":   {Kind: &structpb.Value_StringValue{StringValue: "gpt-3.5-turbo"}},
		},
	}
	in := []*connectorv1alpha.DataPayload{{
		Texts: []string{"how are you doing?"},
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{},
		},
	}}
	c := Init(nil, ConnectorOptions{})
	con, err := c.CreateConnection(c.ListConnectorDefinitionUids()[2], config, nil)
	fmt.Printf("err:%s", err)
	op, err := con.Execute(in)
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func Test_ListModels(t *testing.T) {
	c := openai.Client{
		APIKey:     openAIKey,
		HTTPClient: &http.Client{},
	}
	res, err := c.ListModels()
	fmt.Printf("res: %v, err: %v", res, err)
}

func TestOpenAIAudioTranscription(t *testing.T) {
	config := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"api_key": {Kind: &structpb.Value_StringValue{StringValue: openAIKey}},
			"task":    {Kind: &structpb.Value_StringValue{StringValue: "Speech Recognition"}},
			"model":   {Kind: &structpb.Value_StringValue{StringValue: "whisper-1"}},
		},
	}
	b, _ := ioutil.ReadFile("test_artifacts/recording.m4a")
	in := []*connectorv1alpha.DataPayload{{
		Audios: [][]byte{b},
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{},
		},
	}}
	c := Init(nil, ConnectorOptions{})
	con, err := c.CreateConnection(c.ListConnectorDefinitionUids()[2], config, nil)
	fmt.Printf("err:%s", err)
	op, err := con.Execute(in)
	fmt.Printf("\n op :%v, err:%s", op, err)
}
