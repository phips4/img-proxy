package api

import (
	"encoding/json"
	"github.com/hashicorp/memberlist"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mockCluster struct{}

func (c *mockCluster) Join(bindIP, clusterKey string, knownIPs []string) error {
	return nil
}

func (c *mockCluster) Nodes() []*memberlist.Node {
	return []*memberlist.Node{
		{Addr: net.IP("192.168.1.1:8080").To16()},
		{Addr: net.IP("192.168.1.2:8081").To16()},
	}
}

func (c *mockCluster) WorkerNodes() []*memberlist.Node {
	return []*memberlist.Node{
		{Addr: net.IP("192.168.1.1:8080").To16()},
		{Addr: net.IP("192.168.1.2:8081").To16()},
	}
}

func (c *mockCluster) HealthScore() int {
	// Mock implementation for HealthScore method
	return 90
}

func TestHandleHealth(t *testing.T) {
	mock := &mockCluster{}
	handler := HandleHealth(mock)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("HandleHealth() returned wrong status code: got %v, want %v", rr.Code, http.StatusOK)
	}

	expected := &response{
		Nodes: []string{
			net.IP("192.168.1.1:8080").To16().String(),
			net.IP("192.168.1.2:8081").To16().String(),
		},
		Score: 90,
	}

	var resp response
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Errorf("HandleHealth() error decoding response body: %v", err)
	}

	if !reflect.DeepEqual(&resp, expected) {
		t.Errorf("HandleHealth() returned unexpected body: got %+v, want %+v", &resp, expected)
	}
}
