// Package redis provides access to Redis within Spin components, as well as a
// handler for inbound Redis messages.
package redis

import (
	"errors"
	"fmt"

	incominghandler "github.com/spinframework/spin-go-sdk/v3/exports/spin_up_redis_trigger_4_0_0/export_spin_redis_3_0_0_inbound_redis"
	_ "github.com/spinframework/spin-go-sdk/v3/exports/spin_up_redis_trigger_4_0_0/wit_exports"
	redis "github.com/spinframework/spin-go-sdk/v3/imports/spin_redis_3_0_0_redis"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

// Handle sets the handler function for the inbound Redis trigger.
// It must be called from an init() function.
func Handle(handle func(message []byte) error) {
	incominghandler.Exports.Handle = func(message []byte) wit.Result[wit.Unit, redis.Error] {
		if err := handle(message); err != nil {
			return wit.Err[wit.Unit](redis.MakeErrorOther(err.Error()))
		}
		return wit.Ok[wit.Unit, redis.Error](wit.Unit{})
	}
}

// Client is a Redis client.
type Client struct {
	conn redis.Connection
}

// NewClient returns a Redis client.
func NewClient(address string) (Client, error) {
	result := redis.ConnectionOpen(address)
	if result.IsErr() {
		return Client{}, toError(result.Err())
	}

	return Client{conn: *result.Ok()}, nil
}

// Publish publishes a Redis message to the specified channel.
func (c *Client) Publish(channel string, payload []byte) error {
	result := c.conn.Publish(channel, redis.Payload(payload))
	if result.IsErr() {
		return toError(result.Err())
	}

	return nil
}

// Get returns the value of a key.
func (c *Client) Get(key string) ([]byte, error) {
	result := c.conn.Get(key)
	if result.IsErr() {
		return nil, toError(result.Err())
	}

	if result.Ok().IsNone() {
		return nil, nil
	}

	return result.Ok().Some(), nil
}

// Set sets the value of a key.
//
// If key already holds a value, it is overwritten.
func (c *Client) Set(key string, payload []byte) error {
	result := c.conn.Set(key, redis.Payload(payload))
	if result.IsErr() {
		return toError(result.Err())
	}

	return nil
}

// Incr increments the number stored at key by one.
//
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type
// or contains a string that can not be represented as an integer.
func (c *Client) Incr(key string) (int64, error) {
	result := c.conn.Incr(key)
	if result.IsErr() {
		return 0, toError(result.Err())
	}

	return result.Ok(), nil
}

// Del removes the specified keys.
//
// A key is ignored if it does not exist. It returns the number of keys deleted.
func (c *Client) Del(keys ...string) (uint32, error) {
	result := c.conn.Del(keys)
	if result.IsErr() {
		return 0, toError(result.Err())
	}

	return result.Ok(), nil
}

// Sadd adds the specified values to the set named key, returning the number of newly-added values.
func (c *Client) Sadd(key string, values ...string) (uint32, error) {
	result := c.conn.Sadd(key, values)
	if result.IsErr() {
		return 0, toError(result.Err())
	}

	return result.Ok(), nil
}

// Smembers retrieves the contents of the set named key.
func (c *Client) Smembers(key string) ([]string, error) {
	result := c.conn.Smembers(key)
	if result.IsErr() {
		return nil, toError(result.Err())
	}

	return result.Ok(), nil
}

// Srem removes the specified values from the set named key, returning the number of newly-removed values.
func (c *Client) Srem(key string, values ...string) (uint32, error) {
	result := c.conn.Srem(key, values)
	if result.IsErr() {
		return 0, toError(result.Err())
	}

	return result.Ok(), nil
}

// ResultKind represents a result type returned from executing a Redis command.
type ResultKind uint8

const (
	// ResultKindNil indicates a nil result.
	ResultKindNil ResultKind = iota
	// ResultKindStatus indicates a status string result.
	ResultKindStatus
	// ResultKindInt64 indicates an int64 result.
	ResultKindInt64
	// ResultKindBinary indicates a binary (byte slice) result.
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

	result := c.conn.Execute(command, params)
	if result.IsErr() {
		return nil, toError(result.Err())
	}

	var results []*Result
	for _, r := range result.Ok() {
		results = append(results, toResult(r))
	}

	return results, nil
}

func createParameter(x any) (redis.RedisParameter, error) {
	switch v := x.(type) {
	case int:
		return redis.MakeRedisParameterInt64(int64(v)), nil
	case int64:
		return redis.MakeRedisParameterInt64(v), nil
	case int32:
		return redis.MakeRedisParameterInt64(int64(v)), nil
	case []byte:
		return redis.MakeRedisParameterBinary(redis.Payload(v)), nil
	case string:
		return redis.MakeRedisParameterBinary(redis.Payload(v)), nil
	default:
		return redis.RedisParameter{}, fmt.Errorf("invalid type %T; must be string, []byte, int, int64, or int32", v)
	}
}

func toResult(param redis.RedisResult) *Result {
	switch param.Tag() {
	case redis.RedisResultStatus:
		return &Result{
			Kind: ResultKindStatus,
			Val:  param.Status(),
		}
	case redis.RedisResultInt64:
		return &Result{
			Kind: ResultKindInt64,
			Val:  param.Int64(),
		}
	case redis.RedisResultBinary:
		return &Result{
			Kind: ResultKindBinary,
			Val:  param.Binary(),
		}
	default:
		return &Result{
			Kind: ResultKindNil,
			Val:  nil,
		}
	}
}

func toError(e redis.Error) error {
	switch e.Tag() {
	case redis.ErrorInvalidAddress:
		return errors.New("redis: invalid address")
	case redis.ErrorTooManyConnections:
		return errors.New("redis: too many connections")
	case redis.ErrorTypeError:
		return errors.New("redis: type error")
	case redis.ErrorOther:
		return fmt.Errorf("redis: %s", e.Other())
	default:
		return fmt.Errorf("redis: unknown error %v", e)
	}
}
