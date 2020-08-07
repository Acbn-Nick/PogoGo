package server

import (
	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"
)

type Configuration struct {
	Port     int
	Password string
	Ttl      int
}

func NewConfiguration() *Configuration {
	return setDefaults()
}

func setDefaults() *Configuration {
	var configuration Configuration
	viper.SetConfigName("config")
	viper.SetConfigType("ini")
	viper.AddConfigPath(".")
	viper.SetDefault("Port", 9001)
	viper.SetDefault("Password", "")
	viper.SetDefault("Ttl", 0)
	viper.Unmarshal(&configuration)
	return &configuration
}

func (c *Configuration) loadConfig() error {
	log.Info("Loading config from config.ini")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info("Failed to find config.ini, using defaults")
			return err
		} else {
			log.Info("Failed to access config.ini (Are permissions correct?)")
			return err
		}
	}
	if err := viper.Unmarshal(c); err != nil {
		log.Info("Failed to parse config.ini, using defaults")
		return err
	}
	log.Info("Using: ")
	log.Info("Port: " + string(c.Port))
	log.Info("Password: " + c.Password)
	log.Info("Ttl: " + string(c.Ttl))

	return nil
}
