package api

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/phips4/img-proxy/gateway/internal"
	"github.com/phips4/img-proxy/gateway/internal/workerservice"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func nodeIdFromImgUrl(url string, mod int) int {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	hashBytes := hasher.Sum(nil)

	hashInt := new(big.Int)
	hashInt.SetBytes(hashBytes)

	result := new(big.Int)
	result.Mod(hashInt, big.NewInt(int64(mod)))

	return int(result.Int64())
}

func HandleImage(cluster *internal.Cluster, service *workerservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		imgUrl, err := url.QueryUnescape(r.URL.Query().Get("url"))
		if err != nil {
			http.Error(w, "error un-escaping url", http.StatusBadRequest)
			return
		}

		if !strings.HasPrefix(imgUrl, "https") { // url encoded for https://
			http.Error(w, "invalid url: "+imgUrl, http.StatusBadRequest)
			return
		}

		clusterLen := len(cluster.Nodes())
		if clusterLen == 0 {
			http.Error(w, "cluster not available", http.StatusInternalServerError)
			return
		}

		node := nodeIdFromImgUrl(imgUrl, clusterLen)
		log.Println("nodeId from string is", node)

		w.Header().Set("Node-Id", strconv.Itoa(node))

		n := cluster.Nodes()[node]
		workerUrl := fmt.Sprintf("http://%s:%d", n.Addr.String(), 8080) //TODO:

		raw, err := service.GetImage(workerUrl, imgUrl)
		if errors.Is(err, workerservice.ErrNotFound) { // post image and update raw variable if not cached
			img, err := service.CacheImage(imgUrl)
			if err != nil {
				http.Error(w, "client responded with: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := w.Write(img); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusOK)
			return

		} else if err != nil {
			http.Error(w, "client responded with: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(raw); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandleHealth(cluster *internal.Cluster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var items []string

		for _, node := range cluster.Nodes() {
			items = append(items, node.Addr.String())
		}

		type Response struct {
			Nodes []string `json:"nodes"`
			Score int      `json:"score"`
		}

		resp := &Response{
			Nodes: items,
			Score: cluster.HealthScore(),
		}

		js, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(js); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
