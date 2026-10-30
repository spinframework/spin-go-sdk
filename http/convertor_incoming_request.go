package http

import (
	"fmt"
	"net/http"

	wasi "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_types"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

// convert the wasi.Request to an http.Request
func newHttpRequest(ir *wasi.Request) (*http.Request, error) {
	defer ir.Drop()

	method, err := methodToString(ir.GetMethod())
	if err != nil {
		return nil, err
	}

	var url string
	if pathWithQuery := ir.GetPathWithQuery(); pathWithQuery.IsNone() {
		url = ""
	} else {
		url = pathWithQuery.Some()
	}

	headerResource := ir.GetHeaders()
	headers := headerResource.CopyAll()
	headerResource.Drop()

	rx, trailers := wasi.RequestConsumeBody(ir, unitFuture())
	body := newReader(rx, trailers)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		body.Close()
		return nil, err
	}

	toHttpHeader(headers, &req.Header)

	return req, nil
}

func methodToString(m wasi.Method) (string, error) {
	switch m.Tag() {
	case wasi.Method_Connect:
		return "CONNECT", nil
	case wasi.Method_Delete:
		return "DELETE", nil
	case wasi.Method_Get:
		return "GET", nil
	case wasi.Method_Head:
		return "HEAD", nil
	case wasi.Method_Options:
		return "OPTIONS", nil
	case wasi.Method_Patch:
		return "PATCH", nil
	case wasi.Method_Post:
		return "POST", nil
	case wasi.Method_Put:
		return "PUT", nil
	case wasi.Method_Trace:
		return "TRACE", nil
	case wasi.Method_Other:
		return m.Other(), fmt.Errorf("unknown http method 'other'")
	default:
		return "", fmt.Errorf("failed to convert http method")
	}
}

func toHttpHeader(src []wit.Tuple2[string, []uint8], dest *http.Header) {
	for _, pair := range src {
		key := pair.F0
		value := string(pair.F1)
		dest.Add(key, value)
	}
}
