package server

import (
	"bytes"
	"context"
	"net"

	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"github.com/abuse-mesh/abuse-mesh-go/internal/config"
	"github.com/abuse-mesh/abuse-mesh-go/internal/entities"
	"github.com/abuse-mesh/abuse-mesh-go/internal/pgp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type abuseMeshServer struct {
	config      *config.AbuseMeshConfig
	pgpProvider pgp.PGPProvider
	eventStream entities.EventStream
	tables      *entities.TableSet
}

func (server abuseMeshServer) GetNode(context.Context, *abusemesh.GetNodeRequest) (*abusemesh.Node, error) {
	nodeConfig := server.config.Node

	//Find out what ip family we have
	ip := net.ParseIP(nodeConfig.ListenIP)
	var ipFamily abusemesh.IPAddressFamily
	if ip.To4() != nil {
		ipFamily = abusemesh.IPAddressFamily_IPFAMILY_IPV4
	} else {
		ipFamily = abusemesh.IPAddressFamily_IPFAMILY_IPV6
	}

	contactDetailsConfig := nodeConfig.ContactDetails
	contactPersonsConfig := contactDetailsConfig.ContactPersons

	var contactPersons []*abusemesh.ContactDetails_Person
	for _, personConfig := range contactPersonsConfig {
		contactPersons = append(contactPersons, &abusemesh.ContactDetails_Person{
			EmailAddress: personConfig.EmailAddress,
			FirstName:    personConfig.FirstName,
			JobTitle:     personConfig.JobTitle,
			LastName:     personConfig.LastName,
			MiddleName:   personConfig.MiddleName,
			PhoneNumber:  personConfig.PhoneNumber,
		})
	}

	contactDetails := &abusemesh.ContactDetails{
		OrganizationName: contactDetailsConfig.OrganizationName,
		EmailAddress:     contactDetailsConfig.EmailAddress,
		PhysicalAddress:  contactDetailsConfig.PhysicalAddress,
		PhoneNumber:      contactDetailsConfig.PhoneNumber,
		ContactPersons:   contactPersons,
	}

	var buf bytes.Buffer

	err := server.pgpProvider.GetPublicKey().Serialize(&buf)
	if err != nil {
		log.WithError(err).Error("Error while serializing public key")
		return nil, errors.WithStack(err)
	}

	return &abusemesh.Node{
		Uuid: &abusemesh.UUID{
			Uuid: server.config.Node.UUID,
		},
		ASN: nodeConfig.ASN,
		IpAddress: &abusemesh.IPAddress{
			Address:       nodeConfig.ListenIP,
			AddressFamily: ipFamily,
		},
		ProtocolVersion: abusemesh.AbuseMeshProtocolVersion,
		ContactDetails:  contactDetails,
		PgpEntity: &abusemesh.PGPEntity{
			PgpPackets: buf.Bytes(),
		},
	}, nil
}

// With this call a client offers a signature of the identity to the server
// This allows the server to increase its credibility
func (abuseMeshServer) OfferSignature(context.Context, *abusemesh.OfferSignatureRequest) (*abusemesh.OfferSignatureResponse, error) {
	panic("not implemented")
}

func (abuseMeshServer) NegotiateNeighborship(context.Context, *abusemesh.NegotiateNeighborshipRequest) (*abusemesh.NegotiateNeighborshipResponse, error) {
	panic("not implemented")
}

//Observes a event stream and executes a closure on every update
type eventStreamObserver struct {
	//The closure to be called
	eventUpdate func(entities.Event)
}

func (observer *eventStreamObserver) EventUpdate(event entities.Event) {
	observer.eventUpdate(event)
}

// Opens a stream on which all table events of a node are published
func (server abuseMeshServer) TableEventStream(req *abusemesh.TableEventStreamRequest, stream abusemesh.AbuseMesh_TableEventStreamServer) error {
	ctx := stream.Context()

	eventChan := make(chan entities.Event, 1000) //NOTE the buffer size was chosen arbitrarily and may need to be tweaked

	//Create a observer
	observer := eventStreamObserver{
		//On every event update call this function
		eventUpdate: func(event entities.Event) {
			eventChan <- event
		},
	}

	//Attack to the event stream
	server.eventStream.Attach(&observer)
	//Make sure we always detach
	defer server.eventStream.Detach(&observer)

	//Loop forever
	for {
		select {
		//Get a update from the event stream
		case event := <-eventChan:
			//Check if we are dealing with a generic event
			if genericEvent, ok := event.(*entities.GenericEvent); ok {
				err := stream.Send(&genericEvent.TableEvent)
				if err != nil {
					log.WithError(err).Error("Error while sending table event")
				}
			} else {
				log.Error("Unknown event type")
			}

		//Get a stop signal from the client
		case <-ctx.Done():
			log.Info("Table event stream with '' was closed by client")
			return nil
		}
	}
}

//NewAbuseMeshServer creates a new instance of a AbuseMeshServer
func NewAbuseMeshServer(
	config *config.AbuseMeshConfig,
	pgpProvider pgp.PGPProvider,
	tableSet *entities.TableSet,
	eventStream entities.EventStream,
) *grpc.Server {

	//Configure the AbuseMesh protocol GRPC server
	var grpcOpts []grpc.ServerOption

	//Create a new GRPC server instance
	grpcServer := grpc.NewServer(grpcOpts...)

	//Create a new AbuseMesh protocol server instace
	abuseMeshServerInstance := &abuseMeshServer{
		config:      config,
		pgpProvider: pgpProvider,
		tables:      tableSet,
		eventStream: eventStream,
	}

	//Register the AbuseMeshServer at the GRPC server
	abusemesh.RegisterAbuseMeshServer(grpcServer, abuseMeshServerInstance)

	return grpcServer
}
