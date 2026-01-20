package mqtt

import (
	"fmt"

	mqtt "github.com/spinframework/spin-go-sdk/v3/internal/fermyon_spin_2_0_0_mqtt"
)

type Connection struct {
	conn mqtt.Connection
}

// OpenConnection initializes an MQTT connection
func OpenConnection(address, username, password string, keepAliveIntervalInSecs uint64) (Connection, error) {
	result := mqtt.ConnectionOpen(address, username, password, keepAliveIntervalInSecs)
	if result.IsErr() {
		return Connection{}, toError(result.Err())
	}

	return Connection{conn: *result.Ok()}, nil
}

// Publish publishes an MQTT message
func (c *Connection) Publish(topic string, payload []byte, qos QoS) error {
	result := c.conn.Publish(topic, mqtt.Payload(payload), mqtt.Qos(qos))
	if result.IsErr() {
		return toError(result.Err())
	}

	return nil
}

// QoS for publishing Mqtt messages
type QoS = mqtt.Qos

const (
	QosAtMostOnce  = mqtt.QosAtMostOnce
	QosAtLeastOnce = mqtt.QosAtLeastOnce
	QosExactlyOnce = mqtt.QosExactlyOnce
)

func toError(err mqtt.Error) error {
	switch err.Tag() {
	case mqtt.ErrorConnectionFailed:
		return fmt.Errorf("connection-failed: %s", err.ConnectionFailed())
	default:
		return fmt.Errorf("%s", err.Other())
	}
}
