package api

import (
	"errors"
	"github.com/phips4/img-proxy/worker/internal"
	"net/http"
	"net/http/httptest"
	"testing"
)

const urlEncodedUrl = "https%3A%2F%2Fexample.com%2Fimage.jpg"

func TestGetImageHandler(t *testing.T) {
	expectedBody := "mocked-img-data"
	mockCache := internal.NewCache()
	err := mockCache.Set("mocked-hash", []byte(expectedBody))
	if err != nil {
		t.Errorf("Expected error %v, but got %v", nil, err)
	}
	mockHasher := func(url string) (string, error) {
		return "mocked-hash", nil
	}

	// Create a request with a valid URL
	req := httptest.NewRequest("GET", "/?url="+urlEncodedUrl, nil)
	w := httptest.NewRecorder()

	// Call the handler function
	GetImage(mockCache, mockHasher)(w, req)

	// Check the response status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, w.Code)
	}

	// Check the response body
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body '%s', but got '%s'", expectedBody, w.Body.String())
	}
}

func TestGetImageHandlerNotFound(t *testing.T) {
	mockCache := internal.NewCache()
	mockHasher := func(url string) (string, error) {
		return "mocked-hash", nil
	}

	// Create a request with a valid URL
	req := httptest.NewRequest("GET", "/?url="+urlEncodedUrl, nil)
	w := httptest.NewRecorder()

	GetImage(mockCache, mockHasher)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, but got %d", http.StatusNotFound, w.Code)
	}

	expectedBody := "image not found\n"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body '%s', but got '%s'", expectedBody, w.Body.String())
	}
}

func TestGetImageHandlerInvalidUrl(t *testing.T) {
	mockCache := internal.NewCache()
	mockHasher := func(url string) (string, error) {
		return "mocked-hash", nil
	}

	invalidUrl := "invalid_url_com"
	req := httptest.NewRequest("GET", "/?url="+invalidUrl, nil)
	w := httptest.NewRecorder()

	GetImage(mockCache, mockHasher)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, w.Code)
	}

	expectedBody := "invalid url\n"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body '%s', but got '%s'", expectedBody, w.Body.String())
	}
}
func TestGetImageHandlerInvalidHasher(t *testing.T) {
	mockCache := internal.NewCache()
	mockHasher := func(url string) (string, error) {
		return "", errors.New("mocked hashing error")
	}

	req := httptest.NewRequest("GET", "/?url="+urlEncodedUrl, nil)
	w := httptest.NewRecorder()

	GetImage(mockCache, mockHasher)(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, but got %d", http.StatusInternalServerError, w.Code)
	}

	expectedBody := "mocked hashing error\n"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body '%s', but got '%s'", expectedBody, w.Body.String())
	}
}
