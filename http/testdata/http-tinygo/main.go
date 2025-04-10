package main

import (
	"fmt"
	"net/http"

	spinhttp "github.com/spinframework/spin-go-sdk/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Spin-Raw-Component-Route") != "/hello/..." {
			http.Error(w, "Spin-Raw-Component-Route is not /hello/...", http.StatusInternalServerError)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "Method should be GET", http.StatusInternalServerError)
			return
		}

		w.Header().Set("spin-path-info", r.Header.Get("spin-path-info"))
		w.Header().Set("foo", "bar")

		fmt.Fprintln(w, "Hello world!")
	})
}

func main() {}
