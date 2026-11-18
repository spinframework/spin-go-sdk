package main

import (
	"fmt"
	"net/http"

	spinhttp "github.com/spinframework/spin-go-sdk/v4/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Hello from app! x-middleware=%q path=%s\n", r.Header.Get("x-middleware"), r.URL.RequestURI())
	})
}

func main() {}
