package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/hashicorp/memberlist"
	"github.com/phips4/img-proxy/worker/internal"
	"github.com/phips4/img-proxy/worker/internal/api"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	secret := os.Getenv("CLUSTER_SECRET")
	if secret == "" {
		panic("env CLUSTER_SECRET is not set")
	}
	ip, err := externalIP()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("(worker) using: ", ip)

	initCluster(ip, "8080", secret)
}

func initCluster(bindIP, httpPort, secret string) {
	clusterKey, _ := base64.StdEncoding.DecodeString(secret)
	config := memberlist.DefaultWANConfig()
	config.BindAddr = bindIP
	config.SecretKey = clusterKey
	config.Name = bindIP

	ml, err := memberlist.Create(config)
	if err != nil {
		panic(err)
	}

	ml.LocalNode().Meta = []byte(`{"label":"worker"}`)

	_, err = ml.Join([]string{"172.18.0.2", "172.18.0.3"}) //TODO:
	if err != nil {
		log.Println("couldn't join cluster: ", err.Error())
		return
	}

	log.Printf("new cluster created. key: %s\n", base64.StdEncoding.EncodeToString(clusterKey))

	cache := internal.NewCache()

	http.HandleFunc("/v1/image", api.GetImage(cache, internal.Sha256UrlHasher))
	http.HandleFunc("/v1/cache", api.PostCacheImage(cache, internal.Sha256UrlHasher, internal.DownloadImg))
	http.HandleFunc("/health", api.HandleHealth(ml))
	http.HandleFunc("/dashboard", api.HandleDashboard(cache, ml))

	go func() {
		err := http.ListenAndServe("0.0.0.0:"+httpPort, nil) //TODO: config
		if err != nil {
			log.Println(err.Error())
		}
	}()

	log.Printf("webserver is up. URL: %s:%s/ \n", bindIP, httpPort)

	incomingSigs := make(chan os.Signal, 1)
	signal.Notify(incomingSigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)

	select {
	case <-incomingSigs:
		if err := ml.Leave(time.Second * 5); err != nil {
			panic(err)
		}
	}
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP

			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP

			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
