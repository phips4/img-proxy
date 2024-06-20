package internal

import (
	"errors"
	"log"
	"os"
	"strings"
)

type AppConfig struct {
	secret     []byte
	host       string
	httpPort   string
	knownHosts []string
	name       string
}

func ConfigFromEnv() (*AppConfig, error) {
	conf := &AppConfig{}

	if secret := os.Getenv("CLUSTER_SECRET"); secret != "" {
		conf.secret = []byte(secret)
	} else {
		return nil, errors.New("env CLUSTER_SECRET is not set")
	}

	ip, err := GetIp()
	if err != nil {
		log.Fatal(err)
	}
	conf.host = ip

	conf.httpPort = os.Getenv("PORT")
	if conf.httpPort == "" {
		return nil, errors.New("env PORT is not set")
	}

	conf.knownHosts = strings.Split(os.Getenv("KNOWN_HOSTS"), ",")
	if len(conf.knownHosts) <= 0 {
		return nil, errors.New("env KNOWN_HOSTS is not set")
	}

	conf.name = os.Getenv("NAME")
	if conf.name == "" {
		conf.name = conf.host
	}

	return conf, nil
}

func (c *AppConfig) Secret() []byte {
	return c.secret
}

func (c *AppConfig) Host() string {
	return c.host
}

func (c *AppConfig) HttpPort() string {
	return c.httpPort
}

func (c *AppConfig) KnownHosts() []string {
	return c.knownHosts
}

func (c *AppConfig) Name() string {
	return c.name
}
