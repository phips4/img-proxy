package imageservice

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

const testBytes = "bytes"

type mockClient struct{}

func (c *mockClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     testBytes,
		Body:       io.NopCloser(bytes.NewBufferString(testBytes)),
	}, nil
}

func TestService_GetImage(t *testing.T) {
	service := &Service{client: &mockClient{}}

	img, err := service.GetImage("notaurl:2929", "https://notarealhost.com/image.png")
	if err != nil {
		t.Error("error is not null", err.Error())
	}

	if len(img) != len(testBytes) {
		t.Error("response is not equal to mocked data. expected:", len(testBytes), "got:", len(img))
	}
}
