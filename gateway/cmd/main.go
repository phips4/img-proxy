package main

import (
	"fmt"
	"github.com/phips4/img-proxy/gateway/internal"
	"github.com/phips4/img-proxy/gateway/internal/api"
	"github.com/phips4/img-proxy/gateway/internal/imageservice"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	secret := secretFromEnv()
	ip, err := internal.GetIp()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("listening on:", ip)
	imgService := imageservice.NewService(time.Second * 10) //TODO: config

	cluster := internal.NewCluster()
	go func() {
		// check if fist gateway instance is up, if not, wait a bit to avoid host resolution errors
		// because all containers are started at the same time
		if _, err := net.LookupIP(hostlistFromEnv()[0]); err != nil {
			time.Sleep(time.Second)
		}
		err := cluster.Join(ip, secret, hostlistFromEnv())
		if err != nil {
			log.Fatalln("error joining cluster:", err.Error())
		}
		log.Println("joined cluster")
	}()

	log.Println("server started")
	http.HandleFunc("/image", api.HandleImage(cluster, imgService))
	http.HandleFunc("/health", api.HandleHealth(cluster))

	httpSrvAddr := fmt.Sprintf(":%d", httpPortFromEnv())
	log.Println("starting http server", httpSrvAddr)
	log.Fatalln(http.ListenAndServe(httpSrvAddr, nil))
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

func httpPortFromEnv() int {
	port, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if err != nil {
		panic("env HTTP_PORT is not set")
	}
	return port
}
