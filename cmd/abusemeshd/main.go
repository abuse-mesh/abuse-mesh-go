//This package contains the command to start the AbuseMesh daemon which will permanently run and
package main

import (
	"github.com/abuse-mesh/abuse-mesh-go/internal/config"
	"github.com/abuse-mesh/abuse-mesh-go/pkg/server"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

	err = server.ServeNewAbuseMeshServer(config)
	log.WithError(err).Error("AbuseMesh protocol server has stopped")
}
