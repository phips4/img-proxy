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

const internalErrStr = "internal server error"

// ImageHandler gets a cached image from the worker cluster
func ImageHandler(cluster internal.Cluster, service *imageservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prom.ImageHandlerHits.Inc()

		imgUrl, err := url.QueryUnescape(r.URL.Query().Get("url"))
		if err != nil {
			log.Println("ImageHandler (gateway) error un-escaping url:", err)
			http.Error(w, internalErrStr, http.StatusBadRequest)
			prom.ImageHandlerErrors.Inc()
			return
		}

		if !strings.HasPrefix(imgUrl, "https") { // url encoded for https://
			log.Println("ImageHandler (gateway) error url does not start with https")
			http.Error(w, "invalid url: "+imgUrl, http.StatusBadRequest)
			prom.ImageHandlerErrors.Inc()
			return
		}

		clusterLen := len(cluster.WorkerNodes())
		if clusterLen == 0 {
			log.Println("ImageHandler (gateway) error cluster not available")
			http.Error(w, internalErrStr, http.StatusInternalServerError)
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
				log.Println("ImageHandler (gateway) error client responded with:", err)
				http.Error(w, internalErrStr, http.StatusInternalServerError)
				prom.ImageHandlerErrors.Inc()
				return
			}

			if _, err := w.Write(img); err != nil {
				log.Println("ImageHandler (gateway) error writing response:", err)
				http.Error(w, internalErrStr, http.StatusInternalServerError)
				prom.ImageHandlerErrors.Inc()
				return
			}

		} else if err != nil {
			log.Println("ImageHandler (gateway) error getting image:", err)
			http.Error(w, internalErrStr, http.StatusInternalServerError)
			prom.ImageHandlerErrors.Inc()
			return
		}

		if _, err := w.Write(raw); err != nil {
			log.Println("ImageHandler (gateway) error writing response:", err)
			http.Error(w, internalErrStr, http.StatusInternalServerError)
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

// HealthHandler outputs the health score of the cluster
func HealthHandler(cluster internal.Cluster) http.HandlerFunc {
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
			log.Println("HealthHandler (gateway) error marshaling response:", err)
			http.Error(w, internalErrStr, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(jsn); err != nil {
			log.Println("HealthHandler (gateway) error writing response:", err)
			http.Error(w, internalErrStr, http.StatusInternalServerError)
			return
		}
	}
}
