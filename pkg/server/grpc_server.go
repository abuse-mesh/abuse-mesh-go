package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"github.com/abuse-mesh/abuse-mesh-go/internal/config"
	"github.com/abuse-mesh/abuse-mesh-go/internal/entities"
	"github.com/abuse-mesh/abuse-mesh-go/internal/pgp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
)

type abuseMeshServer struct {
	config      *config.AbuseMeshConfig
	pgpProvider pgp.PGPProvider
	eventStream entities.EventStream
	tables      *entities.TableSet
}

func (server abuseMeshServer) StartEventStreamRoutine(ctx context.Context) error {
	log.Info("Starting EventStream.Run()")
	return server.eventStream.Run(ctx)
}

func (server abuseMeshServer) StartTableRoutine(ctx context.Context) error {
	log.Info("Starting TableSet.Run()")
	return server.tables.Run(ctx)
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

func (server abuseMeshServer) GetNodeTable(ctx context.Context, req *abusemesh.GetNodeTableRequest) (*abusemesh.GetNodeTableResponse, error) {
	respChan := make(chan entities.Node)
	getNodesRequest := entities.GetAllNodesRequest{
		ResponseChan: respChan,
		Context:      ctx,
	}

	//Send the request to the tables routine
	server.tables.Channel <- &getNodesRequest

	//Await the reply
	var nodes []*abusemesh.Node
	for {
		select {
		case node, open := <-respChan:
			//Break from read loop as soon as the response channel has been closed
			if !open {
				break
			}

			pbNode, err := node.ToProtobuf()
			if err != nil {
				return nil, err
			}

			nodes = append(nodes, pbNode)

		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "Request canceled")
		}
	}
}

// With this call a client offers a signature of the identity to the server
// This allows the server to increase its credibility
func (abuseMeshServer) OfferSignature(context.Context, *abusemesh.OfferSignatureRequest) (*abusemesh.OfferSignatureResponse, error) {
	panic("not implemented")
}

func (abuseMeshServer) GetNeighborTable(context.Context, *abusemesh.GetNeighborTableRequest) (*abusemesh.GetNeighborTableResponse, error) {
	panic("not implemented")
}

func (abuseMeshServer) GetReportTable(context.Context, *abusemesh.GetReportTableRequest) (*abusemesh.GetReportTableResponse, error) {
	panic("not implemented")
}

// Returns the contents of the report confirmation table of the node
func (abuseMeshServer) GetReportConfirmationTable(context.Context, *abusemesh.GetReportConfirmationTableRequest) (*abusemesh.GetReportConfirmationTableResponse, error) {
	panic("not implemented")
}

func (abuseMeshServer) GetDelistRequestTable(context.Context, *abusemesh.GetDelistRequestTableRequest) (*abusemesh.GetDelistRequestTableResponse, error) {
	panic("not implemented")
}

// Returns the contents of the report delist acceptance table of the node
func (abuseMeshServer) GetDelistAcceptanceTable(context.Context, *abusemesh.GetDelistAcceptanceTableRequest) (*abusemesh.GetDelistAcceptanceTableResponse, error) {
	panic("not implemented")
}

// Returns all historic table events leading up to this point
func (server abuseMeshServer) GetHistoricTableEvents(context.Context, *abusemesh.GetHistoricTableEventsRequest) (*abusemesh.GetHistoricTableEventsResponse, error) {

	events := server.eventStream.GetAllEvents()
	responseEvents := make([]*abusemesh.TableEvent, 0, len(events))

	for _, event := range events {
		switch e := event.(type) {
		case *entities.GenericEvent:
			responseEvents = append(responseEvents, &e.TableEvent)
		default:
			log.Errorf("Unknown event type '%T'", e)
			return nil, errors.New("Error while retrieving events")
		}
	}

	return &abusemesh.GetHistoricTableEventsResponse{
		Events: responseEvents,
	}, nil
}

// Opens a stream on which all table events of a node are published
func (abuseMeshServer) TableEventStream(*abusemesh.TableEventStreamRequest, abusemesh.AbuseMesh_TableEventStreamServer) error {
	panic("not implemented")
}

