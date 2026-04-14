package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	spinhttp "github.com/spinframework/spin-go-sdk/v3/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		r1, err := spinhttp.Get("https://random-data-api.fermyon.app/animals/json")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if body, err := readToString(r1.Body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			fmt.Fprintln(w, body)
		}
		fmt.Fprintln(w, r1.Header.Get("content-type"))

		r2, err := spinhttp.Post("https://postman-echo.com/post", "text/plain", r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if body, err := readToString(r2.Body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			fmt.Fprintln(w, body)
		}

		req, err := http.NewRequest("PUT", "https://postman-echo.com/put", bytes.NewBufferString("General Kenobi!"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Add("foo", "bar")
		r3, err := spinhttp.Send(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if body, err := readToString(r3.Body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			fmt.Fprintln(w, body)
		}

		// `spin.toml` is not configured to allow outbound HTTP requests to this host,
		// so this request will fail.
		if _, err := spinhttp.Get("https://fermyon.com"); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot send HTTP request: %v", err)
		}
	})
}

func readToString(input io.Reader) (string, error) {
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, input); err != nil {
		return "", err
	} else {
		return buf.String(), nil
	}
}

func main() {}
