package adminapiserver

import (
	"bytes"
	"context"
	"net"

	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"github.com/abuse-mesh/abuse-mesh-go/internal/config"
	"github.com/abuse-mesh/abuse-mesh-go/internal/entities"
	"github.com/abuse-mesh/abuse-mesh-go/internal/pgp"
	"github.com/abuse-mesh/abuse-mesh-go/pkg/adminapi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type abuseMeshAdminApi struct {
	config      *config.AbuseMeshConfig
	pgpProvider pgp.PGPProvider
	eventStream entities.EventStream
	tables      *entities.TableSet
}

// Returns the Node data of the current node
func (api *abuseMeshAdminApi) GetNode(context.Context, *adminapi.GetNodeRequest) (*abusemesh.Node, error) {
	nodeConfig := api.config.Node

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

	err := api.pgpProvider.GetEntity().Serialize(&buf)
	if err != nil {
		log.WithError(err).Error("Error while serializing public key")
		return nil, errors.WithStack(err)
	}

	return &abusemesh.Node{
		Uuid: &abusemesh.UUID{
			Uuid: api.config.Node.UUID,
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

// Returns all clients of this node
func (api *abuseMeshAdminApi) GetClients(context.Context, *adminapi.GetClientsRequest) (*adminapi.GetClientsResponse, error) {
	panic("not implemented")
}

// Returns all servers of this node
func (api *abuseMeshAdminApi) GetServers(context.Context, *adminapi.GetServersRequest) (*adminapi.GetServersResponse, error) {
	panic("not implemented")
}

//NewAbuseMeshServer creates a new instance of a AbuseMeshServer
func NewAbuseMeshAdminAPI(
	config *config.AbuseMeshConfig,
	pgpProvider pgp.PGPProvider,
	tableSet *entities.TableSet,
	eventStream entities.EventStream,
) *grpc.Server {

	//Configure the Admin API GRPC server
	var grpcOpts []grpc.ServerOption

	//Create a new GRPC server instance
	grpcServer := grpc.NewServer(grpcOpts...)

	//Create a new Admin API instance
	adminAPI := &abuseMeshAdminApi{
		config:      config,
		pgpProvider: pgpProvider,
		tables:      tableSet,
		eventStream: eventStream,
	}

	//Register the AbuseMeshServer at the GRPC server
	adminapi.RegisterAdmininterfaceServer(grpcServer, adminAPI)

	return grpcServer
}
