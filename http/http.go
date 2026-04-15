// Package http contains the helper functions for writing Spin HTTP components
// in Go, as well as for sending outbound HTTP requests.
package http

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	handler "github.com/spinframework/spin-go-sdk/v3/exports/wasi_http_service_0_3_0_rc_2026_03_15/export_wasi_http_0_3_0_rc_2026_03_15_handler"
	_ "github.com/spinframework/spin-go-sdk/v3/exports/wasi_http_service_0_3_0_rc_2026_03_15/wit_exports"
	wasi "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

func init() {
	handler.Exports.Handle = wasiHandle
}

const (
	// The application base path.
	HeaderBasePath = "spin-base-path"
	// The component route pattern matched, _excluding_ any wildcard indicator.
	HeaderComponentRoot = "spin-component-route"
	// The full URL of the request. This includes full host and scheme information.
	HeaderFullUrl = "spin-full-url"
	// The part of the request path that was matched by the route (including
	// the base and wildcard indicator if present).
	HeaderMatchedRoute = "spin-matched-route"
	// The request path relative to the component route (including any base).
	HeaderPathInfo = "spin-path-info"
	// The component route pattern matched, as written in the component
	// manifest (that is, _excluding_ the base, but including the wildcard
	// indicator if present).
	HeaderRawComponentRoot = "spin-raw-component-route"
	// The client address for the request.
	HeaderClientAddr = "spin-client-addr"
)

// the function that will be called by the http trigger in Spin.
var handlerFn = defaultHandler

// defaultHandler is a placeholder for returning a useful error to stderr when
// the handler is not set.
var defaultHandler = func(http.ResponseWriter, *http.Request) {
	fmt.Fprintln(os.Stderr, "http handler undefined")
}

// Handle sets the handler function for the http trigger.
// It must be set in an init() function.
func Handle(fn func(http.ResponseWriter, *http.Request)) {
	handlerFn = fn
}

var wasiHandle = func(request *wasi.Request) wit.Result[*wasi.Response, wasi.ErrorCode] {
	httpRes := newHttpResponseWriter()

	go func() {
		defer httpRes.close()

		// convert the incoming request to go's net/http type
		httpReq, err := newHttpRequest(request)
		if err != nil {
			httpRes.channel <- wit.Err[*wasi.Response, wasi.ErrorCode](
				wasi.MakeErrorCodeInternalError(wit.Some(fmt.Sprintf(
					"failed to convert WASI Request to http.Request: %v\n",
					err,
				))),
			)
		} else {
			defer httpReq.Body.Close()

			// run the user's handler
			handlerFn(httpRes, httpReq)

			// if the user's handler never sent a response, we'll
			// send a default one here:
			if err := httpRes.send(); err != nil {
				httpRes.channel <- wit.Err[*wasi.Response, wasi.ErrorCode](
					wasi.MakeErrorCodeInternalError(wit.Some(fmt.Sprintf(
						"failed to produce a response: %v\n",
						err,
					))),
				)
			}
		}
	}()

	return (<-httpRes.channel)
}

func toWasiHeaders(headers http.Header) (*wasi.Fields, error) {
	fields := wasi.MakeFields()

	for key, vals := range headers {
		fieldVals := [][]uint8{}
		for _, val := range vals {
			fieldVals = append(fieldVals, []uint8(val))
		}

		if result := fields.Set(key, fieldVals); result.IsErr() {
			fields.Drop()
			switch result.Err().Tag() {
			case wasi.HeaderErrorInvalidSyntax:
				return nil, fmt.Errorf(
					"failed to set header %s to [%s]: invalid syntax",
					key,
					strings.Join(vals, ","),
				)
			case wasi.HeaderErrorForbidden:
				return nil, fmt.Errorf("failed to set forbidden header key %s", key)
			case wasi.HeaderErrorImmutable:
				return nil, fmt.Errorf("failed to set header on immutable header fields")
			default:
				return nil, fmt.Errorf("error setting header %s", key)
			}
		}
	}

	return fields, nil
}

func trailersFuture() *wit.FutureReader[wit.Result[wit.Option[*wasi.Fields], wasi.ErrorCode]] {
	tx, rx := wasi.MakeFutureResultOptionFieldsErrorCode()
	go tx.Write(wit.Ok[wit.Option[*wasi.Fields], wasi.ErrorCode](wit.None[*wasi.Fields]()))
	return rx
}

func unitFuture() *wit.FutureReader[wit.Result[wit.Unit, wasi.ErrorCode]] {
	tx, rx := wasi.MakeFutureResultUnitErrorCode()
	go tx.Write(wit.Ok[wit.Unit, wasi.ErrorCode](wit.Unit{}))
	return rx
}

func errorString(code wasi.ErrorCode) string {
	// TODO: make this human-readable:
	return fmt.Sprintf("%v", code)
}
