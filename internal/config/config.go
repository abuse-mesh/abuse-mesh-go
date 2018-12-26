package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type AbuseMeshConfig struct {
	Node Node `mapstructure:"node"`
}

func GetConfig(v *viper.Viper) (*AbuseMeshConfig, error) {

	config := &AbuseMeshConfig{}

	var err error
	if err = v.ReadInConfig(); err != nil {
		log.WithFields(log.Fields{
			"Topic": "Config",
		}).WithError(err).Fatal("Can't read config")
		return nil, err
	}
	if err = v.UnmarshalExact(config); err != nil {
		log.WithFields(log.Fields{
			"Topic": "Config",
		}).WithError(err).Fatal("Can't read config")
		return nil, err
	}

	return config, nil
}
