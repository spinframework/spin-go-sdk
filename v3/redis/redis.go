package redis

import (
	"errors"
	"fmt"

	"github.com/spinframework/spin-go-sdk/v3/internal/fermyon/spin/v2.0.0/redis"
	"go.bytecodealliance.org/cm"
)

type Connection struct {
	conn redis.Connection
}

// Open a connection to the Redis instance at `address`.
func Open(address string) (Connection, error) {
	conn, err, isErr := redis.ConnectionOpen(address).Result()
	if isErr {
		return Connection{}, toError(err)
	}

	return Connection{conn: conn}, nil
}

// Publish a Redis message to the specified channel.
func (c *Connection) Publish(channel string, payload []byte) error {
	_, err, isErr := c.conn.Publish(channel, redis.Payload(cm.ToList(payload))).Result()
	if isErr {
		return toError(err)
	}

	return nil
}

// Get the value of a key.
func (c *Connection) Get(key string) ([]byte, error) {
	payload, err, isErr := c.conn.Get(key).Result()
	if isErr {
		return nil, toError(err)
	}

	if payload.None() {
		return nil, nil
	} else {
		return payload.Some().Slice(), nil
	}
}

// Set key to value.
//
// If key already holds a value, it is overwritten.
func (c *Connection) Set(key string, payload []byte) error {
	_, err, isErr := c.conn.Set(key, redis.Payload(cm.ToList(payload))).Result()
	if isErr {
		return toError(err)
	}

	return nil
}

// Increments the number stored at key by one.
//
// If the key does not exist, it is set to 0 before performing the operation.
// An `error::type-error` is returned if the key contains a value of the wrong type
// or contains a string that can not be represented as integer.
func (c *Connection) Incr(key string) (int64, error) {
	incrementedNum, err, isErr := c.conn.Incr(key).Result()
	if isErr {
		return 0, toError(err)
	}

	return incrementedNum, nil
}

// Removes the specified keys.
//
// A key is ignored if it does not exist. Returns the number of keys deleted.
func (c *Connection) Del(keys []string) (uint32, error) {
	numKeysDeleted, err, isErr := c.conn.Del(cm.ToList(keys)).Result()
	if isErr {
		return 0, toError(err)
	}

	return numKeysDeleted, nil
}

// Add the specified `values` to the set named `key`, returning the number of newly-added values.
func (c *Connection) Sadd(key string, values []string) (uint32, error) {
	numValuesAdded, err, isErr := c.conn.Sadd(key, cm.ToList(values)).Result()
	if isErr {
		return 0, toError(err)
	}

	return numValuesAdded, nil
}

// Retrieve the contents of the set named `key`.
func (c *Connection) Smembers(key string) ([]string, error) {
	setValues, err, isErr := c.conn.Smembers(key).Result()
	if isErr {
		return nil, toError(err)
	}

	return setValues.Slice(), nil
}

// Remove the specified `values` from the set named `key`, returning the number of newly-removed values.
func (c *Connection) Srem(key string, values []string) (uint32, error) {
	valuesRemoved, err, isErr := c.conn.Srem(key, cm.ToList(values)).Result()
	if isErr {
		return 0, toError(err)
	}

	return valuesRemoved, nil
}

// Execute runs the specified Redis command with the specified arguments,
// returning zero or more results.  This is a general-purpose function which
// should work with any Redis command.
//
// Arguments must be string, []byte, int, int64, or int32.
func (c *Connection) Execute(command string, arguments ...any) ([]Result, error) {
	var params []redis.RedisParameter
	for _, a := range arguments {
		p, err := createParameter(a)
		if err != nil {
			return nil, err
		}
		params = append(params, p)
	}

	redisResults, err, isErr := c.conn.Execute(command, cm.ToList(params)).Result()
	if isErr {
		return nil, toError(err)
	}

	var results []Result
	for _, r := range redisResults.Slice() {
		result, err := toResult(r)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func createParameter(x any) (redis.RedisParameter, error) {
	switch v := x.(type) {
	case int:
		return redis.RedisParameterInt64(int64(v)), nil
	case int64:
		return redis.RedisParameterInt64(v), nil
	case int32:
		return redis.RedisParameterInt64(int64(v)), nil
	case []byte:
		return redis.RedisParameterBinary(redis.Payload(cm.ToList(v))), nil
	case string:
		return redis.RedisParameterBinary(redis.Payload(cm.ToList([]byte(v)))), nil
	default:
		return redis.RedisParameter{}, fmt.Errorf("invalid type %T; must be string, []byte, int, int64, or int32", v)
	}
}

type Result struct {
	tag       uint8
	int64Val  int64
	bytesVal  []byte
	statusVal string
}

const (
	TagStatus = 0
	TagInt64  = 1
	TagBinary = 2
)

func (r Result) IsInt64() bool  { return r.tag == TagInt64 }
func (r Result) IsBinary() bool { return r.tag == TagBinary }
func (r Result) IsStatus() bool { return r.tag == TagStatus }

func (r Result) AsInt64() (int64, bool) {
	if r.tag == TagInt64 {
		return r.int64Val, true
	}
	return 0, false
}

func (r Result) AsBytes() ([]byte, bool) {
	if r.tag == TagBinary {
		return r.bytesVal, true
	}
	return nil, false
}

func (r Result) AsStatus() (string, bool) {
	if r.tag == TagStatus {
		return r.statusVal, true
	}

	return "", false
}

func toResult(param redis.RedisResult) (Result, error) {
	if param.Status() != nil {
		return Result{
			tag:       TagStatus,
			statusVal: *param.Status(),
		}, nil
	} else if param.Int64() != nil {
		return Result{
			tag:      TagInt64,
			int64Val: *param.Int64(),
		}, nil
	} else if param.Binary() != nil {
		data := param.Binary().Slice()
		return Result{
			tag: TagBinary, bytesVal: data,
		}, nil
	} else { // param.Nil() == true
		return Result{}, fmt.Errorf("internal type error for RedisResult")
	}
}

func toError(e redis.Error) error {
	if e.String() == "other" {
		return fmt.Errorf(*e.Other())
	}

	return errors.New(e.String())
}
