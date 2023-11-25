package workerservice

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
	// a little bit overkill

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
		Client HttpClient
	}
)

var (
	ErrNotFound = errors.New("not found")
)

func NewService(timeout time.Duration) *Service {
	return &Service{Client: &http.Client{Timeout: timeout}}
}

func (s *Service) GetImage(workerUrl, imgUrl string) ([]byte, error) {
	endpointUrl := fmt.Sprintf("%s/?url=%s", workerUrl, url.QueryEscape(imgUrl)) //TODO: url
	req, err := http.NewRequest(http.MethodGet, endpointUrl, nil)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
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

func (s *Service) CacheImage(imgUrl string) ([]byte, error) {
	endpointUrl := fmt.Sprintf("http://%s:%d/v1/cache", "172.18.0.2", 8080) //TODO: url

	requestData := map[string]interface{}{"url": imgUrl}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpointUrl, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
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
