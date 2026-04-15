// Package mqtt provides an MQTT client for publishing messages from Spin components.
package mqtt

import (
	"fmt"

	mqtt "github.com/spinframework/spin-go-sdk/v3/imports/fermyon_spin_2_0_0_mqtt"
)

// Connection represents an MQTT connection.
type Connection struct {
	conn mqtt.Connection
}

// OpenConnection opens a new MQTT connection to the specified address.
func OpenConnection(address, username, password string, keepAliveIntervalInSecs uint64) (Connection, error) {
	result := mqtt.ConnectionOpen(address, username, password, keepAliveIntervalInSecs)
	if result.IsErr() {
		return Connection{}, toError(result.Err())
	}

	return Connection{conn: *result.Ok()}, nil
}

// Publish sends an MQTT message to the specified topic.
func (c *Connection) Publish(topic string, payload []byte, qos QoS) error {
	result := c.conn.Publish(topic, mqtt.Payload(payload), mqtt.Qos(qos))
	if result.IsErr() {
		return toError(result.Err())
	}

	return nil
}

// QoS represents the quality of service level for MQTT messages.
type QoS = mqtt.Qos

const (
	// QosAtMostOnce delivers the message at most once (fire and forget).
	QosAtMostOnce = mqtt.QosAtMostOnce
	// QosAtLeastOnce delivers the message at least once (acknowledged delivery).
	QosAtLeastOnce = mqtt.QosAtLeastOnce
	// QosExactlyOnce delivers the message exactly once (assured delivery).
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
