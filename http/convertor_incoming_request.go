package http

import (
	"fmt"
	"net/http"

	. "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
	. "go.bytecodealliance.org/pkg/wit/types"
)

// convert the Request to an http.Request
func newHttpRequest(ir *Request) (*http.Request, error) {
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

	rx, trailers := RequestConsumeBody(ir, unitFuture())
	body := newReader(rx, trailers)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		body.Close()
		return nil, err
	}

	toHttpHeader(headers, &req.Header)

	return req, nil
}

func methodToString(m Method) (string, error) {
	switch m.Tag() {
	case MethodConnect:
		return "CONNECT", nil
	case MethodDelete:
		return "DELETE", nil
	case MethodGet:
		return "GET", nil
	case MethodHead:
		return "HEAD", nil
	case MethodOptions:
		return "OPTIONS", nil
	case MethodPatch:
		return "PATCH", nil
	case MethodPost:
		return "POST", nil
	case MethodPut:
		return "PUT", nil
	case MethodTrace:
		return "TRACE", nil
	case MethodOther:
		return m.Other(), fmt.Errorf("unknown http method 'other'")
	default:
		return "", fmt.Errorf("failed to convert http method")
	}
}

func toHttpHeader(src []Tuple2[string, []uint8], dest *http.Header) {
	for _, pair := range src {
		key := pair.F0
		value := string(pair.F1)
		dest.Add(key, value)
	}
}
