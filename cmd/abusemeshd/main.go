//This package contains the command to start the AbuseMesh daemon which will permanently run and
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/abuse-mesh/abuse-mesh-go/internal/config"
	"github.com/abuse-mesh/abuse-mesh-go/internal/entities"
	"github.com/abuse-mesh/abuse-mesh-go/internal/pgp"
	"github.com/abuse-mesh/abuse-mesh-go/pkg/adminapiserver"
	"github.com/abuse-mesh/abuse-mesh-go/pkg/server"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	validator "gopkg.in/go-playground/validator.v9"
)

func main() {

	var opts struct {
		ConfigFile string `short:"f" long:"config-file" description:"specify a config file"`
		ConfigType string `short:"t" long:"config-type" description:"specify a config type (json, yaml, toml)" default:"yaml"`
	}

	//Parse the cli flags
	_, err := flags.Parse(&opts)
	if err != nil {
		//Exit the program, flags.Parse will print the error
		return
	}

	//Create a new viper config instance
	viper := viper.New()

	//Point viper to the config file
	viper.SetConfigFile(opts.ConfigFile)
	viper.SetConfigType(opts.ConfigType)

	//Make it so all ENV variables need to be prefixed with AM(AbuseMesh)
	viper.SetEnvPrefix("AM")

	//Overwrite config file with ENV variables
	viper.AutomaticEnv()

	//Get the abuse mesh config
	config, err := config.GetConfig(viper)
	if err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			for _, valErr := range validationErrs {
				log.WithFields(log.Fields{
					"field":      valErr.Field(),
					"tag":        valErr.Tag(),
					"actual_tag": valErr.ActualTag(),
					"kind":       valErr.Kind(),
					"type":       valErr.Type(),
					"param":      valErr.Param(),
					"value":      valErr.Value(),
				}).Error("Error while loading config, config validation")
			}
		} else {
			log.WithError(err).Error("Error while loading config")
		}

		log.Fatal("Stopped due to invalid config")
	}

	log.Info("Config has been loaded")

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
				log.WithError(err).Fatal("Error while reading passphrase")
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
			log.WithError(err).Fatal("Error while creating PGPProvider")
		}
	default:
		log.Fatalf("Unknown PGP provider '%s'", config.Node.PGPProvider)
	}

	tableSet := &entities.TableSet{
		Channel: make(chan entities.TableRequest, 100), //TODO make buffer configurable
	}

	//TODO make event stream type configurable
	eventStream := entities.NewInMemoryEventStream(tableSet, 1000) //TODO make event stream buffer configurable

	//Attach the tableSet as observer to the event stream
	eventStream.Attach(tableSet)

	errChan := make(chan error)

	go func() {
		log.Info("Starting TableSet.Run()")
		errChan <- tableSet.Run(context.Background())
	}()

	go func() {
		log.Info("Starting EventStream.Run()")
		errChan <- eventStream.Run(context.Background())
	}()

	go func() {
		abuseMeshServer := server.NewAbuseMeshServer(config, pgpProvider, tableSet, eventStream)

		abuseMeshAddr := fmt.Sprintf("%s:%d", config.Node.ListenIP, config.Node.ListenPort)
		abuseMeshListener, err := net.Listen("tcp", abuseMeshAddr)
		if err != nil {
			errChan <- errors.Errorf("failed to create TCP listener '%s': '%s'", err, abuseMeshAddr)
		}

		abuseMeshListener, err = wrapListener(abuseMeshListener, config)
		if err != nil {
			errChan <- errors.Errorf("failed to wrap TCP listener: '%s'", err)
		}

		log.Infof("Staring to serve AbuseMesh protocol on '%s'", abuseMeshAddr)

		errChan <- abuseMeshServer.Serve(abuseMeshListener)
	}()

	go func() {
		abuseMeshAdminAPI := adminapiserver.NewAbuseMeshAdminAPI(config, pgpProvider, tableSet, eventStream)

		adminAPIAddr := fmt.Sprintf("%s:%d", config.AdminInterface.ListenIP, config.AdminInterface.ListenPort)
		adminAPIListener, err := net.Listen("tcp", adminAPIAddr)
		if err != nil {
			errChan <- errors.Errorf("failed to create TCP listener '%s': '%s'", err, adminAPIAddr)
		}

		adminAPIListener, err = wrapListener(adminAPIListener, config)
		if err != nil {
			errChan <- errors.Errorf("failed to wrap TCP listener: '%s'", err)
		}

		log.Infof("Staring to serve admin API on '%s'", adminAPIAddr)

		errChan <- abuseMeshAdminAPI.Serve(adminAPIListener)
	}()

	//Wait for one of the go routines to error
	err = <-errChan

	log.WithError(err).Error("AbuseMesh protocol server has stopped")
}

func wrapListener(innerListener net.Listener, config *config.AbuseMeshConfig) (net.Listener, error) {
	var listener net.Listener

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
			return nil, errors.Wrap(err, "Error while loading x.509 certificate and key")
		}

		//Add the certificate to the list
		tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)

		//Create a new tls listener with config and overwrite
		listener = tls.NewListener(innerListener, tlsConfig)

		log.Infof("Starting TLS listener on %s:%d", config.Node.ListenIP, config.Node.ListenPort)
	} else {
		listener = innerListener

		log.Infof("Starting TCP listener on %s:%d", config.Node.ListenIP, config.Node.ListenPort)
	}

	return listener, nil
}
