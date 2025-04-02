// Package http contains the helper functions for writing Spin HTTP components
// in TinyGo, as well as for sending outbound HTTP requests.
package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	incominghandler "github.com/spinframework/spin-go-sdk/v2/internal/wasi/http/v0.2.0/incoming-handler"
	"github.com/spinframework/spin-go-sdk/v2/internal/wasi/http/v0.2.0/types"
	"github.com/spinframework/spin-go-sdk/v2/wit"
)

// force wit files to be shipped with sdk dependency
var _ = wit.Wit

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

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router = httprouter.Router

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params = httprouter.Params

// Param is a single URL parameter, consisting of a key and a value.
type Param = httprouter.Param

// RouterHandle is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// wildcards (variables).
type RouterHandle = httprouter.Handle

// New returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func NewRouter() *Router {
	return httprouter.New()
}

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

var wasiHandle = func(request types.IncomingRequest, responseOut types.ResponseOutparam) {
	// convert the incoming request to go's net/http type
	httpReq, err := NewHttpRequest(request)
	if err != nil {
		//TODO(rajatjindal): return internal error from here
		fmt.Printf("failed to convert wasi/http/types.IncomingRequest to http.Request: %s\n", err)
		return
	}

	// convert the response outparam to go's net/http type
	httpRes := NewHttpResponseWriter(responseOut)

	// run the user's handler
	handler(httpRes, httpReq)

	// ensure default status ok and response body are set
	_ = httpRes.reconcile()
}
