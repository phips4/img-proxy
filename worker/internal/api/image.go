package api

import (
	"encoding/json"
	"errors"
	"github.com/phips4/img-proxy/worker/internal"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const internalErrorStr = "internal server error"

// ImageHandler gets an image from the local cache
func ImageHandler(cache *internal.Cache, hasherFunc internal.UrlHasherFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imgUrl, err := url.QueryUnescape(r.URL.Query().Get("url"))
		if err != nil {
			log.Println("ImageHandler (worker) error while un-escaping url:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}

		urlHash, err := hasherFunc(imgUrl)
		if err != nil {
			log.Println("ImageHandler (worker) error while hashing url:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}

		raw, err := cache.Get(urlHash)
		if err != nil {
			if strings.HasPrefix(err.Error(), "key not found:") {
				log.Println("ImageHandler (worker) cache miss")
				http.Error(w, "image not found", http.StatusNotFound)
				return
			}
			log.Println("ImageHandler (worker) error while getting image:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}

		if isJpeg(raw) {
			w.Header().Set("Content-Type", "image/jpeg")
		} else if isPng(raw) {
			w.Header().Set("Content-Type", "image/png")
		} else {
			log.Println("ImageHandler (worker) unknown image type")
			http.Error(w, "Unknown image type. Only jpeg and png are supported", http.StatusBadRequest)
			return
		}

		_, err = w.Write(raw)
		if err != nil {
			log.Println("ImageHandler (worker) error while writing response:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}
	}
}

// ImageCacheHandler handles uploading images to the local cache
func ImageCacheHandler(cache *internal.Cache, hFunc internal.UrlHasherFunc, dlFunc internal.DownloaderFunc) http.HandlerFunc {
	type bodyJson struct {
		Url string `json:"url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("ImageHandler (worker) error while reading body:", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		var bj bodyJson
		if err := json.Unmarshal(body, &bj); err != nil {
			log.Println("ImageHandler (worker) error while parsing body:", err)
			http.Error(w, internalErrorStr, http.StatusBadRequest)
			return
		}

		raw, err := dlFunc(bj.Url)
		if err != nil {
			if errors.Is(err, internal.ErrFileNotFound) {
				log.Println("ImageHandler (worker) file not found")
				http.NotFound(w, r)
				return
			}

			log.Println("ImageHandler (worker) error while downloading image:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}

		hashedUrl, err := hFunc(bj.Url)
		if err != nil {
			log.Println("ImageHandler (worker) error while hashing image:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}

		//TODO: do resizing, compression etc here

		if err = cache.Set(hashedUrl, raw); err != nil {
			log.Println("ImageHandler (worker) error while writing response:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}

		if _, err = w.Write(raw); err != nil {
			log.Println("ImageHandler (worker) error while writing response:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}
	}
}

func isJpeg(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	return data[0] == 0xFF && data[1] == 0xD8
}

func isPng(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	return string(data[0:8]) == "\x89PNG\r\n\x1a\n"
}
