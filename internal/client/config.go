package client

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Configuration struct {
	Password        string `mapstructure:"Password"`
	Destination     string `mapstructure:"Destination"`
	OpenInBrowser   int    `mapstructure:"OpenInBrowser"`
	CopyToClipboard int    `mapstructure:"CopyToClipboard"`
}

func NewConfiguration() *Configuration {
	return setDefaults()
}

func setDefaults() *Configuration {
	var configuration Configuration
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.SetDefault("Password", "")
	viper.SetDefault("Destination", "127.0.0.1:9001")
	viper.SetDefault("OpenInBrowser", 1)
	viper.SetDefault("CopyToClipboard", 1)
	viper.Unmarshal(&configuration)
	return &configuration
}

func (c *Configuration) loadConfig() error {
	log.Info("loading config from config.toml")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info("failed to find config.toml, using defaults")
			return err
		} else {
			log.Info("failed to access config.toml (Are permissions correct?)")
			return err
		}
	}
	if err := viper.UnmarshalExact(c); err != nil {
		log.Info("failed to parse config.toml, using defaults")
		log.Info(err.Error())
		return err
	}

	log.Info("using")
	log.Info("password: " + c.Password)
	log.Info("destination: " + c.Destination)
	log.Infof("open: %d", c.OpenInBrowser)
	log.Infof("copy: %d", c.CopyToClipboard)

	return nil
}
