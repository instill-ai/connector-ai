package instill

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/instill-ai/connector-ai/config"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

// initModelPublicServiceClient initialises a ModelPublicServiceClient instance
func initModelPublicServiceClient(serverURL string) (modelPB.ModelPublicServiceClient, *grpc.ClientConn) {
	var clientDialOpts grpc.DialOption
	if serverURL == config.Config.InstillCloud.Host && config.Config.InstillCloud.HTTPS.Cert != "" && config.Config.InstillCloud.HTTPS.Key != "" {
		creds, err := credentials.NewServerTLSFromFile(config.Config.InstillCloud.HTTPS.Cert, config.Config.InstillCloud.HTTPS.Key)
		if err != nil {
			log.Fatal(err.Error())
		}
		clientDialOpts = grpc.WithTransportCredentials(creds)
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("%v:%v", serverURL, config.Config.InstillCloud.PublicPort), clientDialOpts)
	if err != nil {
		log.Fatal(err.Error())
		return nil, nil
	}

	return modelPB.NewModelPublicServiceClient(clientConn), clientConn
}
