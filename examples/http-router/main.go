package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	spinhttp "github.com/spinframework/spin-go-sdk/v3/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		router := httprouter.New()
		router.GET("/hello/:name", Hello)
		router.GET("/this/will/*catchAll", CatchAll)

		router.ServeHTTP(w, r)
	})
}

func Hello(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func CatchAll(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "catch all: %s!\n", ps.ByName("catchAll"))
}

func main() {}
