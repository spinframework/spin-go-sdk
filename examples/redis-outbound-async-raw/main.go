package main

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"

	handler "github.com/spinframework/spin-go-sdk/v3/exports/wasi_http_service_0_3_0_rc_2026_03_15/export_wasi_http_0_3_0_rc_2026_03_15_handler"
	_ "github.com/spinframework/spin-go-sdk/v3/exports/wasi_http_service_0_3_0_rc_2026_03_15/wit_exports"
	redis "github.com/spinframework/spin-go-sdk/v3/imports/spin_redis_3_0_0_redis"
	. "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
	. "go.bytecodealliance.org/pkg/wit/types"
)

func Handle(request *Request) Result[*Response, ErrorCode] {
	// addr is the environment variable set in `spin.toml` that points to the
	// address of the Redis server.
	addr := os.Getenv("REDIS_ADDRESS")

	// channel is the environment variable set in `spin.toml` that specifies
	// the Redis channel that the component will publish to.
	channel := os.Getenv("REDIS_CHANNEL")

	// payload is the data publish to the redis channel.
	payload := []byte(`Hello redis from Go!`)

	connectionResult := redis.ConnectionOpen(addr)
	if connectionResult.IsErr() {
		return errorResponse(connectionResult.Err())
	}
	connection := connectionResult.Ok()

	if result := connection.Publish(channel, payload); result.IsErr() {
		return errorResponse(connectionResult.Err())
	}

	if result := connection.Set("mykey", []byte("myvalue")); result.IsErr() {
		return errorResponse(connectionResult.Err())
	}

	buffer := new(bytes.Buffer)

	if result := connection.Get("mykey"); result.IsErr() {
		return errorResponse(connectionResult.Err())
	} else if result.Ok().IsNone() {
		return errorMessageResponse("missing value for key")
	} else {
		buffer.Write([]byte("mykey value was: "))
		buffer.Write(result.Ok().Some())
		buffer.Write([]byte("\n"))
	}

	if result := connection.Incr("sping-go-incr"); result.IsErr() {
		return errorResponse(connectionResult.Err())
	} else {
		buffer.Write([]byte("spin-go-incr value: "))
		buffer.Write([]byte(strconv.FormatInt(result.Ok(), 10)))
		buffer.Write([]byte("\n"))
	}

	if result := connection.Del([]string{"sping-go-incr", "mykey", "non-existing-key"}); result.IsErr() {
		return errorResponse(connectionResult.Err())
	} else {
		buffer.Write([]byte("deleted keys num: "))
		buffer.Write([]byte(strconv.FormatInt(int64(result.Ok()), 10)))
		buffer.Write([]byte("\n"))
	}

	if result := connection.Sadd("myset", []string{"foo", "bar"}); result.IsErr() {
		return errorResponse(connectionResult.Err())
	}

	{
		expected := []string{"bar", "foo"}
		if result := connection.Smembers("myset"); result.IsErr() {
			return errorResponse(connectionResult.Err())
		} else {
			members := result.Ok()
			sort.Strings(members)
			if !reflect.DeepEqual(members, expected) {
				return errorMessageResponse(
					fmt.Sprintf(
						"unexpected SMEMBERS result: expected %v, got %v",
						expected,
						payload,
					),
				)
			}
		}
	}

	if result := connection.Srem("myset", []string{"bar"}); result.IsErr() {
		return errorResponse(connectionResult.Err())
	}

	{
		expected := []string{"foo"}
		if result := connection.Smembers("myset"); result.IsErr() {
			return errorResponse(connectionResult.Err())
		} else {
			members := result.Ok()
			sort.Strings(members)
			if !reflect.DeepEqual(members, expected) {
				return errorMessageResponse(
					fmt.Sprintf(
						"unexpected SMEMBERS result: expected %v, got %v",
						expected,
						payload,
					),
				)
			}
		}
	}

	if result := connection.Execute("set", []redis.RedisParameter{
		redis.MakeRedisParameterBinary([]byte("message")),
		redis.MakeRedisParameterBinary([]byte("hello")),
	}); result.IsErr() {
		return errorResponse(connectionResult.Err())
	}

	if result := connection.Execute("append", []redis.RedisParameter{
		redis.MakeRedisParameterBinary([]byte("message")),
		redis.MakeRedisParameterBinary([]byte(" world")),
	}); result.IsErr() {
		return errorResponse(connectionResult.Err())
	}

	if result := connection.Execute("get", []redis.RedisParameter{
		redis.MakeRedisParameterBinary([]byte("message")),
	}); result.IsErr() {
		return errorResponse(result.Err())
	} else if !reflect.DeepEqual(
		result.Ok(),
		[]redis.RedisResult{redis.MakeRedisResultBinary([]byte("hello world"))},
	) {
		return errorMessageResponse("unexpected GET result")
	}

	return successResponse(buffer.Bytes())
}

func errorResponse(error redis.Error) Result[*Response, ErrorCode] {
	var message string
	switch error.Tag() {
	case redis.ErrorInvalidAddress:
		message = "redis: invalid address"

	case redis.ErrorTooManyConnections:
		message = "redis: too many connections"

	case redis.ErrorTypeError:
		message = "redis: type error"

	case redis.ErrorOther:
		message = fmt.Sprintf("redis: %v", error.Other())

	default:
		panic("unreachable")
	}
	return errorMessageResponse(message)
}

func errorMessageResponse(message string) Result[*Response, ErrorCode] {
	tx, rx := MakeStreamU8()

	go func() {
		defer tx.Drop()
		tx.WriteAll([]byte(message))
	}()

	response, send := ResponseNew(
		FieldsFromList([]Tuple2[string, []uint8]{
			Tuple2[string, []uint8]{"content-type", []uint8("text/plain")},
		}).Ok(),
		Some(rx),
		trailersFuture(),
	)
	send.Drop()
	response.SetStatusCode(500).Ok()

	return Ok[*Response, ErrorCode](response)
}

func successResponse(body []byte) Result[*Response, ErrorCode] {
	tx, rx := MakeStreamU8()

	go func() {
		defer tx.Drop()
		tx.WriteAll(body)
	}()

	response, send := ResponseNew(
		FieldsFromList([]Tuple2[string, []uint8]{
			Tuple2[string, []uint8]{"content-type", []uint8("text/plain")},
		}).Ok(),
		Some(rx),
		trailersFuture(),
	)
	send.Drop()

	return Ok[*Response, ErrorCode](response)
}

func trailersFuture() *FutureReader[Result[Option[*Fields], ErrorCode]] {
	tx, rx := MakeFutureResultOptionFieldsErrorCode()
	go tx.Write(Ok[Option[*Fields], ErrorCode](None[*Fields]()))
	return rx
}

func init() {
	handler.Exports.Handle = Handle
}

func main() {}
