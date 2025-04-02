package main

import (
	"fmt"
	"net/http"

	spinhttp "github.com/spinframework/spin-go-sdk/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("foo", "bar")

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "== RESPONSE ==")
		fmt.Fprintln(w, "Hello spinframework!")
		fmt.Fprintln(w, "Hello again spinframework!")
	})
}

func main() {}
