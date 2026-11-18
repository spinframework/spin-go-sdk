package main

import (
	"io"
	"net/http"

	spinhttp "github.com/spinframework/spin-go-sdk/v4/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/blocked" {
			http.Error(w, "blocked by middleware", http.StatusForbidden)
			return
		}

		r.Header.Set("x-middleware", "hello from middleware")

		resp, err := spinhttp.Next(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.Header().Set("x-middleware-seen", "true")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})
}

func main() {}
