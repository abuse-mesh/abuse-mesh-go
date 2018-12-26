//This package contains the command to start the AbuseMesh daemon which will permanently run and
package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/abuse-mesh/abuse-mesh-go-stubs/abusemesh"
	"github.com/abuse-mesh/abuse-mesh-go/pkg/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

const (
	abuseMeshPort = 180
)

var (
	tls       = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile  = flag.String("cert_file", "", "The TLS cert file")
	keyFile   = flag.String("key_file", "", "The TLS key file")
	ipAddress = flag.String("ip_address", "", "The IP address on which the daemon will listen")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ipAddress, abuseMeshPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var ip net.IP
	ip = net.ParseIP(*ipAddress)
	if ip == nil {
		log.Fatalf("Invalid IP address '%s'", *ipAddress)
	}

	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			*certFile = testdata.Path("server1.pem")
		}
		if *keyFile == "" {
			*keyFile = testdata.Path("server1.key")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	abusemesh.RegisterAbuseMeshServer(grpcServer, server.NewAbuseMeshServer(ip))
	grpcServer.Serve(lis)
}
