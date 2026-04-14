package http

import (
	"fmt"
	"io"
	"net/http"

	. "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
	. "go.bytecodealliance.org/pkg/wit/types"
)

// convert the http.Request to a Request
func newOutgoingHttpRequest(req *http.Request) (*Request, error) {
	headers, err := toWasiHeaders(req.Header)
	if err != nil {
		return nil, err
	}
	defer headers.Drop()

	var body Option[*StreamReader[uint8]]
	if req.Body == nil {
		body = None[*StreamReader[uint8]]()
	} else {
		tx, rx := MakeStreamU8()
		body = Some(rx)
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

	request, send := RequestNew(
		headers,
		body,
		trailersFuture(),        // TODO: support trailers
		None[*RequestOptions](), // TODO: support options
	)
	send.Drop()
	request.SetMethod(toWasiMethod(req.Method))
	request.SetAuthority(Some(req.Host))
	request.SetPathWithQuery(Some(req.URL.Path))

	switch req.URL.Scheme {
	case "http":
		request.SetScheme(Some(MakeSchemeHttp()))
	case "https":
		request.SetScheme(Some(MakeSchemeHttps()))
	default:
		request.SetScheme(Some(MakeSchemeOther(req.URL.Scheme)))
	}

	return request, nil
}

func toWasiMethod(s string) Method {
	switch s {
	case http.MethodConnect:
		return MakeMethodConnect()
	case http.MethodDelete:
		return MakeMethodDelete()
	case http.MethodGet:
		return MakeMethodGet()
	case http.MethodHead:
		return MakeMethodHead()
	case http.MethodOptions:
		return MakeMethodOptions()
	case http.MethodPatch:
		return MakeMethodPatch()
	case http.MethodPost:
		return MakeMethodPost()
	case http.MethodPut:
		return MakeMethodPut()
	case http.MethodTrace:
		return MakeMethodTrace()
	default:
		return MakeMethodOther(s)
	}
}
