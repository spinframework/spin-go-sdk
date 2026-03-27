package http

import (
	"fmt"
	"io"
	"net/http"

	outgoinghandler "github.com/spinframework/spin-go-sdk/v3/internal/wasi_http_0_2_0_outgoing_handler"
	types "github.com/spinframework/spin-go-sdk/v3/internal/wasi_http_0_2_0_types"
	wit "go.bytecodealliance.org/pkg/wit/types"
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
	or, err := NewOutgoingHttpRequest(req)
	if err != nil {
		return nil, err
	}

	result := outgoinghandler.Handle(&or, wit.None[*types.RequestOptions]())
	if result.IsErr() {
		return nil, fmt.Errorf("TODO: convert to readable error")
	}

	if result.IsErr() {
		return nil, fmt.Errorf("error is %v", result.Err())
	}

	okresult := result.Ok()

	//wait until resp is returned
	okresult.Subscribe().Block()

	incomingResp := okresult.Get()
	if incomingResp.IsNone() {
		return nil, fmt.Errorf("incoming resp is None")
	} else if incomingResp.Some().IsErr() {
		return nil, fmt.Errorf("error is %v", incomingResp.Some().Err())
	} else if incomingResp.Some().Ok().IsErr() {
		return nil, fmt.Errorf("error is %v", incomingResp.Some().Ok().Err())
	}

	okresp := incomingResp.Some().Ok().Ok()
	var body io.ReadCloser
	if consumeResult := okresp.Consume(); consumeResult.IsErr() {
		return nil, fmt.Errorf("failed to consume incoming request %s", consumeResult.Err())
	} else if streamResult := consumeResult.Ok().Stream(); streamResult.IsErr() {
		return nil, fmt.Errorf("failed to consume incoming requests's stream %s", streamResult.Err())
	} else {
		body = NewReadCloser(*streamResult.Ok())
	}

	resp := &http.Response{
		StatusCode: int(okresp.Status()),
		Body:       body,
	}

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
