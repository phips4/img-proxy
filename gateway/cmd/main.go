package main

import (
	"fmt"
	"github.com/phips4/img-proxy/gateway/internal"
	"github.com/phips4/img-proxy/gateway/internal/api"
	"github.com/phips4/img-proxy/gateway/internal/workerservice"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	secret := secretFromEnv()
	ip := ipFromEnv()

	fmt.Println("listening on:", ip)

	cluster := internal.NewCluster()
	go func() {
		err := cluster.Join(ip, secret, hostlistFromEnv())
		if err != nil {
			log.Fatalln("error joining cluster:", err.Error())
		}
		log.Println("joined cluster")
	}()

	service := workerservice.NewService(time.Second * 10) //TODO: config

	log.Println("server started")
	http.HandleFunc("/image", api.HandleImage(cluster, service))
	http.HandleFunc("/health", api.HandleHealth(cluster))

	httpSrvAddr := fmt.Sprintf(":%d", httpPortFromEnv())
	log.Println("starting http server", httpSrvAddr)
	log.Fatalln(http.ListenAndServe(httpSrvAddr, nil))
}

func ipFromEnv() string {
	ip := os.Getenv("HOST")
	if ip == "" {
		panic("env CLUSTER_SECRET is not set")
	}
	return ip
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
