package api

import (
	"encoding/json"
	"github.com/hashicorp/memberlist"
	"net"
	"net/http"
	"time"
)

type onlineHosts struct {
	Ip     string `json:"ip"`
	Status string `json:"status"`
}

func HandleHealth(memberlist *memberlist.Memberlist) http.HandlerFunc {
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

		js, err := json.Marshal(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
