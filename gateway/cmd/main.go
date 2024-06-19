package main

import (
	"github.com/phips4/img-proxy/gateway/internal"
	"github.com/phips4/img-proxy/gateway/internal/api"
	"github.com/phips4/img-proxy/gateway/internal/imageservice"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	addrs, err := net.InterfaceAddrs()
	must(err)

	ip, err := internal.FindIp(addrs)
	must(err)

	conf, err := internal.ConfigFromEnv()
	must(err)

	imgService := imageservice.NewService(time.Second * 10) //TODO: config
	cluster := internal.NewCluster()
	go func() {
		// check if fist gateway instance is up, if not, wait a bit to avoid host resolution errors
		// because all containers can be started at the same time
		if _, err := net.LookupIP(conf.HostList()[0]); err != nil {
			time.Sleep(time.Second)
		}

		err := cluster.Join(ip, conf.Secret(), conf.HostList())
		if err != nil {
			log.Fatalln("error joining cluster:", err.Error())
		}
		log.Println("joined cluster")
	}()

	http.HandleFunc("/image", api.HandleImage(cluster, imgService))
	http.HandleFunc("/health", api.HandleHealth(cluster))
	http.Handle("/metrics", promhttp.Handler())

	httpSrvAddr := net.JoinHostPort("", conf.HttpPort()) //TODO: use host from config
	log.Println("listening on:", httpSrvAddr)
	log.Fatalln(http.ListenAndServe(httpSrvAddr, nil))
}

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
