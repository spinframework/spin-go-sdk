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
	// HeaderBasePath is the application base path.
	HeaderBasePath = "spin-base-path"
	// HeaderComponentRoot is the component route pattern matched, excluding any wildcard indicator.
	HeaderComponentRoot = "spin-component-route"
	// HeaderFullUrl is the full URL of the request, including full host and scheme information.
	HeaderFullUrl = "spin-full-url"
	// HeaderMatchedRoute is the part of the request path that was matched by the route,
	// including the base and wildcard indicator if present.
	HeaderMatchedRoute = "spin-matched-route"
	// HeaderPathInfo is the request path relative to the component route, including any base.
	HeaderPathInfo = "spin-path-info"
	// HeaderRawComponentRoot is the component route pattern matched, as written in the component
	// manifest. It excludes the base but includes the wildcard indicator if present.
	HeaderRawComponentRoot = "spin-raw-component-route"
	// HeaderClientAddr is the client address for the request.
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

			httpRes.writeTrailers()
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
	switch code.Tag() {
	case wasi.ErrorCodeDnsTimeout:
		return "DNS timeout"
	case wasi.ErrorCodeDnsError:
		p := code.DnsError()
		var parts []string
		if p.Rcode.IsSome() {
			parts = append(parts, fmt.Sprintf("rcode=%s", p.Rcode.Some()))
		}
		if p.InfoCode.IsSome() {
			parts = append(parts, fmt.Sprintf("info-code=%d", p.InfoCode.Some()))
		}
		if len(parts) == 0 {
			return "DNS error"
		}
		return "DNS error (" + strings.Join(parts, ", ") + ")"
	case wasi.ErrorCodeDestinationNotFound:
		return "destination not found"
	case wasi.ErrorCodeDestinationUnavailable:
		return "destination unavailable"
	case wasi.ErrorCodeDestinationIpProhibited:
		return "destination IP prohibited"
	case wasi.ErrorCodeDestinationIpUnroutable:
		return "destination IP unroutable"
	case wasi.ErrorCodeConnectionRefused:
		return "connection refused"
	case wasi.ErrorCodeConnectionTerminated:
		return "connection terminated"
	case wasi.ErrorCodeConnectionTimeout:
		return "connection timeout"
	case wasi.ErrorCodeConnectionReadTimeout:
		return "connection read timeout"
	case wasi.ErrorCodeConnectionWriteTimeout:
		return "connection write timeout"
	case wasi.ErrorCodeConnectionLimitReached:
		return "connection limit reached"
	case wasi.ErrorCodeTlsProtocolError:
		return "TLS protocol error"
	case wasi.ErrorCodeTlsCertificateError:
		return "TLS certificate error"
	case wasi.ErrorCodeTlsAlertReceived:
		p := code.TlsAlertReceived()
		var parts []string
		if p.AlertId.IsSome() {
			parts = append(parts, fmt.Sprintf("alert-id=%d", p.AlertId.Some()))
		}
		if p.AlertMessage.IsSome() {
			parts = append(parts, fmt.Sprintf("alert-message=%s", p.AlertMessage.Some()))
		}
		if len(parts) == 0 {
			return "TLS alert received"
		}
		return "TLS alert received (" + strings.Join(parts, ", ") + ")"
	case wasi.ErrorCodeHttpRequestDenied:
		return "HTTP request denied"
	case wasi.ErrorCodeHttpRequestLengthRequired:
		return "HTTP request length required"
	case wasi.ErrorCodeHttpRequestBodySize:
		v := code.HttpRequestBodySize()
		if v.IsSome() {
			return fmt.Sprintf("HTTP request body size: %d", v.Some())
		}
		return "HTTP request body size"
	case wasi.ErrorCodeHttpRequestMethodInvalid:
		return "HTTP request method invalid"
	case wasi.ErrorCodeHttpRequestUriInvalid:
		return "HTTP request URI invalid"
	case wasi.ErrorCodeHttpRequestUriTooLong:
		return "HTTP request URI too long"
	case wasi.ErrorCodeHttpRequestHeaderSectionSize:
		v := code.HttpRequestHeaderSectionSize()
		if v.IsSome() {
			return fmt.Sprintf("HTTP request header section size: %d", v.Some())
		}
		return "HTTP request header section size"
	case wasi.ErrorCodeHttpRequestHeaderSize:
		v := code.HttpRequestHeaderSize()
		if v.IsSome() {
			return "HTTP request header size " + fieldSizeString(v.Some())
		}
		return "HTTP request header size"
	case wasi.ErrorCodeHttpRequestTrailerSectionSize:
		v := code.HttpRequestTrailerSectionSize()
		if v.IsSome() {
			return fmt.Sprintf("HTTP request trailer section size: %d", v.Some())
		}
		return "HTTP request trailer section size"
	case wasi.ErrorCodeHttpRequestTrailerSize:
		return "HTTP request trailer size " + fieldSizeString(code.HttpRequestTrailerSize())
	case wasi.ErrorCodeHttpResponseIncomplete:
		return "HTTP response incomplete"
	case wasi.ErrorCodeHttpResponseHeaderSectionSize:
		v := code.HttpResponseHeaderSectionSize()
		if v.IsSome() {
			return fmt.Sprintf("HTTP response header section size: %d", v.Some())
		}
		return "HTTP response header section size"
	case wasi.ErrorCodeHttpResponseHeaderSize:
		return "HTTP response header size " + fieldSizeString(code.HttpResponseHeaderSize())
	case wasi.ErrorCodeHttpResponseBodySize:
		v := code.HttpResponseBodySize()
		if v.IsSome() {
			return fmt.Sprintf("HTTP response body size: %d", v.Some())
		}
		return "HTTP response body size"
	case wasi.ErrorCodeHttpResponseTrailerSectionSize:
		v := code.HttpResponseTrailerSectionSize()
		if v.IsSome() {
			return fmt.Sprintf("HTTP response trailer section size: %d", v.Some())
		}
		return "HTTP response trailer section size"
	case wasi.ErrorCodeHttpResponseTrailerSize:
		return "HTTP response trailer size " + fieldSizeString(code.HttpResponseTrailerSize())
	case wasi.ErrorCodeHttpResponseTransferCoding:
		v := code.HttpResponseTransferCoding()
		if v.IsSome() {
			return fmt.Sprintf("HTTP response transfer coding: %s", v.Some())
		}
		return "HTTP response transfer coding"
	case wasi.ErrorCodeHttpResponseContentCoding:
		v := code.HttpResponseContentCoding()
		if v.IsSome() {
			return fmt.Sprintf("HTTP response content coding: %s", v.Some())
		}
		return "HTTP response content coding"
	case wasi.ErrorCodeHttpResponseTimeout:
		return "HTTP response timeout"
	case wasi.ErrorCodeHttpUpgradeFailed:
		return "HTTP upgrade failed"
	case wasi.ErrorCodeHttpProtocolError:
		return "HTTP protocol error"
	case wasi.ErrorCodeLoopDetected:
		return "loop detected"
	case wasi.ErrorCodeConfigurationError:
		return "configuration error"
	case wasi.ErrorCodeInternalError:
		v := code.InternalError()
		if v.IsSome() {
			return "internal error: " + v.Some()
		}
		return "internal error"
	default:
		return fmt.Sprintf("unknown error code: %d", code.Tag())
	}
}

func fieldSizeString(p wasi.FieldSizePayload) string {
	var parts []string
	if p.FieldName.IsSome() {
		parts = append(parts, fmt.Sprintf("field-name=%s", p.FieldName.Some()))
	}
	if p.FieldSize.IsSome() {
		parts = append(parts, fmt.Sprintf("field-size=%d", p.FieldSize.Some()))
	}
	if len(parts) == 0 {
		return ""
	}
	return "(" + strings.Join(parts, ", ") + ")"
}
