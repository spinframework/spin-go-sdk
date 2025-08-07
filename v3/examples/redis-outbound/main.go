package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	spinhttp "github.com/spinframework/spin-go-sdk/v3/http"
	"github.com/spinframework/spin-go-sdk/v3/redis"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		redisEndpoint := os.Getenv("REDIS_ENDPOINT")
		conn, err := redis.Open(redisEndpoint)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to open redis connection: " + err.Error()))
			return
		}

		// Set command
		if err := conn.Set("key1", []byte("value1")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to perform set operation: " + err.Error()))
			return
		}

		// Get command
		value, err := conn.Get("key1")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to perform get operation: " + err.Error()))
			return
		}

		fmt.Println("key1: " + string(value))

		// Incr command
		if _, err := conn.Incr("incr1"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to perform incr operation: " + err.Error()))
			return
		}

		// Incrementing the same key twice
		incrVal, err := conn.Incr("incr1")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to perform incr operation: " + err.Error()))
			return
		}

		fmt.Printf("incr1: %d\n", incrVal)

		// Del command
		numKeysDeleted, err := conn.Del([]string{"incr1", "key1"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to perform del operation: " + err.Error()))
			return
		}

		fmt.Printf("deleted %d keys\n", numKeysDeleted)

		// Sadd command
		languages := []string{"Go", "Rust", "JavaScript", "Python"}
		setName := "programming_languages"
		numAdded, err := conn.Sadd(setName, languages)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to perform sadd operation: " + err.Error()))
			return
		}

		fmt.Printf("added %d items to the %s set\n", numAdded, setName)

		// Smembers command
		setEntries, err := conn.Smembers(setName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to perform smembers operation: " + err.Error()))
			return
		}

		fmt.Println(setName + ": " + strings.Join(setEntries, ", "))

		// Srem command
		numRemoved, err := conn.Srem(setName, []string{"JavaScript"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to perform srem operation: " + err.Error()))
			return
		}

		fmt.Printf("deleted %d entries from set %s\n", numRemoved, setName)

		// Execute command
		if _, err := conn.Execute("SET", "execKey", "execValue"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to perform execute set operation: " + err.Error()))
			return
		}

		// Validating initial Execute command
		execResults, err := conn.Execute("GET", "execKey")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to perform execute get operation: " + err.Error()))
			return
		}

		for _, r := range execResults {
			if r.IsBinary() {
				data, ok := r.AsBytes()
				if !ok {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("failed to convert result type to bytes"))
					return
				}

				fmt.Println("execKey: " + string(data))

			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("incorrect result type received"))
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, Spin!"))
	})
}

func main() {}
