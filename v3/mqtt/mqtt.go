package mqtt

import (
	"errors"
	"fmt"

	"github.com/spinframework/spin-go-sdk/v3/internal/fermyon/spin/v2.0.0/mqtt"
	"go.bytecodealliance.org/cm"
)

type Connection struct {
	conn mqtt.Connection
}

// OpenConnection initializes an MQTT connection
func OpenConnection(address, username, password string, keepAliveIntervalInSecs uint64) (Connection, error) {
	conn, err, isErr := mqtt.ConnectionOpen(address, username, password, keepAliveIntervalInSecs).Result()
	if isErr {
		return Connection{}, toError(&err)
	}

	return Connection{conn: conn}, nil
}

// Publish publishes an MQTT message
func (c *Connection) Publish(topic string, payload []byte, qos QoS) error {
	_, err, isErr := c.conn.Publish(topic, mqtt.Payload(cm.ToList(payload)), mqtt.Qos(qos)).Result()
	if isErr {
		return toError(&err)
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

func toError(err *mqtt.Error) error {
	if err == nil {
		return nil
	}

	if err.String() == "connection-failed" {
		return fmt.Errorf("connection-failed: %s", *err.ConnectionFailed())
	}

	if err.String() == "other" {
		return fmt.Errorf(*err.Other())
	}

	return errors.New(err.String())
}
