package internal

import (
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/memberlist"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Cluster interface {
	Join(bindIP, clusterKey string, knownIPs []string) error
	Nodes() []*memberlist.Node
	WorkerNodes() []*memberlist.Node
	HealthScore() int
}

type ClusterImpl struct {
	memberlist *memberlist.Memberlist
}

func NewCluster() *ClusterImpl {
	return &ClusterImpl{}
}

func (c *ClusterImpl) Join(bindIP, clusterKey string, knownIPs []string) error {
	config := memberlist.DefaultWANConfig()
	config.BindAddr = bindIP
	config.SecretKey, _ = base64.StdEncoding.DecodeString(clusterKey)
	config.Name = bindIP

	ml, err := memberlist.Create(config)
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	ml.LocalNode().Meta = []byte(`{"label":"gateway"}`)
	_, err = ml.Join(knownIPs)
	if err != nil {
		return fmt.Errorf("failed to join cluster: %w", err)
	}

	c.memberlist = ml
	log.Printf("joined the cluster")

	incomingSigs := make(chan os.Signal, 1)
	signal.Notify(incomingSigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)

	select {
	case <-incomingSigs:
		if err := ml.Leave(time.Second * 5); err != nil {
			return fmt.Errorf("error leaving cluster: %w", err)
		}
	}
	return nil
}

func (c *ClusterImpl) Nodes() []*memberlist.Node {
	if c.memberlist == nil {
		return []*memberlist.Node{}
	}
	return c.memberlist.Members()
}

func (c *ClusterImpl) WorkerNodes() []*memberlist.Node {
	if c.memberlist == nil {
		return []*memberlist.Node{}
	}

	var workers []*memberlist.Node
	for _, n := range c.memberlist.Members() {
		if strings.Contains(string(n.Meta), "worker") {
			workers = append(workers, n)
		}
	}

	return workers
}

func (c *ClusterImpl) HealthScore() int {
	return c.memberlist.GetHealthScore()
}
