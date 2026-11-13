package http

import (
	"fmt"
	"io"
	"net/http"

	wasi "github.com/spinframework/spin-go-sdk/v4/imports/wasi_http_0_3_2_types"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

// convert the http.Request to a wasi.Request
func newOutgoingHttpRequest(req *http.Request) (*wasi.Request, error) {
	headers, err := toWasiHeaders(req.Header)
	if err != nil {
		return nil, err
	}
	defer headers.Drop()

	var body wit.Option[*wit.StreamReader[uint8]]
	if req.Body == nil {
		body = wit.None[*wit.StreamReader[uint8]]()
	} else {
		tx, rx := wasi.MakeStreamU8()
		body = wit.Some(rx)
		go func() {
			defer tx.Drop()
			defer req.Body.Close()

			buffer := make([]uint8, 16*1024)
			for !tx.ReaderDropped() {
				count, err := req.Body.Read(buffer)
				if err != nil {
					if err != io.EOF {
						fmt.Printf("error reading request body: %v", err)
					}
					return
				}
				tx.WriteAll(buffer[:count])
			}
		}()
	}

	request, send := wasi.RequestNew(
		headers,
		body,
		trailersFuture(),                 // TODO: support trailers
		wit.None[*wasi.RequestOptions](), // TODO: support options
	)
	send.Drop()
	request.SetMethod(toWasiMethod(req.Method))
	request.SetPathWithQuery(wit.Some(req.URL.RequestURI()))

	authority := req.Host
	if authority == "" {
		authority = req.URL.Host
	}
	if authority != "" {
		request.SetAuthority(wit.Some(authority))
	}

	switch req.URL.Scheme {
	case "":
		// Leave unset so the receiver picks a default, e.g. when forwarding an incoming request.
	case "http":
		request.SetScheme(wit.Some(wasi.MakeSchemeHttp()))
	case "https":
		request.SetScheme(wit.Some(wasi.MakeSchemeHttps()))
	default:
		request.SetScheme(wit.Some(wasi.MakeSchemeOther(req.URL.Scheme)))
	}

	return request, nil
}

func toWasiMethod(s string) wasi.Method {
	switch s {
	case http.MethodConnect:
		return wasi.MakeMethodConnect()
	case http.MethodDelete:
		return wasi.MakeMethodDelete()
	case http.MethodGet:
		return wasi.MakeMethodGet()
	case http.MethodHead:
		return wasi.MakeMethodHead()
	case http.MethodOptions:
		return wasi.MakeMethodOptions()
	case http.MethodPatch:
		return wasi.MakeMethodPatch()
	case http.MethodPost:
		return wasi.MakeMethodPost()
	case http.MethodPut:
		return wasi.MakeMethodPut()
	case http.MethodTrace:
		return wasi.MakeMethodTrace()
	default:
		return wasi.MakeMethodOther(s)
	}
}
