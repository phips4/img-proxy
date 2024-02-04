package internal

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

var ErrFileNotFound = errors.New("file not found")

type DownloaderFunc func(input string) ([]byte, error)

func DownloadImg(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Second * 3,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Println("couldnt close download body: " + err.Error())
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrFileNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("HTTP request failed with status: " + resp.Status)
	}

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return imageBytes, nil
}