//ServeNewAbuseMeshServer creates a new instance of a AbuseMeshServer and starts serving requests
func ServeNewAbuseMeshServer(config *config.AbuseMeshConfig) error {
	var listener net.Listener
	//Start a TCP listener for the AbuseMesh protocol
	tcpListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Node.ListenIP, config.Node.ListenPort))
	if err != nil {
		return errors.Wrap(err, "failed to create TCP listener")
	}

	if !config.Node.Insecure {
		//Config from https://cipherli.st/
		tlsConfig := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}

		//Create a new certificate from the cert and key files
		certificate, err := tls.LoadX509KeyPair(config.Node.TLSCertFile, config.Node.TLSKeyFile)
		if err != nil {
			return errors.Wrap(err, "Error while loading x.509 certificate and key")
		}

		//Add the certificate to the list
		tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)

		//Create a new tls listener with config and overwrite
		listener = tls.NewListener(tcpListener, tlsConfig)

		log.Infof("Starting TLS listener on %s:%d", config.Node.ListenIP, config.Node.ListenPort)
	} else {
		listener = tcpListener

		log.Infof("Starting TCP listener on %s:%d", config.Node.ListenIP, config.Node.ListenPort)
	}

	//Configure the AbuseMesh protocol GRPC server
	var grpcOpts []grpc.ServerOption

	//Create a new GRPC server instance
	grpcServer := grpc.NewServer(grpcOpts...)

	//Create a new AbuseMesh protocol server instace
	abuseMeshServerInstance, err := NewAbuseMeshServer(config)
	if err != nil {
		return err
	}

	//Register the AbuseMeshServer at the GRPC server
	abusemesh.RegisterAbuseMeshServer(grpcServer, abuseMeshServerInstance)

	//Create a error channel on which any goroutine can dump a error to stop the server in a unrecoverable scenario
	errChan := make(chan error)

	go func() {
		errChan <- abuseMeshServerInstance.(*abuseMeshServer).StartTableRoutine(context.Background())
	}()

	go func() {
		errChan <- abuseMeshServerInstance.(*abuseMeshServer).StartEventStreamRoutine(context.Background())
	}()

	go func() {
		//Start service requests using the listener
		errChan <- grpcServer.Serve(listener)
	}()

	//Wait until one of the goroutines stop working
	err = <-errChan

	return err
}

//NewAbuseMeshServer creates a new instance of a AbuseMeshServer
func NewAbuseMeshServer(config *config.AbuseMeshConfig) (abusemesh.AbuseMeshServer, error) {

	var pgpProvider pgp.PGPProvider

	switch config.Node.PGPProvider {
	case "file":
		var passphrase []byte
		if config.Node.PGPProviders.File.AskPassphrase {
			fmt.Printf("The PGP provider requires a passphrase for the private key\nPassphrase: ")

			var err error
			passphrase, err = terminal.ReadPassword(0)
			//Print a new line to make the output nice
			fmt.Print("\n")

			if err != nil {
				return nil, errors.Wrap(err, "Error while reading passphrase")
			}
		} else {
			passphrase = []byte(config.Node.PGPProviders.File.Passphrase)
		}

		var err error
		pgpProvider, err = pgp.NewFilePGPProvider(
			config.Node.PGPProviders.File.KeyRingFile,
			config.Node.PGPProviders.File.KeyID,
			passphrase,
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("Unknown PGP provider '%s'", config.Node.PGPProvider)
	}

	tableSet := &entities.TableSet{
		Channel: make(chan entities.TableRequest, 100), //TODO make buffer configurable
	}

	//TODO make event stream type configurable
	eventStream := entities.NewInMemoryEventStream(tableSet, 1000) //TODO make event stream buffer configurable

	//Attach the tableSet as observer to the event stream
	eventStream.Attach(tableSet)

	return &abuseMeshServer{
		config:      config,
		pgpProvider: pgpProvider,
		tables:      tableSet,
		eventStream: eventStream,
	}, nil
}
