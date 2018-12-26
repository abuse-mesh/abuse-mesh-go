package server

import (
	"context"
	"net"

	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
)

type abuseMeshServer struct {
	IPAddress net.IP
}

func (abuseMeshServer) GetNode(context.Context, *abusemesh.GetNodeRequest) (*abusemesh.Node, error) {
	panic("not implemented")
}

func (abuseMeshServer) GetNodeTable(context.Context, *abusemesh.GetNodeTableRequest) (*abusemesh.GetNodeTableResponse, error) {
	panic("not implemented")
}

func (abuseMeshServer) GetNeighborTable(context.Context, *abusemesh.GetNeighborTableRequest) (*abusemesh.GetNeighborTableResponse, error) {
	panic("not implemented")
}

func (abuseMeshServer) GetReportTable(context.Context, *abusemesh.GetReportTableRequest) (*abusemesh.GetReportTableResponse, error) {
	panic("not implemented")
}

func (abuseMeshServer) GetDelistRequestTable(context.Context, *abusemesh.GetDelistRequestTableRequest) (*abusemesh.GetDelistRequestTableResponse, error) {
	panic("not implemented")
}

func (abuseMeshServer) GetDelistAcceptanceTable(context.Context, *abusemesh.GetDelistAcceptanceTableRequest) (*abusemesh.GetDelistAcceptanceTableResponse, error) {
	panic("not implemented")
}

func (abuseMeshServer) TableUpdateStream(*abusemesh.TableUpdateStreamRequest, abusemesh.AbuseMesh_TableUpdateStreamServer) error {
	panic("not implemented")
}

//NewAbuseMeshServer creates a new instance of a AbuseMeshServer
func NewAbuseMeshServer(listeningIP net.IP) abusemesh.AbuseMeshServer {
	return &abuseMeshServer{
		IPAddress: listeningIP,
	}
}
