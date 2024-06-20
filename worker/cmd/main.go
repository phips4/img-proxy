package main

import (
	"encoding/base64"
	"github.com/hashicorp/memberlist"
	"github.com/phips4/img-proxy/worker/internal"
	"github.com/phips4/img-proxy/worker/internal/api"
	"github.com/phips4/img-proxy/worker/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	conf, err := internal.EnvConfig()
	if err != nil {
		log.Fatalln("error parsing env", err.Error())
		return
	}

	log.Printf("starting worker URL: %s:%s/ \n", conf.Host(), conf.Port())

	ml, err := joinCluster(conf.Host(), conf.Name(), conf.ClusterSecret(), conf.KnownHosts())
	if err != nil {
		log.Fatalln("could not join cluster: ", err.Error())
		return
	}

	cache := internal.NewCache()

	http.HandleFunc("/v1/image", middleware.OnlyGet(api.ImageHandler(cache, internal.Sha256UrlHasher)))
	http.HandleFunc("/v1/cache", middleware.OnlyPost(api.ImageCacheHandler(cache, internal.Sha256UrlHasher, internal.DownloadImg)))
	http.HandleFunc("/health", middleware.OnlyGet(api.HealthHandler(ml)))
	http.HandleFunc("/dashboard", middleware.OnlyGet(api.DashboardHandler(cache, ml)))
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		err := http.ListenAndServe(net.JoinHostPort(conf.Host(), conf.Port()), nil)
		if err != nil {
			log.Println(err.Error())
		}
	}()

	onShutdown(func() {
		if err := ml.Leave(time.Second * 5); err != nil {
			log.Println("error shutting down worker:", err.Error())
		}
	})
}

func onShutdown(closeHandler func()) {
	incomingSigs := make(chan os.Signal, 1)
	signal.Notify(incomingSigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)

	select {
	case <-incomingSigs:
		log.Println("shutting down worker")
		closeHandler()
	}
}

func joinCluster(host, name string, secret []byte, knownHosts []string) (*memberlist.Memberlist, error) {
	clusterKey, err := base64.StdEncoding.DecodeString(string(secret[:]))
	if err != nil {
		return nil, err
	}
	config := memberlist.DefaultWANConfig()
	config.BindAddr = host
	config.SecretKey = clusterKey
	config.Name = name

	ml, err := memberlist.Create(config)
	if err != nil {
		log.Fatalln("error creating cluster:" + err.Error())
		return nil, err
	}

	ml.LocalNode().Meta = []byte(`{"label":"worker"}`)

	// check if fist gateway instance is up, if not, wait a bit to avoid host resolution errors
	// because all containers are started at the same time
	if _, err := net.LookupIP(knownHosts[0]); err != nil {
		time.Sleep(time.Second)
	}

	_, err = ml.Join(knownHosts)
	if err != nil {
		log.Fatalln("worker couldn't join cluster: ", err.Error())
		return nil, err
	}

	log.Println("worker joined cluster")

	return ml, nil
}
