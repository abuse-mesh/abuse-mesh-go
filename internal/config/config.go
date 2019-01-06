package config

import (
	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"
)

type AbuseMeshConfig struct {
	Node           NodeConfig           `mapstructure:"node" json:"node"`
	AdminInterface AdminInterfaceConfig `mapstructure:"admin-interface" json:"admin-interface"`
}

func GetConfig(v *viper.Viper) (*AbuseMeshConfig, error) {

	config := &AbuseMeshConfig{}

	var err error
	if err = v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err = v.UnmarshalExact(config); err != nil {
		return nil, err
	}

	//Validate the AbuseMeshConfig, use the 'validate' tag to specify the validation rules
	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

type AdminInterfaceConfig struct {
	ListenIP   string `mapstructure:"listen-ip" json:"listen-ip" yaml:"listen-ip" validate:"required,ip"`
	ListenPort int    `mapstructure:"listen-port" json:"listen-port,omitempty" yaml:"listen-port,omitempty" validate:"min=1,max=65535"`
}
