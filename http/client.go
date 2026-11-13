package http

import (
	"fmt"
	"io"
	"net/http"

	client "github.com/spinframework/spin-go-sdk/v4/imports/wasi_http_0_3_2_client"
	next "github.com/spinframework/spin-go-sdk/v4/imports/wasi_http_0_3_2_handler"
	wasi "github.com/spinframework/spin-go-sdk/v4/imports/wasi_http_0_3_2_types"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

// NewTransport returns an [http.RoundTripper] backed by the Spin SDK.
func NewTransport() http.RoundTripper {
	return &Transport{}
}

// Transport implements [http.RoundTripper] using the Spin SDK.
type Transport struct{}

// RoundTrip executes a single HTTP transaction using the Spin SDK.
func (r *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return Send(req)
}

// NewClient returns a new HTTP client compatible with the Spin SDK.
func NewClient() *http.Client {
	return &http.Client{
		Transport: &Transport{},
	}
}

// Send sends an HTTP request using the Spin SDK and returns the response.
func Send(req *http.Request) (*http.Response, error) {
	request, err := newOutgoingHttpRequest(req)
	if err != nil {
		return nil, err
	}
	defer request.Drop()

	return toHttpResponse(client.Send(request), "error sending request")
}

// Next forwards an HTTP request to the next handler in the middleware chain
// and returns its response.
//
// Components calling Next must be built against the
// `spin:up/http-middleware@4.1.0` world, e.g.
// `go tool componentize-go -w spin:up/http-middleware@4.1.0 build`.
func Next(req *http.Request) (*http.Response, error) {
	request, err := newOutgoingHttpRequest(req)
	if err != nil {
		return nil, err
	}
	defer request.Drop()

	return toHttpResponse(next.Handle(request), "error forwarding request to next handler")
}

func toHttpResponse(result wit.Result[*wasi.Response, wasi.ErrorCode], errContext string) (*http.Response, error) {
	if result.IsErr() {
		return nil, fmt.Errorf("%s: %s", errContext, errorString(result.Err()))
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

// Get issues a GET request to the specified URL using the Spin SDK.
func Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return Send(req)
}

// Post issues a POST request to the specified URL using the Spin SDK
// with the given content type and body.
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
