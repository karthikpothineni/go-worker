package config

import (
	"log"
	"os"
	"strings"

	"github.com/fatih/structs"
	"github.com/spf13/viper"
)

const (
	ConfigurationType     = "toml"
	DevConfigurationPath  = "config/"
	ProdConfigurationPath = "/etc/go-worker/config/"
	ConfigName            = "config"
	DefaultTagName        = "json"
	DevEnvironment        = "dev"
	ProdEnvironment       = "prod"
)

var config *viper.Viper

// Init - takes the environment and starts the viper for preparing the config
func Init(env string) {
	var err error
	v := viper.New()
	v.SetConfigType(ConfigurationType)

	if strings.ToLower(env) == DevEnvironment {
		v.AddConfigPath(DevConfigurationPath)
	} else {
		v.AddConfigPath(ProdConfigurationPath)
	}
	_ = os.Setenv("env", env)

	v.SetConfigName(strings.ToLower(ConfigName))
	err = v.MergeInConfig()
	if err != nil {
		log.Fatal("Error on parsing configuration file. Error " + err.Error())
	}
	// set structs default tag name
	structs.DefaultTagName = DefaultTagName

	config = v
}

// GetConfig - to expose the config object
func GetConfig() *viper.Viper {
	return config
}
