package imageservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type (
	ImageGetter interface {
		GetImage(urlHash string) ([]byte, error)
	}

	ImageCacher interface {
		CacheImage(workerUrl, url string) ([]byte, error)
	}

	Worker interface {
		ImageGetter
		ImageCacher
	}

	HttpClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	Service struct {
		client HttpClient
	}
)

var (
	ErrNotFound = errors.New("not found")
)

func NewService(timeout time.Duration) *Service {
	return &Service{client: &http.Client{Timeout: timeout}}
}

func (s *Service) GetImage(workerUrl, imgUrl string) ([]byte, error) {
	endpointUrl := fmt.Sprintf("%s/?url=%s", workerUrl, url.QueryEscape(imgUrl))
	log.Println("downloading from ", workerUrl)
	req, err := http.NewRequest(http.MethodGet, endpointUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Println("error closing GetImage body: ", err.Error())
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %s", resp.Status)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

func (s *Service) CacheImage(workerUrl, imgUrl string) ([]byte, error) {
	endpointUrl := fmt.Sprintf("%s/v1/cache", workerUrl) //TODO: url

	requestBody, err := json.Marshal(map[string]interface{}{"url": imgUrl})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpointUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %s", resp.Status)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return raw, nil
}
