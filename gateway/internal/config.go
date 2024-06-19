package internal

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	hostList []string
	httpPort string
	secret   []byte
}

func ConfigFromEnv() (*Config, error) {
	conf := &Config{}

	if secret := os.Getenv("CLUSTER_SECRET"); secret != "" {
		conf.secret = []byte(secret)
	} else {
		return nil, errors.New("env CLUSTER_SECRET not set")
	}

	if hosts := os.Getenv("KNOWN_HOSTS"); hosts != "" {
		conf.hostList = strings.Split(hosts, ",")
	} else {
		return nil, errors.New("env KNOWN_HOSTS not set")
	}

	if httpPort := os.Getenv("HTTP_PORT"); httpPort != "" {
		conf.httpPort = httpPort
	} else {
		return nil, errors.New("env CLUSTER_SECRET not set")
	}

	return conf, nil
}

func (conf *Config) HostList() []string {
	return conf.hostList
}

func (conf *Config) HttpPort() string {
	return conf.httpPort
}

func (conf *Config) Secret() []byte {
	return conf.secret
}
