package main

import (
	"strconv"
	"strings"
	"testing"
)

func Test_hostlistFromEnv(t *testing.T) {
	const expected = "a_host_1,b_host_1"
	t.Setenv("KNOWN_HOSTS", expected)
	got := strings.Join(hostlistFromEnv(), ",")

	if got != expected {
		t.Error("expected:", expected, "got:", got)
	}
}

func Test_httpPortFromEnv(t *testing.T) {
	const expected = 9090
	t.Setenv("HTTP_PORT", strconv.Itoa(expected))
	got := httpPortFromEnv()

	if got != expected {
		t.Error("expected:", expected, "got:", got)
	}
}

func Test_secretFromEnv(t *testing.T) {
	const expected = "secretsecretsecret"
	t.Setenv("CLUSTER_SECRET", expected)
	got := secretFromEnv()

	if got != expected {
		t.Error("expected:", expected, "got:", got)
	}
}
