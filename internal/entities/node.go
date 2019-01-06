package entities

import (
	"bytes"
	"context"
	"net"

	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"github.com/abuse-mesh/abuse-mesh-go/internal/pgp"
	"github.com/abuse-mesh/abuse-mesh-go/internal/utils/conv"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/openpgp"
)

//A Node is a node as defined by the AbuseMesh protocol[inset link to docs] with extra information internal to the node
type Node struct {
	UUID            uuid.UUID
	ProtocolVersion string
	IPAddress       net.IP
	ContactDetails  abusemesh.ContactDetails
	ASN             int32
	PGPEntity       *openpgp.Entity
}

//ToProtobuf converts the node struct into a protobuf stub
func (node Node) ToProtobuf() (*abusemesh.Node, error) {
	//Find out what ip family we have
	var ipFamily abusemesh.IPAddressFamily
	if node.IPAddress.To4() != nil {
		ipFamily = abusemesh.IPAddressFamily_IPFAMILY_IPV4
	} else {
		ipFamily = abusemesh.IPAddressFamily_IPFAMILY_IPV6
	}

	var buf bytes.Buffer
	err := node.PGPEntity.Serialize(&buf)
	if err != nil {
		logrus.WithError(err).Error("Error while serializing pgp entity")
		return nil, errors.WithStack(err)
	}

	return &abusemesh.Node{
		Uuid: &abusemesh.UUID{
			Uuid: node.UUID.String(),
		},
		ASN: node.ASN,
		IpAddress: &abusemesh.IPAddress{
			Address:       node.IPAddress.String(),
			AddressFamily: ipFamily,
		},
		ProtocolVersion: abusemesh.AbuseMeshProtocolVersion,
		ContactDetails:  &node.ContactDetails,
		PgpEntity: &abusemesh.PGPEntity{
			PgpPackets: buf.Bytes(),
		},
	}, nil
}

//NodeFromProtobuf creates a node object from the node stub of the protobuf definition
func NodeFromProtobuf(protobufNode *abusemesh.Node) (Node, error) {
	uuid, err := conv.AuuidToGuuid(protobufNode.Uuid)
	if err != nil {
		return Node{}, err
	}

	//Read the packets into a entity which can be used to check signatures
	var pgpEntity *openpgp.Entity
	pgpEntity, err = pgp.PGPEntityFromBytes(protobufNode.PgpEntity.PgpPackets)
	if err != nil {
		return Node{}, errors.WithStack(err)
	}

	return Node{
		UUID:            uuid,
		ProtocolVersion: protobufNode.ProtocolVersion,
		IPAddress:       net.ParseIP(protobufNode.IpAddress.Address),
		ContactDetails:  *protobufNode.ContactDetails,
		ASN:             protobufNode.ASN,
		PGPEntity:       pgpEntity,
	}, nil
}

//A NodeTable holds the current derived state of all nodes in the network known to the current node
type NodeTable struct {
	Entities map[uuid.UUID]Node
}

func (table *NodeTable) handleTableEvent(eventType abusemesh.TableEventType, entity *abusemesh.TableEvent_Node) error {
	switch eventType {
	case abusemesh.TableEventType_TABLE_UPDATE_NEW:
		node, err := NodeFromProtobuf(entity.Node)
		if err != nil {
			return err
		}

		table.Entities[node.UUID] = node

	case abusemesh.TableEventType_TABLE_UPDATE_EDIT:
		node, err := NodeFromProtobuf(entity.Node)
		if err != nil {
			return err
		}

		table.Entities[node.UUID] = node

	case abusemesh.TableEventType_TABLE_UPDATE_DELETE:
		uuid, err := conv.AuuidToGuuid(entity.Node.Uuid)
		if err != nil {
			return err
		}
		delete(table.Entities, uuid)

	default:
		return errors.Errorf("Unknown abusemesh.TableEventType type '%T'", eventType)
	}

	return nil
}

//GetNodeRequest can be used to request a specific node from a table
type GetNodeRequest struct {
	ResponseChan chan<- *Node
	NodeID       uuid.UUID
}

//Process processes the request and sends a pointer to the node on the ResponseChan or nil if the node was not found
func (req *GetNodeRequest) Process(tables *TableSet) error {
	node, found := tables.nodeTable.Entities[req.NodeID]
	if found {
		req.ResponseChan <- &node
	} else {
		req.ResponseChan <- nil
	}

	return nil
}

//GetAllNodesRequest can be used to request all nodes from the nodes table
type GetAllNodesRequest struct {
	//ResponseChan is the channel over which multiple nodes will be sent
	ResponseChan chan<- Node

	//Context can be used to cancel the sending of nodes
	Context context.Context
}

//Process processes the request and sends a slice of nodes on the ResponseChan or nil if the node was not found
func (req *GetAllNodesRequest) Process(tables *TableSet) error {
	defer close(req.ResponseChan)

	for _, node := range tables.nodeTable.Entities {
		//Of the requester no longer wishes to receive nodes we return
		if req.Context.Err() != nil {

			return nil
		}

		req.ResponseChan <- node
	}

	return nil
}
