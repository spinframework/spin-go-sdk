// Package redis provides the handler function for the Redis trigger, as well
// as access to Redis within Spin components.

package redis

import (
	"errors"
	"fmt"

	"github.com/spinframework/spin-go-sdk/v3/internal/fermyon/spin/v2.0.0/redis"
	"go.bytecodealliance.org/cm"
)

// Client is a Redis client.
type Client struct {
	conn redis.Connection
}

// NewClient returns a Redis client.
func NewClient(address string) (Client, error) {
	conn, err, isErr := redis.ConnectionOpen(address).Result()
	if isErr {
		return Client{}, toError(err)
	}

	return Client{conn: conn}, nil
}

// Publish a Redis message to the specified channel.
func (c *Client) Publish(channel string, payload []byte) error {
	_, err, isErr := c.conn.Publish(channel, redis.Payload(cm.ToList(payload))).Result()
	if isErr {
		return toError(err)
	}

	return nil
}

// Get the value of a key.
func (c *Client) Get(key string) ([]byte, error) {
	payload, err, isErr := c.conn.Get(key).Result()
	if isErr {
		return nil, toError(err)
	}

	if payload.None() {
		return nil, nil
	}

	return payload.Some().Slice(), nil
}

// Set key to value.
//
// If key already holds a value, it is overwritten.
func (c *Client) Set(key string, payload []byte) error {
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
func (c *Client) Incr(key string) (int64, error) {
	incrementedNum, err, isErr := c.conn.Incr(key).Result()
	if isErr {
		return 0, toError(err)
	}

	return incrementedNum, nil
}

// Removes the specified keys.
//
// A key is ignored if it does not exist. Returns the number of keys deleted.
func (c *Client) Del(keys ...string) (uint32, error) {
	numKeysDeleted, err, isErr := c.conn.Del(cm.ToList(keys)).Result()
	if isErr {
		return 0, toError(err)
	}

	return numKeysDeleted, nil
}

// Add the specified `values` to the set named `key`, returning the number of newly-added values.
func (c *Client) Sadd(key string, values ...string) (uint32, error) {
	numValuesAdded, err, isErr := c.conn.Sadd(key, cm.ToList(values)).Result()
	if isErr {
		return 0, toError(err)
	}

	return numValuesAdded, nil
}

// Retrieve the contents of the set named `key`.
func (c *Client) Smembers(key string) ([]string, error) {
	setValues, err, isErr := c.conn.Smembers(key).Result()
	if isErr {
		return nil, toError(err)
	}

	return setValues.Slice(), nil
}

// Remove the specified `values` from the set named `key`, returning the number of newly-removed values.
func (c *Client) Srem(key string, values ...string) (uint32, error) {
	valuesRemoved, err, isErr := c.conn.Srem(key, cm.ToList(values)).Result()
	if isErr {
		return 0, toError(err)
	}

	return valuesRemoved, nil
}

// ResultKind represents a result type returned from executing a Redis command.
type ResultKind uint8

const (
	ResultKindNil ResultKind = iota
	ResultKindStatus
	ResultKindInt64
	ResultKindBinary
)

// String implements fmt.Stringer.
func (r ResultKind) String() string {
	switch r {
	case ResultKindNil:
		return "nil"
	case ResultKindStatus:
		return "status"
	case ResultKindInt64:
		return "int64"
	case ResultKindBinary:
		return "binary"
	default:
		return "unknown"
	}
}

// GoString implements fmt.GoStringer.
func (r ResultKind) GoString() string { return r.String() }

// Result represents a value returned from a Redis command.
type Result struct {
	Kind ResultKind
	Val  any
}

// Execute runs the specified Redis command with the specified arguments,
// returning zero or more results.  This is a general-purpose function which
// should work with any Redis command.
//
// Arguments must be string, []byte, int, int64, or int32.
func (c *Client) Execute(command string, arguments ...any) ([]*Result, error) {
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

	var results []*Result
	for _, r := range redisResults.Slice() {
		results = append(results, toResult(r))
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

func toResult(param redis.RedisResult) *Result {
	switch {
	case param.Status() != nil:
		return &Result{
			Kind: ResultKindStatus,
			Val:  param.Status(),
		}
	case param.Int64() != nil:
		return &Result{
			Kind: ResultKindInt64,
			Val:  param.Int64(),
		}
	case param.Binary() != nil:
		return &Result{
			Kind: ResultKindBinary,
			Val:  param.Binary().Slice(),
		}
	default:
		return &Result{
			Kind: ResultKindNil,
			Val:  param.Nil(),
		}
	}
}

func toError(e redis.Error) error {
	if e.String() == "other" {
		return fmt.Errorf(*e.Other())
	}

	return errors.New(e.String())
}
