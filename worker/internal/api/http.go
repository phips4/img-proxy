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

func GetImage(cache *internal.Cache, hasherFunc internal.UrlHasherFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imgUrl, err := url.QueryUnescape(r.URL.Query().Get("url"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("error while un-escaping url: ", err.Error())
			return
		}

		urlHash, err := hasherFunc(imgUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("error while hashing url: ", err.Error())
			return
		}

		raw, err := cache.Get(urlHash)
		if err != nil {
			if strings.HasPrefix(err.Error(), "key not found:") {
				http.Error(w, "image not found", http.StatusNotFound) //TODO: better error handling
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("error while getting image: ", err.Error())
			return
		}

		if isJpeg(raw) {
			w.Header().Set("Content-Type", "image/jpeg")
		} else if isPng(raw) {
			w.Header().Set("Content-Type", "image/png")
		} else {
			http.Error(w, "Unknown image type. Only jpeg and png are supported", http.StatusBadRequest)
			return
		}

		_, err = w.Write(raw)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func PostCacheImage(cache *internal.Cache, hFunc internal.UrlHasherFunc, dlFunc internal.DownloaderFunc) http.HandlerFunc {
	type bodyJson struct {
		Url string `json:"url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		var bj bodyJson
		if err := json.Unmarshal(body, &bj); err != nil {
			http.Error(w, "Failed to unmarshal JSON data:"+err.Error(), http.StatusBadRequest)
			return
		}

		raw, err := dlFunc(bj.Url)
		if err != nil {
			if errors.Is(err, internal.ErrFileNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, "Failed to download image "+err.Error(), http.StatusInternalServerError)
			log.Println("error downloading image:", err.Error())
			return
		}

		hashedUrl, err := hFunc(bj.Url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("error hashing image:", err.Error())
			return
		}

		//TODO: do resizing, compression etc here

		if err = cache.Set(hashedUrl, raw); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err = w.Write(raw); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
