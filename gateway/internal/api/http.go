package api

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/phips4/img-proxy/gateway/internal"
	"github.com/phips4/img-proxy/gateway/internal/imageservice"
	"github.com/phips4/img-proxy/gateway/internal/prom"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
)

func HandleImage(cluster internal.Cluster, service *imageservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prom.ImageHandlerHits.Inc()

		imgUrl, err := url.QueryUnescape(r.URL.Query().Get("url"))
		if err != nil {
			http.Error(w, "error un-escaping url", http.StatusBadRequest)
			prom.ImageHandlerErrors.Inc()
			return
		}

		if !strings.HasPrefix(imgUrl, "https") { // url encoded for https://
			http.Error(w, "invalid url: "+imgUrl, http.StatusBadRequest)
			prom.ImageHandlerErrors.Inc()
			return
		}

		clusterLen := len(cluster.WorkerNodes())
		if clusterLen == 0 {
			http.Error(w, "cluster not available", http.StatusInternalServerError)
			prom.ImageHandlerErrors.Inc()
			return
		}

		workerId := idFromUrl(imgUrl, clusterLen)
		worker := cluster.WorkerNodes()[workerId]
		workerUrl := fmt.Sprintf("http://%s:%d", worker.Addr.String(), 8080)

		log.Println("nodeId from string is", workerId, workerUrl)

		raw, err := service.GetImage(workerUrl, imgUrl)
		if errors.Is(err, imageservice.ErrNotFound) { // post image and update raw variable if not cached
			img, err := service.CacheImage(workerUrl, imgUrl)
			if err != nil {
				http.Error(w, "client responded with: "+err.Error(), http.StatusInternalServerError)
				prom.ImageHandlerErrors.Inc()
				return
			}

			if _, err := w.Write(img); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				prom.ImageHandlerErrors.Inc()
				return
			}

		} else if err != nil {
			http.Error(w, "client responded with: "+err.Error(), http.StatusInternalServerError)
			prom.ImageHandlerErrors.Inc()
			return
		}

		if _, err := w.Write(raw); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			prom.ImageHandlerErrors.Inc()
			return
		}
	}
}

// keep it simple for now
func idFromUrl(url string, mod int) int {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	hashBytes := hasher.Sum(nil)

	hashInt := new(big.Int)
	hashInt.SetBytes(hashBytes)

	result := new(big.Int)
	result.Mod(hashInt, big.NewInt(int64(mod)))

	return int(result.Int64())
}

type response struct {
	Nodes []string `json:"nodes"`
	Score int      `json:"score"`
}

func HandleHealth(cluster internal.Cluster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var nodes []string
		for _, node := range cluster.Nodes() {
			nodes = append(nodes, node.Addr.String())
		}

		resp := &response{
			Nodes: nodes,
			Score: cluster.HealthScore(),
		}

		jsn, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(jsn); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
