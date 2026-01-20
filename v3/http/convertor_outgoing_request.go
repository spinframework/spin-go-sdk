package http

import (
	"net/http"

	wit "github.com/bytecodealliance/wit-bindgen/wit_types"
	types "github.com/spinframework/spin-go-sdk/v3/internal/wasi_http_0_2_0_types"
)

// convert the IncomingRequest to http.Request
func NewOutgoingHttpRequest(req *http.Request) (types.OutgoingRequest, error) {
	headers := types.MakeFields()
	toWasiHeader(req.Header, *headers)

	or := types.MakeOutgoingRequest(headers)
	or.SetAuthority(wit.Some(req.Host))
	or.SetMethod(toWasiMethod(req.Method))
	or.SetPathWithQuery(wit.Some(req.URL.RawPath))

	switch req.URL.Scheme {
	case "http":
		or.SetScheme(wit.Some(types.MakeSchemeHttp()))
	case "https":
		or.SetScheme(wit.Some(types.MakeSchemeHttps()))
	default:
		or.SetScheme(wit.Some(types.MakeSchemeOther(req.URL.Scheme)))
	}

	return *or, nil
}

func toWasiHeader(src http.Header, dest types.Fields) {
	for k, v := range src {
		key := types.FieldKey(k)
		fieldVals := []types.FieldValue{}

		for _, val := range v {
			fieldVals = append(fieldVals, types.FieldValue(val))
		}

		if result := dest.Set(key, fieldVals); result.IsErr() {
			panic("failed to set WASI headers")
		}
	}
}

func toWasiMethod(s string) types.Method {
	switch s {
	case http.MethodConnect:
		return types.MakeMethodConnect()
	case http.MethodDelete:
		return types.MakeMethodDelete()
	case http.MethodGet:
		return types.MakeMethodGet()
	case http.MethodHead:
		return types.MakeMethodHead()
	case http.MethodOptions:
		return types.MakeMethodOptions()
	case http.MethodPatch:
		return types.MakeMethodPatch()
	case http.MethodPost:
		return types.MakeMethodPost()
	case http.MethodPut:
		return types.MakeMethodPut()
	case http.MethodTrace:
		return types.MakeMethodTrace()
	default:
		return types.MakeMethodOther(s)
	}
}
