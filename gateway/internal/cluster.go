package internal

import (
	"encoding/base64"
	"errors"
	"github.com/hashicorp/memberlist"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Cluster struct {
	id         int
	memberlist *memberlist.Memberlist
}

func NewCluster() *Cluster {
	return &Cluster{}
}

func (c *Cluster) Join(bindIP, clusterKey string, knownIPs []string) error {
	config := memberlist.DefaultWANConfig()
	config.BindAddr = bindIP
	config.SecretKey, _ = base64.StdEncoding.DecodeString(clusterKey)
	config.Name = bindIP

	ml, err := memberlist.Create(config)
	if err != nil {
		panic(err)
	}

	ml.LocalNode().Meta = []byte(`{"label":"gateway"}`)
	_, err = ml.Join(knownIPs)
	if err != nil {
		return errors.New("Failed to join cluster: " + err.Error())
	}

	c.memberlist = ml
	log.Printf("Joined the cluster")

	incomingSigs := make(chan os.Signal, 1)
	signal.Notify(incomingSigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)

	select {
	case <-incomingSigs:
		if err := ml.Leave(time.Second * 5); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cluster) Nodes() []*memberlist.Node {
	if c.memberlist == nil {
		log.Println("accessing Node() when memberlist is nil")
		return []*memberlist.Node{}
	}
	return c.memberlist.Members()
}

func (c *Cluster) HealthScore() int {
	return c.memberlist.GetHealthScore()
}
