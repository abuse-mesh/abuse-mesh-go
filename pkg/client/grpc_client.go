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

	//The request timeout for unary requests
	unaryRequestTimeout time.Duration
}

func NewAbuseMeshClient() *AbuseMeshClient {
	var opts []grpc.DialOption

	//TODO add options
	opts = append(opts, grpc.WithInsecure())

	//TODO make url a parameter
	conn, err := grpc.Dial("localhost:1180", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	//defer conn.Close()

	client := abusemesh.NewAbuseMeshClient(conn)

	return &AbuseMeshClient{
		unaryRequestTimeout: 10 * time.Second,
		grpcClient:          client,
		grpcConnection:      conn,
	}
}

func (client *AbuseMeshClient) Close() error {
	return client.grpcConnection.Close()
}

func (client *AbuseMeshClient) GetNode(request *abusemesh.GetNodeRequest) (*abusemesh.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), client.unaryRequestTimeout)
	defer cancel()
	return client.grpcClient.GetNode(ctx, request)
}

func (client *AbuseMeshClient) NegotiateNeighborship(
	request *abusemesh.NegotiateNeighborshipRequest,
) (
	*abusemesh.NegotiateNeighborshipResponse, error,
) {
	ctx, cancel := context.WithTimeout(context.Background(), client.unaryRequestTimeout)
	defer cancel()
	return client.grpcClient.NegotiateNeighborship(ctx, request)
}

func (client *AbuseMeshClient) TableEventStream(
	request *abusemesh.TableEventStreamRequest,
) (
	abusemesh.AbuseMesh_TableEventStreamClient,
	context.CancelFunc,
	error,
) {
	ctx, cancel := context.WithCancel(context.Background())

	streamClient, err := client.grpcClient.TableEventStream(ctx, request)
	return streamClient, cancel, err
}
