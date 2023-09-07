package worker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type (
	// a little bit overkill

	ImageGetter interface {
		GetImage(urlHash string) ([]byte, error)
	}
	ImageCacher interface {
		CacheImage(url string) ([]byte, error)
	}
	Worker interface {
		ImageGetter
		ImageCacher
	}
	Service struct {
		Client *http.Client
	}
)

var (
	ErrNotFound = errors.New("not found")
)

func (s *Service) GetImage(imgUrl string) ([]byte, error) {
	endpointUrl := fmt.Sprintf("http://%s:%d/v1/image?url=%s", "172.23.0.2", 8080, url.QueryEscape(imgUrl))

	resp, err := s.Client.Get(endpointUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
	endpointUrl := fmt.Sprintf("http://%s:%d/v1/cache", "172.23.0.2", 8080)

	requestData := map[string]interface{}{"url": imgUrl}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	resp, err := s.Client.Post(endpointUrl, "application/json", bytes.NewBuffer(requestBody))
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
