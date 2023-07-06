package instill

import (
	"crypto/tls"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

// initModelPublicServiceClient initialises a ModelPublicServiceClient instance
func initModelPublicServiceClient(serverURL string) (modelPB.ModelPublicServiceClient, *grpc.ClientConn) {
	var clientDialOpts grpc.DialOption
	if serverURL == instillCloudHost {
		clientDialOpts = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(serverURL, clientDialOpts)
	if err != nil {
		log.Fatal(err.Error())
		return nil, nil
	}

	return modelPB.NewModelPublicServiceClient(clientConn), clientConn
}
