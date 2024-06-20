package api

import (
	"encoding/json"
	"github.com/hashicorp/memberlist"
	"log"
	"net"
	"net/http"
	"time"
)

type onlineHosts struct {
	Ip     string `json:"ip"`
	Status string `json:"status"`
}

// HealthHandler outputs the health status of cluster members
func HealthHandler(memberlist *memberlist.Memberlist) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var items []onlineHosts

		for _, member := range memberlist.Members() {
			hostName := member.Addr.String()
			portNum := "8080"

			_, err := net.DialTimeout("tcp", hostName+":"+portNum, 5*time.Second)
			if err != nil {
				items = append(items, onlineHosts{Ip: hostName + ":" + portNum, Status: "DOWN"})
			} else {
				items = append(items, onlineHosts{Ip: hostName + ":" + portNum, Status: "UP"})
			}
		}

		jsn, err := json.Marshal(items)
		if err != nil {
			log.Println("HealthHandler (worker) error marshalling json:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsn)
		if err != nil {
			log.Println("HealthHandler (worker) error writing response:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}
	}
}
