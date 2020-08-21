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
	HttpPort string        `mapstructure:"HttpPort"`
}

func NewConfiguration() *Configuration {
	return setDefaults()
}

func setDefaults() *Configuration {
	var configuration Configuration
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.SetDefault("Port", "127.0.0.1:9001")
	viper.SetDefault("Password", "")
	viper.SetDefault("Ttl", time.Duration(0))
	viper.SetDefault("HttpPort", "127.0.0.1:8080")
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
	log.Info("port: " + c.Port)
	log.Info("password: " + c.Password)
	log.Infof("ttl: %+v", c.Ttl)
	log.Info("http port: " + c.HttpPort)

	h := sha1.New()
	if _, err := h.Write([]byte(c.Password)); err != nil {
		log.Fatalf("failed to hash password from config.ini", err.Error())
	}
	c.Password = string(h.Sum(nil))

	return nil
}
