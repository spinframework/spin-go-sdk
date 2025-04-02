package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	spinhttp "github.com/spinframework/spin-go-sdk/v2/http"
	"github.com/spinframework/spin-go-sdk/v2/kv"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		store, err := kv.OpenDefault()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = store.Set("foo", []byte("bar"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		value, err := store.Get("foo")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if string(value) != "bar" {
			http.Error(w, fmt.Sprintf("expected: %q, got: %q", "bar", value), http.StatusInternalServerError)
			return
		}

		keys, err := store.GetKeys()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(keys)
	})
}

func main() {}
