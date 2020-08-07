package server

import (
	"crypto/sha1"
	"time"

	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"
)

type Configuration struct {
	Port     string        `mapstructure:"Port"`
	Password string        `mapstructure:"Password"`
	Ttl      time.Duration `mapstructure:"Ttl"`
}

func NewConfiguration() *Configuration {
	return setDefaults()
}

func setDefaults() *Configuration {
	var configuration Configuration
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.SetDefault("Port", ":9001")
	viper.SetDefault("Password", "")
	viper.SetDefault("Ttl", time.Duration(0))
	viper.Unmarshal(&configuration)
	return &configuration
}

func (c *Configuration) loadConfig() error {
	log.Info("Loading config from config.toml")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info("Failed to find config.toml, using defaults")
			return err
		} else {
			log.Info("Failed to access config.toml (Are permissions correct?)")
			return err
		}
	}
	if err := viper.UnmarshalExact(c); err != nil {
		log.Info("Failed to parse config.toml, using defaults")
		log.Info(err.Error())
		return err
	}

	log.Info("Using")
	log.Info("Port: " + c.Port)
	log.Info("Password: " + c.Password)
	log.Infof("Ttl: %+v", c.Ttl)

	h := sha1.New()
	if _, err := h.Write([]byte(c.Password)); err != nil {
		log.Info("Failed to hash password from config.ini", err.Error())
		return err
	}
	c.Password = string(h.Sum(nil))

	return nil
}
