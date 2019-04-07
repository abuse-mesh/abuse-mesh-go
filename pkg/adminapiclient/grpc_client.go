package adminapiclient

import (
	"context"
	"log"
	"time"

	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"github.com/abuse-mesh/abuse-mesh-go/pkg/adminapi"
	"google.golang.org/grpc"
)

func NewAbuseMeshAdminClient() *AdminClient {
	var opts []grpc.DialOption

	//TODO add options
	opts = append(opts, grpc.WithInsecure())

	//TODO make url a parameter
	conn, err := grpc.Dial("localhost:1181", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	client := adminapi.NewAdmininterfaceClient(conn)

	return &AdminClient{
		unaryRequestTimeout: 10 * time.Second,
		grpcClient:          client,
		grpcConnection:      conn,
	}
}

type AdminClient struct {
	grpcClient     adminapi.AdmininterfaceClient
	grpcConnection *grpc.ClientConn

	//The request timeout for unary requests
	unaryRequestTimeout time.Duration
}

//GetNode requests the server to send back information about itself
func (client *AdminClient) GetNode(request *adminapi.GetNodeRequest) (*abusemesh.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.unaryRequestTimeout)
	defer cancel()
	return client.grpcClient.GetNode(ctx, request)
}
