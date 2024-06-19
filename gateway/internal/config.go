package internal

import (
	"errors"
	"os"
	"strings"
)

type AppConfig struct {
	hostList []string
	httpPort string
	secret   []byte
}

func ConfigFromEnv() (*AppConfig, error) {
	conf := &AppConfig{}

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

func (conf *AppConfig) HostList() []string {
	return conf.hostList
}

func (conf *AppConfig) HttpPort() string {
	return conf.httpPort
}

func (conf *AppConfig) Secret() []byte {
	return conf.secret
}
