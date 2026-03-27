// Package http contains the helper functions for writing Spin HTTP components
// in Go, as well as for sending outbound HTTP requests.
package http

import (
	"fmt"
	"net/http"
	"os"

	incominghandler "github.com/spinframework/spin-go-sdk/v3/exports/fermyon_spin_http_trigger_3_0_0/export_wasi_http_0_2_0_incoming_handler"
	_ "github.com/spinframework/spin-go-sdk/v3/exports/fermyon_spin_http_trigger_3_0_0/wit_exports"
	types "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_2_0_types"
)

func init() {
	incominghandler.Exports.Handle = wasiHandle
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

// handler is the function that will be called by the http trigger in Spin.
var handler = defaultHandler

// defaultHandler is a placeholder for returning a useful error to stderr when
// the handler is not set.
var defaultHandler = func(http.ResponseWriter, *http.Request) {
	fmt.Fprintln(os.Stderr, "http handler undefined")
}

// Handle sets the handler function for the http trigger.
// It must be set in an init() function.
func Handle(fn func(http.ResponseWriter, *http.Request)) {
	handler = fn
}

var wasiHandle = func(request *types.IncomingRequest, responseOut *types.ResponseOutparam) {
	// convert the incoming request to go's net/http type
	httpReq, err := NewHttpRequest(*request)
	if err != nil {
		//TODO(rajatjindal): return internal error from here
		fmt.Printf("failed to convert wasi/http/types.IncomingRequest to http.Request: %s\n", err)
		return
	}

	// convert the response outparam to go's net/http type
	httpRes := NewHttpResponseWriter(*responseOut)

	// run the user's handler
	handler(httpRes, httpReq)

	// ensure default status ok and response body are set
	_ = httpRes.reconcile()
}
