// Code generated by wit-bindgen-go. DO NOT EDIT.

package redis

import (
	redistypes "github.com/fermyon/spin-go-sdk/v2/internal/fermyon/spin/redis-types"
	"go.bytecodealliance.org/cm"
)

// This file contains wasmimport and wasmexport declarations for "fermyon:spin".

//go:wasmimport fermyon:spin/redis publish
//go:noescape
func wasmimport_Publish(address0 *uint8, address1 uint32, channel0 *uint8, channel1 uint32, payload0 *uint8, payload1 uint32, result *cm.Result[Error, struct{}, Error])

//go:wasmimport fermyon:spin/redis get
//go:noescape
func wasmimport_Get(address0 *uint8, address1 uint32, key0 *uint8, key1 uint32, result *cm.Result[redistypes.Payload, Payload, Error])

//go:wasmimport fermyon:spin/redis set
//go:noescape
func wasmimport_Set(address0 *uint8, address1 uint32, key0 *uint8, key1 uint32, value0 *uint8, value1 uint32, result *cm.Result[Error, struct{}, Error])

//go:wasmimport fermyon:spin/redis incr
//go:noescape
func wasmimport_Incr(address0 *uint8, address1 uint32, key0 *uint8, key1 uint32, result *cm.Result[int64, int64, Error])

//go:wasmimport fermyon:spin/redis del
//go:noescape
func wasmimport_Del(address0 *uint8, address1 uint32, keys0 *string, keys1 uint32, result *cm.Result[int64, int64, Error])

//go:wasmimport fermyon:spin/redis sadd
//go:noescape
func wasmimport_Sadd(address0 *uint8, address1 uint32, key0 *uint8, key1 uint32, values0 *string, values1 uint32, result *cm.Result[int64, int64, Error])

//go:wasmimport fermyon:spin/redis smembers
//go:noescape
func wasmimport_Smembers(address0 *uint8, address1 uint32, key0 *uint8, key1 uint32, result *cm.Result[cm.List[string], cm.List[string], Error])

//go:wasmimport fermyon:spin/redis srem
//go:noescape
func wasmimport_Srem(address0 *uint8, address1 uint32, key0 *uint8, key1 uint32, values0 *string, values1 uint32, result *cm.Result[int64, int64, Error])

//go:wasmimport fermyon:spin/redis execute
//go:noescape
func wasmimport_Execute(address0 *uint8, address1 uint32, command0 *uint8, command1 uint32, arguments0 *RedisParameter, arguments1 uint32, result *cm.Result[cm.List[RedisResult], cm.List[RedisResult], Error])
