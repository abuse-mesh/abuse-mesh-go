package client

import (
	"context"
	"log"
	"time"

	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"google.golang.org/grpc"
)

type AbuseMeshClient struct {
	grpcClient     abusemesh.AbuseMeshClient
	grpcConnection *grpc.ClientConn
	requestTimeout time.Duration
}

func NewAbuseMeshClient() *AbuseMeshClient {
	var opts []grpc.DialOption

	//TODO add options
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial("localhost:1180", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	//defer conn.Close()

	client := abusemesh.NewAbuseMeshClient(conn)

	return &AbuseMeshClient{
		requestTimeout: 10 * time.Second,
		grpcClient:     client,
		grpcConnection: conn,
	}
}

func (client *AbuseMeshClient) Close() error {
	return client.grpcConnection.Close()
}

func (client *AbuseMeshClient) GetNode(request *abusemesh.GetNodeRequest) (*abusemesh.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.requestTimeout)
	defer cancel()
	return client.grpcClient.GetNode(ctx, request)
}
