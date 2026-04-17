package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	spinhttp "github.com/spinframework/spin-go-sdk/v3/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		// List files
		files, err := os.ReadDir(".")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading dir: %#v", err), http.StatusInternalServerError)
			return
		}
		if len(files) != 1 || files[0].Name() != "test.data" {
			http.Error(w, fmt.Sprintf("Files don't match: %#v", files), http.StatusInternalServerError)
			return
		}

		// Reading a missing file
		_, err = os.ReadFile("nope")
		if !errors.Is(err, fs.ErrNotExist) {
			http.Error(
				w,
				fmt.Sprintf("Should have path error, got: %#v\n", err),
				http.StatusInternalServerError,
			)
			return
		}

		content := "This is used for testing"
		dat, err := os.ReadFile("test.data")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error calling os.ReadFile: ", err), http.StatusInternalServerError)
			return
		}
		if string(dat) != content {
			http.Error(
				w,
				fmt.Sprintf("Error calling os.ReadFile(). Files contents don't match"),
				http.StatusInternalServerError,
			)
			return
		}
	})
}

func main() {}
