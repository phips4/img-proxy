package internal

import (
	"github.com/hashicorp/memberlist"
	"testing"
)

func TestCluster_Nodes(t *testing.T) {
	c := &Cluster{}
	c.memberlist = &memberlist.Memberlist{}

	got := len(c.Nodes())
	exp := 0
	if got != exp {
		t.Errorf("expexted: '%d', got: '%d'", exp, got)
	}
}
