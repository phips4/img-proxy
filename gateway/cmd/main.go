package main

import (
	"github.com/phips4/img-proxy/gateway/internal"
	"github.com/phips4/img-proxy/gateway/internal/api"
	"github.com/phips4/img-proxy/gateway/internal/imageservice"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	addrs, err := net.InterfaceAddrs()
	must(err)

	ip, err := internal.FindIp(addrs)
	must(err)

	imgService := imageservice.NewService(time.Second * 10) //TODO: config
	cluster := internal.NewCluster()
	secret := secretFromEnv()
	go func() {
		// check if fist gateway instance is up, if not, wait a bit to avoid host resolution errors
		// because all containers can be started at the same time
		if _, err := net.LookupIP(hostlistFromEnv()[0]); err != nil {
			time.Sleep(time.Second)
		}

		err := cluster.Join(ip, secret, hostlistFromEnv())
		if err != nil {
			log.Fatalln("error joining cluster:", err.Error())
		}
		log.Println("joined cluster")
	}()

	http.HandleFunc("/image", api.HandleImage(cluster, imgService))
	http.HandleFunc("/health", api.HandleHealth(cluster))
	http.Handle("/metrics", promhttp.Handler())

	httpSrvAddr := net.JoinHostPort("", httpPortFromEnv()) //TODO: use host from config
	log.Println("listening on:", httpSrvAddr)
	log.Fatalln(http.ListenAndServe(httpSrvAddr, nil))
}

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func secretFromEnv() string {
	secret := os.Getenv("CLUSTER_SECRET")
	if secret == "" {
		panic("env CLUSTER_SECRET is not set")
	}
	return secret
}

func hostlistFromEnv() []string {
	hosts := os.Getenv("KNOWN_HOSTS")
	if hosts == "" {
		panic("env KNOWN_HOSTS is not set")
	}
	return strings.Split(hosts, ",")
}

func httpPortFromEnv() string {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		panic("env HTTP_PORT is not set")
	}
	return port
}
