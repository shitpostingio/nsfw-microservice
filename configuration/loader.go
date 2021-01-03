package configuration

import (
	"github.com/shitpostingio/nsfw-microservice/configuration/structs"
	"log"

	"github.com/spf13/viper"
)

// Load reads a configuration file and returns its config instance
func Load(path string) (cfg structs.Config, err error) {

	viper.SetConfigFile(path)
	err = viper.ReadInConfig()
	if err != nil {
		return cfg, err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return cfg, err
	}

	err = CheckMandatoryFields(cfg)
	if err != nil {
		log.Fatal(err)
	}

	return

}
