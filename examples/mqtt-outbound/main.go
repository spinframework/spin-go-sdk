package main

import (
	"net/http"
	"os"
	"strconv"

	spinhttp "github.com/spinframework/spin-go-sdk/v3/http"
	"github.com/spinframework/spin-go-sdk/v3/mqtt"
)

func main() {}

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		addr := os.Getenv("MQTT_ADDRESS")
		usr := os.Getenv("MQTT_USERNAME")
		pass := os.Getenv("MQTT_PASSWORD")
		keepAliveStr := os.Getenv("MQTT_KEEP_ALIVE_INTERVAL")
		topic := os.Getenv("MQTT_TOPIC")

		keepAlive, err := strconv.Atoi(keepAliveStr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("MQTT_KEEP_ALIVE_INTERVAL is not valid: must be an integer"))
		}

		conn, err := mqtt.OpenConnection(addr, usr, pass, uint64(keepAlive))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		message := []byte("Eureka!")

		if err := conn.Publish(topic, message, mqtt.QosAtMostOnce); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		w.WriteHeader(200)
		w.Write([]byte("Message successfully published!\n"))
	})
}
