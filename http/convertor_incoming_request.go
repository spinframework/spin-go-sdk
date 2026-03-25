package http

import (
	"fmt"
	"io"
	"net/http"

	types "github.com/spinframework/spin-go-sdk/v3/internal/wasi_http_0_2_0_types"
)

type IncomingRequest = types.IncomingRequest

// convert the IncomingRequest to http.Request
func NewHttpRequest(ir IncomingRequest) (req *http.Request, err error) {
	// convert the http method to string
	method, err := methodToString(ir.Method())
	if err != nil {
		return nil, err
	}

	// convert the path with query to a url
	var url string
	if pathWithQuery := ir.PathWithQuery(); pathWithQuery.IsNone() {
		url = ""
	} else {
		url = pathWithQuery.Some()
	}

	// convert the body to a reader
	var body io.Reader
	if consumeResult := ir.Consume(); consumeResult.IsErr() {
		return nil, fmt.Errorf("failed to consume incoming request %s", consumeResult.Err())
	} else if streamResult := consumeResult.Ok().Stream(); streamResult.IsErr() {
		return nil, fmt.Errorf("failed to consume incoming requests's stream %s", streamResult.Err())
	} else {
		body = NewReader(*streamResult.Ok())
	}

	// create a new request
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// update additional fields
	toHttpHeader(*ir.Headers(), &req.Header)

	return req, nil
}

func methodToString(m types.Method) (string, error) {
	switch m.Tag() {
	case types.MethodConnect:
		return "CONNECT", nil
	case types.MethodDelete:
		return "DELETE", nil
	case types.MethodGet:
		return "GET", nil
	case types.MethodHead:
		return "HEAD", nil
	case types.MethodOptions:
		return "OPTIONS", nil
	case types.MethodPatch:
		return "PATCH", nil
	case types.MethodPost:
		return "POST", nil
	case types.MethodPut:
		return "PUT", nil
	case types.MethodTrace:
		return "TRACE", nil
	case types.MethodOther:
		return m.Other(), fmt.Errorf("unknown http method 'other'")
	default:
		return "", fmt.Errorf("failed to convert http method")
	}
}

func toHttpHeader(src types.Fields, dest *http.Header) {
	for _, f := range src.Entries() {
		key := f.F0
		value := string(f.F1)
		dest.Add(key, value)
	}
}
