package http

import (
	"fmt"
	"io"
	"net/http"

	client "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_client"
	wasi "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
)

// NewTransport returns http.RoundTripper backed by Spin SDK
func NewTransport() http.RoundTripper {
	return &Transport{}
}

// Transport implements http.RoundTripper
type Transport struct{}

// RoundTrip makes roundtrip using Spin SDK
func (r *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return Send(req)
}

// NewClient returns a new HTTP client compatible with the Spin SDK
func NewClient() *http.Client {
	return &http.Client{
		Transport: &Transport{},
	}
}

func Send(req *http.Request) (*http.Response, error) {
	request, err := newOutgoingHttpRequest(req)
	if err != nil {
		return nil, err
	}
	defer request.Drop()

	result := client.Send(request)
	if result.IsErr() {
		return nil, fmt.Errorf("error sending request: %s", errorString(result.Err()))
	}

	response := result.Ok()
	status := response.GetStatusCode()

	headerResource := response.GetHeaders()
	headers := headerResource.CopyAll()
	headerResource.Drop()

	rx, trailers := wasi.ResponseConsumeBody(response, unitFuture())
	body := newReader(rx, trailers)

	resp := &http.Response{
		StatusCode: int(status),
		Body:       body,
		Header:     http.Header{},
	}

	toHttpHeader(headers, &resp.Header)

	return resp, nil
}

func Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return Send(req)
}

func Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return Send(req)
}
