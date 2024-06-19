package internal

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfigFromEnv(t *testing.T) {
	expectedSecret := []byte("walrus123")
	const expectedHosts = "host1,host2,host3"
	const expectedPort = "8080"

	t.Setenv("CLUSTER_SECRET", string(expectedSecret))
	t.Setenv("KNOWN_HOSTS", expectedHosts)
	t.Setenv("HTTP_PORT", expectedPort)

	conf, err := ConfigFromEnv()
	if err != nil {
		t.Error("expected:", nil, "got:", err)
	}

	if !bytes.Equal(expectedSecret, conf.Secret()) {
		t.Error("expected:", string(expectedSecret), "got:", string(conf.Secret()))
	}

	gotHostsStr := strings.Join(conf.HostList(), ",")
	if expectedHosts != gotHostsStr {
		t.Error("expected:", expectedHosts, "got:", gotHostsStr)
	}

	if expectedPort != conf.HttpPort() {
		t.Error("expected:", expectedPort, "got:", conf.HttpPort())
	}
}

func TestConfig_Hosts(t *testing.T) {
	conf := &AppConfig{hostList: []string{"host1", "host2", "host3"}}
	expectedHostsStr := strings.Join(conf.HostList(), ",")
	gotHostStr := strings.Join(conf.HostList(), ",")

	if expectedHostsStr != gotHostStr {
		t.Error("expected:", expectedHostsStr, "got:", gotHostStr)
	}
}

func TestConfig_HttpPort(t *testing.T) {
	expectedPort := "8080"
	conf := &AppConfig{httpPort: expectedPort}
	gotPort := conf.HttpPort()

	if expectedPort != gotPort {
		t.Error("expected:", expectedPort, "got:", gotPort)
	}
}

func TestConfig_Secret(t *testing.T) {
	expected := []byte("helloworld")
	conf := &AppConfig{secret: expected}
	got := conf.Secret()

	if !bytes.Equal(expected, got) {
		t.Error("expected:", string(expected), "got:", string(got))
	}
}
