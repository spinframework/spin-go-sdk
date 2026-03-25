package main

import (
	"fmt"

	"github.com/spinframework/spin-go-sdk/v3/inbound_redis"
)

func init() {
	// inbound_redis.Handle() must be called in the init() function.
	inbound_redis.Handle(func(payload []byte) error {
		fmt.Println("Payload::::")
		fmt.Println(string(payload))
		return nil
	})
}

// main function must be included for the compiler but is not executed.
func main() {}
