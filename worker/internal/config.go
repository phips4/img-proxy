package internal

import (
	"errors"
	"log"
	"os"
	"strings"
)

type AppConfig struct {
	clusterSecret []byte
	host          string
	port          string
	knownHosts    []string
	name          string
}

func EnvConfig() (*AppConfig, error) {
	conf := &AppConfig{}

	conf.clusterSecret = []byte(os.Getenv("CLUSTER_SECRET"))
	if conf.clusterSecret == nil || len(conf.clusterSecret) <= 0 {
		return nil, errors.New("env CLUSTER_SECRET is not set")
	}

	ip, err := GetIp()
	if err != nil {
		log.Fatal(err)
	}
	conf.host = ip

	conf.port = os.Getenv("PORT")
	if conf.port == "" {
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

func (c *AppConfig) ClusterSecret() []byte {
	return c.clusterSecret
}

func (c *AppConfig) Host() string {
	return c.host
}

func (c *AppConfig) Port() string {
	return c.port
}

func (c *AppConfig) KnownHosts() []string {
	return c.knownHosts
}

func (c *AppConfig) Name() string {
	return c.name
}
